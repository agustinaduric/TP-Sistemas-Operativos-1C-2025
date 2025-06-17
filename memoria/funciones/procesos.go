package fmemoria

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

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

func InicializarProceso(pid int, tamanio int, instrucciones []structs.Instruccion) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProceso: inicio PID=%d, tamaño=%d", pid, tamanio))

	// 1. Crear el ProcesoMemoria con métricas a cero
	proc := structs.ProcesoMemoria{
		PID:           pid,
		Tamanio:       tamanio,
		EnSwap:        false,
		Metricas:      structs.MetricasMemoria{}, // todas las métricas en 0
		Path:          "",                        // TODO: ayuda
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

	// 4. Crear y agregar ProcesoTP
	procTP := AgregarProcesoTP(pid)
	// 5. Asignar esos marcos a las tablas de páginas
	AsignarMarcosAProcesoTP(procTP)

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"InicializarProceso: PID=%d listo para ejecutar",
		pid,
	))
	return nil
}

func FinalizarProceso(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf("FinalizarProceso: inicio PID=%d", pid))

	LiberarMarcos(pid)
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

	// 2. Eliminar de global.ProcesosTP
	antesLenTP := len(global.ProcesosTP)
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			global.ProcesosTP = append(global.ProcesosTP[:i], global.ProcesosTP[i+1:]...)
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"  PID=%d removido de ProcesosTP (antes %d, ahora %d)",
				pid, antesLenTP, len(global.ProcesosTP),
			))
			break
		}
	}
	if len(global.ProcesosTP) == antesLenTP {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"  FinalizarProceso: PID=%d no encontrado en ProcesosTP", pid,
		))
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("FinalizarProceso: fin PID=%d", pid))
}

// AgregarProcesoTP construye la paginación jerárquica para un PID dado:
// - NumberOfLevels tablas,
// - EntriesPerPage entradas por tabla,
// todas inicializadas a -1.
// Devuelve un puntero al ProcesoTP recién creado.
func AgregarProcesoTP(pid int) *structs.ProcesoTP {
	cfg := global.MemoriaConfig
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"AgregarProcesoTP: iniciando para PID=%d con %d tablas de %d páginas cada una",
		pid, cfg.NumberOfLevels, cfg.EntriesPerPage,
	))

	// 1. Construir el slice de Tp
	tps := make([]structs.Tp, cfg.NumberOfLevels)
	for i := range tps {
		pags := make([]int, cfg.EntriesPerPage)
		for j := range pags {
			pags[j] = -1
		}
		tps[i] = structs.Tp{Paginas: pags}
		global.MemoriaLogger.Debug(fmt.Sprintf(
			"  Tabla %d inicializada: páginas = %v",
			i, pags,
		))
	}

	// 2. Crear el ProcesoTP y agregarlo a la lista global
	procTP := &structs.ProcesoTP{
		PID: pid,
		TPS: tps,
	}
	global.ProcesosTP = append(global.ProcesosTP, *procTP)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"AgregarProcesoTP: PID=%d agregado con éxito (total procesosTP=%d)",
		pid, len(global.ProcesosTP),
	))
	return procTP
}

// AsignarMarcosAProcesoTP recorre MapMemoriaDeUsuario y asigna a
// procTP.TPS las posiciones de marco donde el PID coincide. Reparte
// los marcos en orden, llenando primero la Tabla[0], luego Tabla[1], etc.
func AsignarMarcosAProcesoTP(procTP *structs.ProcesoTP) {
	pid := procTP.PID
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"AsignarMarcosAProcesoTP: inicio PID=%d", pid,
	))

	tablaIdx := 0
	paginaIdx := 0
	numTablas := len(procTP.TPS)
	if numTablas == 0 {
		global.MemoriaLogger.Error("AsignarMarcosAProcesoTP: no hay tablas en el ProcesoTP")
		return
	}
	numPaginas := len(procTP.TPS[0].Paginas)

	global.MarcosMutex.Lock()

	for marco, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante != pid {
			continue
		}

		procTP.TPS[tablaIdx].Paginas[paginaIdx] = marco
		global.MemoriaLogger.Debug(fmt.Sprintf(
			"  PID=%d asigna marco %d a Tabla[%d].Paginas[%d]",
			pid, marco, tablaIdx, paginaIdx,
		))

		paginaIdx++
		if paginaIdx >= numPaginas {
			paginaIdx = 0
			tablaIdx++
			if tablaIdx >= numTablas {
				break
			}
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"AsignarMarcosAProcesoTP: fin PID=%d", pid,
	))
	global.MarcosMutex.Unlock()
}
