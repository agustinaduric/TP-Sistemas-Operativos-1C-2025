// File: fmemoria/metrics.go

package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// BuscarProcesoI devuelve el índice en global.Procesos donde está el PID.
// Retorna (índice, true) si existe, o (-1, false) si no.
// Usa logging para depuración.
func BuscarProcesoI(pid int) (int, bool) {
	global.MemoriaLogger.Debug(fmt.Sprintf("BuscarProcesoI: inicio PID=%d", pid))

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.MemoriaLogger.Debug(fmt.Sprintf("BuscarProcesoI: Proceso encontrado: PID=%d, Índice=%d", pid, i))
			return i, true
		}
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("BuscarProcesoI: Proceso con PID=%d no encontrado", pid))
	return -1, false
}

// IncrementarBajadasSwap incrementa en uno la métrica BajadasSwap del proceso con PID dado.
func IncrementarBajadasSwap(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarBajadasSwap: inicio PID=%d", pid))

	idx, encontrado := BuscarProcesoI(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarBajadasSwap: PID=%d no encontrado", pid))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	// Proteger actualización de métricas
	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	// Verificar consistencia: que global.Procesos[idx] siga siendo el PID
	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarBajadasSwap: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}
	global.Procesos[idx].Metricas.BajadasSwap++
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarBajadasSwap: PID=%d BajadasSwap=%d",
		pid, global.Procesos[idx].Metricas.BajadasSwap,
	))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarBajadasSwap: fin PID=%d", pid))
	return nil
}

// IncrementarInstSolicitadas incrementa en uno la métrica InstSolicitadas del proceso con PID dado.
func IncrementarInstSolicitadas(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarInstSolicitadas: inicio PID=%d", pid))

	idx, encontrado := BuscarProcesoI(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarInstSolicitadas: PID=%d no encontrado", pid))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarInstSolicitadas: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}
	global.Procesos[idx].Metricas.InstSolicitadas++
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarInstSolicitadas: PID=%d InstSolicitadas=%d",
		pid, global.Procesos[idx].Metricas.InstSolicitadas,
	))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarInstSolicitadas: fin PID=%d", pid))
	return nil
}

// IncrementarSubidasMem incrementa en uno la métrica SubidasMem del proceso con PID dado.
func IncrementarSubidasMem(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarSubidasMem: inicio PID=%d", pid))

	idx, encontrado := BuscarProcesoI(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarSubidasMem: PID=%d no encontrado", pid))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarSubidasMem: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}
	global.Procesos[idx].Metricas.SubidasMem++
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarSubidasMem: PID=%d SubidasMem=%d",
		pid, global.Procesos[idx].Metricas.SubidasMem,
	))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarSubidasMem: fin PID=%d", pid))
	return nil
}

// IncrementarLecturasMem incrementa en uno la métrica LecturasMem del proceso con PID dado.
func IncrementarLecturasMem(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarLecturasMem: inicio PID=%d", pid))

	idx, encontrado := BuscarProcesoI(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarLecturasMem: PID=%d no encontrado", pid))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarLecturasMem: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}
	global.Procesos[idx].Metricas.LecturasMem++
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarLecturasMem: PID=%d LecturasMem=%d",
		pid, global.Procesos[idx].Metricas.LecturasMem,
	))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarLecturasMem: fin PID=%d", pid))
	return nil
}

// IncrementarEscriturasMem incrementa en uno la métrica EscriturasMem del proceso con PID dado.
func IncrementarEscriturasMem(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarEscriturasMem: inicio PID=%d", pid))

	idx, encontrado := BuscarProcesoI(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarEscriturasMem: PID=%d no encontrado", pid))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarEscriturasMem: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}
	global.Procesos[idx].Metricas.EscriturasMem++
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarEscriturasMem: PID=%d EscriturasMem=%d",
		pid, global.Procesos[idx].Metricas.EscriturasMem,
	))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarEscriturasMem: fin PID=%d", pid))
	return nil
}

// IncrementarAccesosTabla incrementa en uno la métrica AccesosTabla del proceso con PID dado.
func IncrementarAccesosTabla(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarAccesosTabla: inicio PID=%d", pid))

	idx, encontrado := BuscarProcesoI(pid)
	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarAccesosTabla: PID=%d no encontrado", pid))
		return fmt.Errorf("PID=%d no existe en memoria principal", pid)
	}

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarAccesosTabla: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}
	global.Procesos[idx].Metricas.AccesosTabla++
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"IncrementarAccesosTabla: PID=%d AccesosTabla=%d",
		pid, global.Procesos[idx].Metricas.AccesosTabla,
	))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarAccesosTabla: fin PID=%d", pid))
	return nil
}
