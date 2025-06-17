package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func IncrementarBajadasSwap(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarBajadasSwap: inicio PID=%d", pid,
	))

	// 1. Verificar existencia
	_, encontrado := BuscarProceso(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"IncrementarBajadasSwap: PID=%d no encontrado", pid,
		))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	// 2. Recorrer global.Procesos para actualizar la métrica
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].Metricas.BajadasSwap++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"IncrementarBajadasSwap: PID=%d BajadasSwap=%d",
				pid, global.Procesos[i].Metricas.BajadasSwap,
			))
			break
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarBajadasSwap: fin PID=%d", pid,
	))
	return nil
}

// IncrementarInstSolicitadas incrementa en uno la métrica InstSolicitadas
// del proceso con el PID dado.
func IncrementarInstSolicitadas(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarInstSolicitadas: inicio PID=%d", pid,
	))
	_, encontrado := BuscarProceso(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"IncrementarInstSolicitadas: PID=%d no encontrado", pid,
		))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].Metricas.InstSolicitadas++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"IncrementarInstSolicitadas: PID=%d InstSolicitadas=%d",
				pid, global.Procesos[i].Metricas.InstSolicitadas,
			))
			break
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarInstSolicitadas: fin PID=%d", pid,
	))
	return nil
}

// IncrementarSubidasMem incrementa en uno la métrica SubidasMem
// del proceso con el PID dado.
func IncrementarSubidasMem(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarSubidasMem: inicio PID=%d", pid,
	))
	_, encontrado := BuscarProceso(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"IncrementarSubidasMem: PID=%d no encontrado", pid,
		))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].Metricas.SubidasMem++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"IncrementarSubidasMem: PID=%d SubidasMem=%d",
				pid, global.Procesos[i].Metricas.SubidasMem,
			))
			break
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarSubidasMem: fin PID=%d", pid,
	))
	return nil
}

// IncrementarLecturasMem incrementa en uno la métrica LecturasMem
// del proceso con el PID dado.
func IncrementarLecturasMem(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarLecturasMem: inicio PID=%d", pid,
	))
	_, encontrado := BuscarProceso(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"IncrementarLecturasMem: PID=%d no encontrado", pid,
		))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].Metricas.LecturasMem++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"IncrementarLecturasMem: PID=%d LecturasMem=%d",
				pid, global.Procesos[i].Metricas.LecturasMem,
			))
			break
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarLecturasMem: fin PID=%d", pid,
	))
	return nil
}

// IncrementarEscriturasMem incrementa en uno la métrica EscriturasMem
// del proceso con el PID dado.
func IncrementarEscriturasMem(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarEscriturasMem: inicio PID=%d", pid,
	))
	_, encontrado := BuscarProceso(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"IncrementarEscriturasMem: PID=%d no encontrado", pid,
		))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].Metricas.EscriturasMem++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"IncrementarEscriturasMem: PID=%d EscriturasMem=%d",
				pid, global.Procesos[i].Metricas.EscriturasMem,
			))
			break
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarEscriturasMem: fin PID=%d", pid,
	))
	return nil
}

// IncrementarAccesosTabla incrementa en uno la métrica AccesosTabla
// del proceso con el PID dado.
func IncrementarAccesosTabla(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarAccesosTabla: inicio PID=%d", pid,
	))
	_, encontrado := BuscarProceso(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"IncrementarAccesosTabla: PID=%d no encontrado", pid,
		))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}
	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].Metricas.AccesosTabla++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"IncrementarAccesosTabla: PID=%d AccesosTabla=%d",
				pid, global.Procesos[i].Metricas.AccesosTabla,
			))
			break
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarAccesosTabla: fin PID=%d", pid,
	))
	return nil
}
