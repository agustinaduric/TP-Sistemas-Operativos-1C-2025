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

	proc := structs.ProcesoMemoria{
		PID:           pid,
		Tamanio:       tamanio,
		EnSwap:        false,
		Metricas:      structs.MetricasMemoria{}, // todas las métricas en 0
		Path:          "",                        // CHAVALES NECESITO AYUDA CON ESTE
		Instrucciones: instrucciones,
	}
	global.MemoriaLogger.Debug("  ProcesoMemoria construido con métricas a cero")

	global.Procesos = append(global.Procesos, proc)
	global.MemoriaLogger.Debug(fmt.Sprintf("  PID=%d agregado a memoria principal (total procesos=%d)", pid, len(global.Procesos)))

	global.MemoriaLogger.Debug("  reservando marcos en MapMemoriaDeUsuario")
	OcuparMarcos(pid) // Memoria de usuario  ockeada
	//TODO
	//Hacer todo el laburito de la paginacion jerarquica
	//posdata: cada vez q vean un TODO seguramente es referido a la paginacion jerarquica

	global.MemoriaLogger.Debug("  marcos reservados con éxito")

	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProceso: PID=%d listo para ejecutar", pid))
	return nil
}
