package fmemoria

import (
	"fmt"

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

func construirTablas(marcos *[]int, nivelActual int) structs.Tp {
	cfg := global.MemoriaConfig
	entradas := cfg.EntriesPerPage
	totalNiveles := cfg.NumberOfLevels
	var tpRetornable structs.Tp

	// Verificamos si hay suficientes marcos para construir la tabla
	if len(*marcos) == 0 {
		global.MemoriaLogger.Debug(fmt.Sprintf("construirTablas: No hay suficientes marcos para construir la tabla en nivel %d", nivelActual))
		return tpRetornable // Retornamos una tabla vacía
	}

	if nivelActual == totalNiveles {
		return construirTablaUltimoNivel(marcos, tpRetornable)
	} else {
		tpRetornable.TablaSiguienteNivel = make([]structs.Tp, entradas)
		for j := 0; j < entradas; j++ {
			tpRetornable.TablaSiguienteNivel[j] = construirTablaIntermedia(marcos, nivelActual+1, tpRetornable.TablaSiguienteNivel[j])
		}
	}
	return tpRetornable
}

func construirTablaIntermedia(marcos *[]int, nivelActual int, tablaActual structs.Tp) structs.Tp {
	cfg := global.MemoriaConfig
	entradas := cfg.EntriesPerPage
	totalNiveles := cfg.NumberOfLevels

	// Verificamos si hay suficientes marcos para construir la tabla
	if len(*marcos) == 0 {
		global.MemoriaLogger.Debug(fmt.Sprintf("construirTablaIntermedia: No hay suficientes marcos para construir la tabla en nivel %d", nivelActual))
		return tablaActual // Retornamos la tabla actual sin cambios
	}

	if nivelActual == totalNiveles {
		return construirTablaUltimoNivel(marcos, tablaActual)
	} else {
		tablaActual.TablaSiguienteNivel = make([]structs.Tp, entradas)
		for j := 0; j < entradas; j++ {
			tablaActual.TablaSiguienteNivel[j] = construirTablaIntermedia(marcos, nivelActual+1, tablaActual.TablaSiguienteNivel[j])
		}
	}
	return tablaActual
}

func construirTablaUltimoNivel(marcos *[]int, tablaActual structs.Tp) structs.Tp {
	cfg := global.MemoriaConfig
	entradas := cfg.EntriesPerPage

	// Verificamos si hay suficientes marcos para construir la tabla
	if len(*marcos) == 0 {
		global.MemoriaLogger.Debug("construirTablaUltimoNivel: No hay suficientes marcos para construir la tabla del último nivel")
		return tablaActual // Retornamos la tabla actual sin cambios
	}

	// Inicializamos el slice de marcos en este nivel
	tablaActual.MarcosEnEsteNivel = make([]int, entradas)

	// Llamamos a guardarMarcos para extraer los marcos y copiarlos al slice de marcos en este nivel
	guardarMarcos(marcos, entradas, &tablaActual.MarcosEnEsteNivel)

	tablaActual.EsUltimoNivel = true
	return tablaActual
}

func guardarMarcos(marcos *[]int, cantidad int, destino *[]int) {
	// Verificamos que la cantidad no exceda el tamaño del slice
	if cantidad > len(*marcos) {
		cantidad = len(*marcos)
	}

	// Copiamos los marcos seleccionados al slice destino
	copy(*destino, (*marcos)[:cantidad])

	// Modificamos el slice original para que contenga solo los marcos restantes
	*marcos = (*marcos)[cantidad:]
}

// InicializarProcesoTP construye la tabla multinivel completa para el PID
// y la añade a global.ProcesosTP.
func InicializarProcesoTP(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf("InicializarProcesoTP: PID=%d inicio", pid))
	marcos := RecolectarMarcos(pid)
	tabla := construirTablas(&marcos, 1)
	procTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: tabla,
	}
	if ÍndiceProcesoTP(pid) == -1 {
		global.ProcesosTP = append(global.ProcesosTP, procTP)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"InicializarProcesoTP: PID=%d tabla creada", pid,
	))
}
