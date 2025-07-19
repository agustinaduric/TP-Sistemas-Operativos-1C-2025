// File: fmemoria/metrics.go

package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// BuscarProcesoI devuelve el índice en global.Procesos donde está el PID.
// Ustedes diran, pero nardo, ya tenes una funcion llamada buscarProceso
// era necesario esta¿
// Bueno, si, esta al devolver el indice nos permite tratar con el proceso REAL
// y no con una copia del mismo, asi puedo aumentar las metricas de forma correcta
// y entonces para que existe buscarProceso¿
// escribiendo este comentario me di cuenta que yo tampoco se la respuesta
// pero se quedara
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

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()

	if idx < 0 || idx >= len(global.Procesos) || global.Procesos[idx].PID != pid {
		global.MemoriaLogger.Error(fmt.Sprintf("IncrementarBajadasSwap: inconsistencia tras BuscarProcesoI para PID=%d", pid))
		return fmt.Errorf("inconsistencia al actualizar métricas para PID=%d", pid)
	}

	// Valor antes de incrementar
	old := global.Procesos[idx].Metricas.BajadasSwap
	global.MemoriaLogger.Debug(fmt.Sprintf("Antes IncrementarBajadasSwap: PID=%d BajadasSwap=%d", pid, old))

	// Incremento
	global.Procesos[idx].Metricas.BajadasSwap++

	// Valor después de incrementar
	newVal := global.Procesos[idx].Metricas.BajadasSwap
	global.MemoriaLogger.Debug(fmt.Sprintf("Después IncrementarBajadasSwap: PID=%d BajadasSwap=%d", pid, newVal))

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

	// Valor antes de incrementar
	old := global.Procesos[idx].Metricas.InstSolicitadas
	global.MemoriaLogger.Debug(fmt.Sprintf("Antes IncrementarInstSolicitadas: PID=%d InstSolicitadas=%d", pid, old))

	// Incremento
	global.Procesos[idx].Metricas.InstSolicitadas++

	// Valor después de incrementar
	newVal := global.Procesos[idx].Metricas.InstSolicitadas
	global.MemoriaLogger.Debug(fmt.Sprintf("Después IncrementarInstSolicitadas: PID=%d InstSolicitadas=%d", pid, newVal))

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

	// Valor antes de incrementar
	old := global.Procesos[idx].Metricas.SubidasMem
	global.MemoriaLogger.Debug(fmt.Sprintf("Antes IncrementarSubidasMem: PID=%d SubidasMem=%d", pid, old))

	// Incremento
	global.Procesos[idx].Metricas.SubidasMem++

	// Valor después de incrementar
	newVal := global.Procesos[idx].Metricas.SubidasMem
	global.MemoriaLogger.Debug(fmt.Sprintf("Después IncrementarSubidasMem: PID=%d SubidasMem=%d", pid, newVal))

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

	// Valor antes de incrementar
	old := global.Procesos[idx].Metricas.LecturasMem
	global.MemoriaLogger.Debug(fmt.Sprintf("Antes IncrementarLecturasMem: PID=%d LecturasMem=%d", pid, old))

	// Incremento
	global.Procesos[idx].Metricas.LecturasMem++

	// Valor después de incrementar
	newVal := global.Procesos[idx].Metricas.LecturasMem
	global.MemoriaLogger.Debug(fmt.Sprintf("Después IncrementarLecturasMem: PID=%d LecturasMem=%d", pid, newVal))

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

	// Valor antes de incrementar
	old := global.Procesos[idx].Metricas.EscriturasMem
	global.MemoriaLogger.Debug(fmt.Sprintf("Antes IncrementarEscriturasMem: PID=%d EscriturasMem=%d", pid, old))

	// Incremento
	global.Procesos[idx].Metricas.EscriturasMem++

	// Valor después de incrementar
	newVal := global.Procesos[idx].Metricas.EscriturasMem
	global.MemoriaLogger.Debug(fmt.Sprintf("Después IncrementarEscriturasMem: PID=%d EscriturasMem=%d", pid, newVal))

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

	// Valor antes de incrementar
	old := global.Procesos[idx].Metricas.AccesosTabla
	global.MemoriaLogger.Debug(fmt.Sprintf("Antes IncrementarAccesosTabla: PID=%d AccesosTabla=%d", pid, old))

	// Incremento según niveles
	for i := 0; i < CalcularAccesosTablas(len(RecolectarMarcos(pid))); i++ {
		global.Procesos[idx].Metricas.AccesosTabla++
	}

	// Valor después de incrementar
	newVal := global.Procesos[idx].Metricas.AccesosTabla
	global.MemoriaLogger.Debug(fmt.Sprintf("Después IncrementarAccesosTabla: PID=%d AccesosTabla=%d", pid, newVal))

	global.MemoriaLogger.Debug(fmt.Sprintf("IncrementarAccesosTabla: fin PID=%d", pid))
	return nil
}

func CalcularAccesosTablas(cantidadMarcos int) int {
	return global.MemoriaConfig.NumberOfLevels
}

/*
func CalcularAccesosTablas(cantidadMarcos int) int {
	total := 0
	for nivel := 1; nivel <= global.MemoriaConfig.NumberOfLevels; nivel++ {
		// exponent = niveles - nivel + 1
		exp := global.MemoriaConfig.NumberOfLevels - nivel + 1
		divisor := 1
		for i := 0; i < exp; i++ {
			divisor *= global.MemoriaConfig.EntriesPerPage
		}
		// accesos en este nivel = ceil(marcos/divisor)
		cnt := (cantidadMarcos + divisor - 1) / divisor
		total += cnt
	}
	return total
}
	*/

