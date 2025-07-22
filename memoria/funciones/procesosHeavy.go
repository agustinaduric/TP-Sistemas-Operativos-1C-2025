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

// construirSubtabla crea recursivamente una tabla de páginas de tipo Tp
// para el nivel `nivel` (1..NumberOfLevels). Devuelve un único Tp,
// no un slice, ya que cada nivel está encapsulado en un Tp.
func construirSubtabla(marcos []int, nivel int) structs.Tp {
	cfg := global.MemoriaConfig
	entradas := cfg.EntriesPerPage
	totalNiveles := cfg.NumberOfLevels

	tp := structs.Tp{}
	// Si aún no alcanzamos el último nivel, creamos hijos
	if nivel < totalNiveles {
		// calcular cuántas entradas necesita este nodo
		factor := int(math.Pow(float64(entradas), float64(totalNiveles-nivel)))
		nEntradas := (len(marcos) + factor - 1) / factor // ceil
		tp.TablaSiguienteNivel = make([]structs.Tp, nEntradas)

		idx := 0
		for i := 0; i < nEntradas; i++ {
			// cuántos marcos asignar a este subnodo
			count := factor
			if len(marcos)-idx < count {
				count = len(marcos) - idx
			}
			submarcos := marcos[idx : idx+count]
			idx += count
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"  Nivel %d. Entrada[%d]: marcos=%v", nivel, i, submarcos,
			))
			// recursión
			tp.TablaSiguienteNivel[i] = construirSubtabla(submarcos, nivel+1)
		}
		tp.EsUltimoNivel = false

	} else {
		// nivel hoja: aquí almacenamos todos los marcos en MarcosEnEsteNivel
		tp.MarcosEnEsteNivel = make([]int, len(marcos))
		for i, m := range marcos {
			tp.MarcosEnEsteNivel[i] = m
		}
		global.MemoriaLogger.Debug(fmt.Sprintf(
			"  Nivel hoja. NumeroMarco=%v", tp.MarcosEnEsteNivel,
		))
		tp.EsUltimoNivel = true
	}

	return tp
}

// InicializarProcesoTP construye la tabla multinivel completa para el PID
// y la añade a global.ProcesosTP.
func InicializarProcesoTP(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProcesoTP: PID=%d inicio", pid))
	marcos := RecolectarMarcos(pid)
	tabla := construirSubtabla(marcos, 0)
	procTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: tabla,
	}
	global.ProcesosTP = append(global.ProcesosTP, procTP)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"InicializarProcesoTP: PID=%d tabla creada", pid,
	))
}

// AsignarMarcosAProcesoTPPorPID reconstruye la tabla multinivel de un PID
// (idéntico a Inicializar, usado tras des-suspender)
func AsignarMarcosAProcesoTPPorPID(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf("AsignarMarcosAProcesoTPPorPID: PID=%d inicio", pid))
	marcos := RecolectarMarcos(pid)
	if len(marcos) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: PID=%d sin marcos", pid))
		return
	}
	nuevaTabla := construirSubtabla(marcos, 0)

	// localiza el índice en ProcesosTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			global.ProcesosTP[i].TablaNivel1 = nuevaTabla
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"AsignarMarcos: PID=%d tabla actualizada", pid,
			))
			return
		}
	}
	global.MemoriaLogger.Error(fmt.Sprintf(
		"AsignarMarcos: PID=%d no encontrado en ProcesosTP", pid,
	))
}

func DumpTablas(pid int) {
	procTP := getProcesoTP(pid)
	if procTP == nil {
		global.MemoriaLogger.Error(fmt.Sprintf("DumpTablas: PID=%d sin ProcesoTP", pid))
		return
	}
	var rec func(path string, tp structs.Tp)
	rec = func(path string, tp structs.Tp) {
		if tp.EsUltimoNivel {
			global.MemoriaLogger.Info(fmt.Sprintf(
				"Hoja %s → MarcosEnEsteNivel=%v", path, tp.MarcosEnEsteNivel,
			))
		} else {
			for idx, sub := range tp.TablaSiguienteNivel {
				rec(fmt.Sprintf("%s/%d", path, idx), sub)
			}
		}
	}
	global.MemoriaLogger.Info(fmt.Sprintf("DumpTablas: PID=%d", pid))
	rec("", procTP.TablaNivel1)
}
