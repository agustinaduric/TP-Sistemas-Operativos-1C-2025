package fmemoria

import (
	"fmt"
	"math"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// getProcesoTP devuelve puntero al ProcesoTP para pid, o nil si no existe.
func getProcesoTP(pid int) *structs.ProcesoTP {
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			global.MemoriaLogger.Debug(fmt.Sprintf("getProcesoTP: encontrado PID=%d en índice %d", pid, i))
			return &global.ProcesosTP[i]
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("getProcesoTP: PID=%d no encontrado", pid))
	return nil
}

// ÍndiceProcesoTP devuelve el índice en global.ProcesosTP para pid, o -1 si no existe.
func ÍndiceProcesoTP(pid int) int {
	for i, proc := range global.ProcesosTP {
		if proc.PID == pid {
			global.MemoriaLogger.Debug(fmt.Sprintf("ÍndiceProcesoTP: PID=%d está en índice %d", pid, i))
			return i
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("ÍndiceProcesoTP: PID=%d no está en ProcesosTP", pid))
	return -1
}

// construirTabla crea recursivamente la jerarquía multinivel de tablas de páginas.
//
//	marcos   slice de índices de marcos a mapear
//	nivel    nivel actual (1..NumberOfLevels)
func construirTabla(marcos []int, nivel int) []structs.Tp {
	cfg := global.MemoriaConfig
	entradas := cfg.EntriesPerPage
	maxNiveles := cfg.NumberOfLevels

	// calcular cuántas entradas necesitamos en este nivel
	var nEntradas int
	if nivel < maxNiveles {
		factor := int(math.Pow(float64(entradas), float64(maxNiveles-nivel)))
		nEntradas = (len(marcos) + factor - 1) / factor // ceil
	} else {
		nEntradas = len(marcos)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("construirTabla: nivel=%d, entradas=%d, marcos totales=%d", nivel, nEntradas, len(marcos)))

	tps := make([]structs.Tp, nEntradas)
	idx := 0
	for i := 0; i < nEntradas; i++ {
		if nivel < maxNiveles {
			factor := int(math.Pow(float64(entradas), float64(maxNiveles-nivel)))
			count := factor
			if len(marcos)-idx < count {
				count = len(marcos) - idx
			}
			sub := marcos[idx : idx+count]
			idx += count
			global.MemoriaLogger.Debug(fmt.Sprintf("  nivel %d entrada %d: sub-marcos=%v", nivel, i, sub))
			tps[i].TablaSiguienteNivel = construirTabla(sub, nivel+1)
			tps[i].EsUltimoNivel = false
		} else {
			// último nivel: una sola entrada por marco
			tps[i].NumeroMarco = []int{marcos[idx]}
			global.MemoriaLogger.Debug(fmt.Sprintf("  nivel final entrada %d: marco=%d", i, marcos[idx]))
			idx++
			tps[i].EsUltimoNivel = true
		}
	}
	return tps
}

// InicializarProcesoTP crea la tabla multinivel para el proceso y lo agrega a global.ProcesosTP.
func InicializarProcesoTP(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProcesoTP: inicio PID=%d", pid))
	marcos := RecolectarMarcos(pid)
	tabla1 := construirTabla(marcos, 1)
	procTP := structs.ProcesoTP{PID: pid, TablaNivel1: tabla1}
	global.ProcesosTP = append(global.ProcesosTP, procTP)

	IncrementarAccesosTabla(pid)

	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProcesoTP: PID=%d tabla creada con %d entradas en nivel 1", pid, len(tabla1)))
}

// AsignarMarcosAProcesoTPPorPID vuelve a construir la tabla tras des-suspensión.
func AsignarMarcosAProcesoTPPorPID(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf("AsignarMarcosAProcesoTPPorPID: inicio PID=%d", pid))
	marcos := RecolectarMarcos(pid)
	if len(marcos) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: PID=%d sin marcos", pid))
		return
	}
	tabla1 := construirTabla(marcos, 1)
	idx := ÍndiceProcesoTP(pid)
	if idx < 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: PID=%d no en ProcesosTP", pid))
		return
	}
	//IncrementarAccesosTabla(pid)
	global.ProcesosTP[idx].TablaNivel1 = tabla1
	global.MemoriaLogger.Debug(fmt.Sprintf("AsignarMarcos: PID=%d tabla reconstruida", pid))

}
