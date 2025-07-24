package fmemoria

import (
	"fmt"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

var MarcosMutex sync.Mutex

func MarcosDisponibles() int {
	MarcosMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[MarcosDisponibles] Tomé MarcosMutex"))
	MarcosMutex.Unlock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[MarcosDisponibles] Liberé MarcosMutex"))

	global.MemoriaLogger.Debug("MarcosDisponibles: contando marcos libres")

	var contadorMarcosDisponibles int

	totalMarcos := CantidadMarcos()

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosDisponibles =  TamanioMemoria = %d, TamanioPagina = %d, TotalMarcos = %d",
			global.MemoriaConfig.MemorySize,
			global.MemoriaConfig.PageSize,
			totalMarcos,
		),
	)

	for i := 0; i < totalMarcos; i++ {
		if global.MapMemoriaDeUsuario[i] == -1 {
			contadorMarcosDisponibles++
		}
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosDisponibles = %d", contadorMarcosDisponibles),
	)
	return contadorMarcosDisponibles
}

func MarcosNecesitados(tamanioProceso int) int {
	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosNecesitados: TamanioProceso=%d", tamanioProceso),
	)

	var necesarios int
	if rem := tamanioProceso % global.MemoriaConfig.PageSize; rem == 0 {
		necesarios = tamanioProceso / global.MemoriaConfig.PageSize
	} else {
		necesarios = (tamanioProceso / global.MemoriaConfig.PageSize) + 1
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosNecesitados: Necesarios=%d (para TamanioProceso=%d)", necesarios, tamanioProceso),
	)
	return necesarios
}

func CantidadMarcos() int {
	// Config immutable, no mutex needed
	return global.MemoriaConfig.MemorySize / global.MemoriaConfig.PageSize
}

func LiberarMarcos(pid int) {
	MarcosMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[LiberarMarcos] Tomé MarcosMutex"))
	defer func() {
		MarcosMutex.Unlock()
		global.MemoriaLogger.Debug(fmt.Sprintf("[LiberarMarcos] Liberé MarcosMutex"))
	}()
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
}

func OcuparMarcos(pid int) {
	MarcosMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[OcuparMarcos] Tomé MarcosMutex"))
	defer func() {
		MarcosMutex.Unlock()
		global.MemoriaLogger.Debug(fmt.Sprintf("[OcuparMarcos] Liberé MarcosMutex"))
	}()

	global.MemoriaLogger.Debug(fmt.Sprintf("OcuparMarcos: inicio PID=%d", pid))

	proc, _ := BuscarProceso(pid)
	necesarios := MarcosNecesitados(proc.Tamanio)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"OcuparMarcos: PID=%d requiere %d marcos (tamaño %d bytes)",
		pid, necesarios, proc.Tamanio,
	))

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
}

func RecolectarMarcos(pid int) []int {
	MarcosMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[RecolectarMarcos] Tomé MarcosMutex"))
	defer func() {
		MarcosMutex.Unlock()
		global.MemoriaLogger.Debug(fmt.Sprintf("[RecolectarMarcos] Liberé MarcosMutex"))
	}()

	var marcos []int
	for idx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, idx)
		}
	}
	return marcos
}

func Marco(pid int, indices []int) int {

	global.MemoriaLogger.Debug(fmt.Sprintf("Entre en Marco, indices=%v", indices))

	var marco int
	niveles := global.MemoriaConfig.NumberOfLevels
	if len(indices) != niveles {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Marco: longitud de índices inválida %d, se esperaban %d", len(indices), niveles),
		)
		return -1
	}

	procTP := getProcesoTP(pid)
	if procTP == nil {
		global.MemoriaLogger.Error(fmt.Sprintf("Marco: PID %d sin ProcesoTP", pid))
		return -1
	}

	tpActual := procTP.TablaNivel1
	for i := 0; i < niveles; i++ {
		if tpActual.EsUltimoNivel {
			marco = tpActual.MarcosEnEsteNivel[indices[niveles-1]]
			break
		}
		tpActual = tpActual.TablaSiguienteNivel[indices[i]]
	}

	IncrementarAccesosTabla(pid)
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Marco: PID=%d, indices=%v → marco=%d", pid, indices, marco),
	)
	return marco
}
