package fmemoria

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var MarcosMutex sync.Mutex

func BuscarProceso(pid int) (structs.ProcesoMemoria, bool) {
	// 1. Debug inicial con PID
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Buscando Proceso: PID=%d", pid),
	)

	for i, proc := range global.Procesos {
		if proc.PID == pid {
			global.MemoriaLogger.Debug(
				fmt.Sprintf("Proceso encontrado: PID=%d, Índice=%d", pid, i),
			)
			return proc, true
		}
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Proceso con PID=%d no encontrado", pid),
	)
	return structs.ProcesoMemoria{}, false
}

func BuscarInstruccion(pid int, pc int) (structs.Instruccion, error) {
	global.MemoriaLogger.Debug(fmt.Sprintf("Buscando Instrucción: PID=%d, PC=%d", pid, pc))

	proceso, existe := BuscarProceso(pid)
	if !existe {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error al buscar instrucción en memoria: proceso con PID %d no existe", pid),
		)
		return structs.Instruccion{}, fmt.Errorf("BuscarInstruccion: el pid %d no existe", pid)
	}

	if pc < 0 || pc >= len(proceso.Instrucciones) {

		global.MemoriaLogger.Error(
			fmt.Sprintf(
				"Error al buscar instrucción en memoria: PC %d fuera de rango [0,%d) para PID %d",
				pc, len(proceso.Instrucciones), pid,
			),
		)
		return structs.Instruccion{}, fmt.Errorf(
			"BuscarInstruccion: pc %d fuera de rango [0,%d) para pid %d",
			pc, len(proceso.Instrucciones), pid,
		)
	}
	IncrementarInstSolicitadas(pid)
	instr := proceso.Instrucciones[pc]
	global.MemoriaLogger.Info(
		fmt.Sprintf(
			"## PID: %d - Obtener instrucción: %d - Operación: %s - Argumentos: %v",
			pid, pc, instr.Operacion, instr.Argumentos,
		),
	)

	return instr, nil
}

func CargarInstrucciones(ruta string) ([]structs.Instruccion, error) {
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Cargando instrucciones desde ruta: %s", ruta),
	)

	archivo, err := os.Open(ruta)
	if err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error al abrir archivo de instrucciones '%s': %s", ruta, err.Error()),
		)
		return nil, err
	}
	defer archivo.Close()

	var instrucciones []structs.Instruccion
	escáner := bufio.NewScanner(archivo)
	for escáner.Scan() {
		línea := strings.TrimSpace(escáner.Text())
		if línea == "" {
			continue
		}
		partes := strings.Fields(línea)
		instrucciones = append(instrucciones, structs.Instruccion{
			Operacion:  partes[0],
			Argumentos: partes[1:],
		})
	}
	if err := escáner.Err(); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error leyendo líneas de '%s': %s", ruta, err.Error()),
		)
		return nil, err
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Total de instrucciones cargadas desde '%s': %d", ruta, len(instrucciones)),
	)

	return instrucciones, nil
}

func LiberarMarcos(pid int) {
	MarcosMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("LiberarMarcos: inicio PID=%d", pid))

	for marco, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			global.MapMemoriaDeUsuario[marco] = -1
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"LiberarMarcos: marco %d liberado (antes ocupado por PID=%d)",
				marco, pid,
			))
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("LiberarMarcos: fin PID=%d", pid))
	MarcosMutex.Unlock()
}

func OcuparMarcos(pid int) {
	MarcosMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("OcuparMarcos: inicio PID=%d", pid))

	// 1. Obtener el proceso para conocer su tamaño
	proc, _ := BuscarProceso(pid) // asumimos que ya existe porque hay espacio
	necesarios := MarcosNecesitados(proc.Tamanio)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"OcuparMarcos: PID=%d requiere %d marcos (tamaño %d bytes)",
		pid, necesarios, proc.Tamanio,
	))

	// 2. Asignar exactamente 'necesarios' marcos libres
	asignados := 0

	for idx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == -1 {
			global.MapMemoriaDeUsuario[idx] = pid
			asignados++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"OcuparMarcos: PID=%d ocupa marco %d (%d/%d)",
				pid, idx, asignados, necesarios,
			))
			if asignados == necesarios {
				break
			}
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"OcuparMarcos: fin PID=%d — se asignaron %d marcos",
		pid, asignados,
	))
	MarcosMutex.Unlock()
}

func InicializarProceso(pid int, tamanio int, instrucciones []structs.Instruccion, PATH string) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProceso: inicio PID=%d, tamaño=%d", pid, tamanio))

	// 1. Crear el ProcesoMemoria con métricas a cero
	proc := structs.ProcesoMemoria{
		PID:           pid,
		Tamanio:       tamanio,
		EnSwap:        false,
		Metricas:      structs.MetricasMemoria{}, // todas las métricas en 0
		Path:          PATH,                      // TODO: ayuda
		Instrucciones: instrucciones,
	}
	global.MemoriaLogger.Debug("  ProcesoMemoria construido con métricas a cero")

	// 2. Agregar a memoria principal
	global.Procesos = append(global.Procesos, proc)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"  PID=%d agregado a memoria principal (total procesos=%d)",
		pid, len(global.Procesos),
	))

	// 3. Reservar marcos en la memoria de usuario
	global.MemoriaLogger.Debug("  reservando marcos en MapMemoriaDeUsuario")
	OcuparMarcos(pid) // asume siempre hay espacio
	global.MemoriaLogger.Debug("  marcos reservados con éxito")

	InicializarProcesoTP(pid)

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"InicializarProceso: PID=%d listo para ejecutar",
		pid,
	))
	return nil
}

func FinalizarProceso(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("FinalizarProceso: inicio PID=%d", pid))

	// 1. Liberar marcos asignados
	LiberarMarcos(pid)

	// 2. Buscar métricas antes de eliminar
	var met structs.MetricasMemoria
	encontrado := false
	for _, proc := range global.Procesos {
		if proc.PID == pid {
			met = proc.Metricas
			encontrado = true
			break
		}
	}
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"FinalizarProceso: PID=%d no encontrado para métricas", pid,
		))
		return nil
	} else {
		// 3. Log obligatorio de métricas al destruir
		global.MemoriaLogger.Info(fmt.Sprintf(
			"## PID: %d - Proceso Destruido - Métricas - Acc.T.Pag: %d; Inst.Sol.: %d; SWAP: %d; Mem.Prin.: %d; Lec.Mem.: %d; Esc.Mem.: %d",
			pid,
			met.AccesosTabla,
			met.InstSolicitadas,
			met.BajadasSwap,
			met.SubidasMem,
			met.LecturasMem,
			met.EscriturasMem,
		))
	}

	// 4. Eliminar de memoria principal
	antesLen := len(global.Procesos)
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos = append(global.Procesos[:i], global.Procesos[i+1:]...)
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"  PID=%d removido de Procesos (antes %d, ahora %d)",
				pid, antesLen, len(global.Procesos),
			))
			break
		}
	}
	if len(global.Procesos) == antesLen {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"  FinalizarProceso: PID=%d no encontrado en Procesos", pid,
		))
	}

	// 5. Eliminar de paginación jerárquica
	for i, proc := range global.ProcesosTP {
		if proc.PID == pid {
			global.ProcesosTP = append(global.ProcesosTP[:i], global.ProcesosTP[i+1:]...)
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("FinalizarProceso: fin PID=%d", pid))
	return nil
}

func InicializarProcesoTP(pid int) {
	// 1) Recoger todos los marcos asignados al pid
	var marcos []int
	for marco, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, marco)
		}
	}
	// 2) Crear un índice para consumir marcos en el último nivel
	idxMarco := 0
	// 3) Función recursiva para construir niveles
	var construir func(nivel int) []structs.Tp
	construir = func(nivel int) []structs.Tp {
		tps := make([]structs.Tp, global.MemoriaConfig.EntriesPerPage)
		for i := range tps {
			if nivel < global.MemoriaConfig.NumberOfLevels {
				tps[i].TablaSiguienteNivel = construir(nivel + 1)
			}
			if nivel == global.MemoriaConfig.NumberOfLevels {
				// asignar puntero a byte de MemoriaUsuario para este marco
				if idxMarco < len(marcos) {
					offset := marcos[idxMarco] * global.MemoriaConfig.PageSize
					tps[i].ByteMemUsuario = []*byte{&global.MemoriaUsuario[offset]}
				}
				tps[i].EsUltimoNivel = true
				idxMarco++
			}
		}
		return tps
	}
	// 4) Construir la tabla de primer nivel
	procesoTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: construir(1),
	}
	// 5) Agregar a la lista global
	global.ProcesosTP = append(global.ProcesosTP, procesoTP)
}

func Marco(pid int, nivelesIndices []int) int {
	niveles := global.MemoriaConfig.NumberOfLevels

	if len(nivelesIndices) != niveles {
		return -1
	}

	// Busca el ProcesoTP
	var procTP *structs.ProcesoTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			procTP = &global.ProcesosTP[i]
			break
		}
	}
	if procTP == nil {
		return -1
	}

	// Recorre los niveles
	nivelActual := procTP.TablaNivel1
	for nivel, idx := range nivelesIndices {
		if idx < 0 || idx >= len(nivelActual) {
			return -1
		}
		entrada := nivelActual[idx]

		// si no es el último nivel, descendemos
		if nivel+1 < niveles {
			nivelActual = entrada.TablaSiguienteNivel
			continue
		}

		// nivel hoja: usamos idx para seleccionar el puntero al marco
		if len(entrada.ByteMemUsuario) == 0 {
			return -1
		}
		leafIdx := idx
		ptr := entrada.ByteMemUsuario[leafIdx]
		if ptr == nil {
			return -1
		}

		// --- Aritmética de punteros para hallar el marco ---
		base := uintptr(unsafe.Pointer(&global.MemoriaUsuario[0]))
		target := uintptr(unsafe.Pointer(ptr))
		if target < base {
			return -1
		}
		offset := target - base
		frameSize := uintptr(global.MemoriaConfig.PageSize)
		frame := int(offset / frameSize)

		// Validaciones finales
		if frame < 0 || frame >= len(global.MapMemoriaDeUsuario) {
			return -1
		}
		if owner := global.MapMemoriaDeUsuario[frame]; owner != pid {
			return -1
		}

		return frame
	}

	return -1
}
