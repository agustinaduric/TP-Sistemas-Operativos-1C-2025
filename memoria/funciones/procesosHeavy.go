package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// getProcesoTP devuelve puntero al structs.ProcesoTP para pid, o nil si no existe.
func getProcesoTP(pid int) *structs.ProcesoTP {
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			return &global.ProcesosTP[i]
		}
	}
	return nil
}

// collectMarcos devuelve slice de índices de marcos en MapMemoriaDeUsuario asignados a pid.
func collectMarcos(pid int) []int {
	var marcos []int
	for idx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, idx)
		}
	}
	return marcos
}

func construirTablaNivel(marcos []int, idx int, nivel int) ([]structs.Tp, int) {
	entries := global.MemoriaConfig.EntriesPerPage
	maxLevels := global.MemoriaConfig.NumberOfLevels

	tps := make([]structs.Tp, entries)
	for i := range tps {
		if nivel < maxLevels {
			sub, newIdx := construirTablaNivel(marcos, idx, nivel+1)
			tps[i].TablaSiguienteNivel = sub
			tps[i].EsUltimoNivel = false
			tps[i].NumeroMarco = nil
			idx = newIdx
		} else {
			// último nivel: asignamos todos los marcos restantes en las entradas de NumeroMarco
			tps[i].EsUltimoNivel = true
			tps[i].TablaSiguienteNivel = nil
			tps[i].NumeroMarco = make([]int, entries)
			for j := 0; j < entries; j++ {
				if idx < len(marcos) {
					tps[i].NumeroMarco[j] = marcos[idx]
					idx++
				} else {
					tps[i].NumeroMarco[j] = -1
				}
			}
		}
	}
	return tps, idx
}

func InicializarProcesoTP(pid int) {
	marcos := collectMarcos(pid)
	tabla1, _ := construirTablaNivel(marcos, 0, 1)
	procTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: tabla1,
	}
	global.ProcesosTP = append(global.ProcesosTP, procTP)
}

// Marco recorre la jerarquía multinivel según nivelesIndices y devuelve el primer NúmeroMarco en hoja o -1.
func Marco(pid int, indicesNiveles []int) int {
	cfg := global.MemoriaConfig
	niveles := cfg.NumberOfLevels

	// Validar longitud de índices
	if len(indicesNiveles) != niveles+1 {
		return -1
	}

	// 1) Obtener el ProcesoTP
	procesoTP := getProcesoTP(pid)
	if procesoTP == nil {
		return -1
	}

	// 2) Navegar por los primeros 'niveles' índices
	entradaActual := procesoTP.TablaNivel1
	var hoja structs.Tp

	for nivel := 1; nivel <= niveles; nivel++ {
		idx := indicesNiveles[nivel-1]
		if idx < 0 || idx >= len(entradaActual) {
			return -1
		}
		entrada := entradaActual[idx]

		if nivel < niveles {
			// Bajar al siguiente nivel
			entradaActual = entrada.TablaSiguienteNivel
			if entradaActual == nil {
				return -1
			}
		} else {
			// Último nivel → guardamos la entrada hoja
			hoja = entrada
		}
	}

	// 3) En la hoja, usamos el último índice para elegir dentro de NumeroMarco
	indicePuntero := indicesNiveles[niveles]
	if indicePuntero < 0 || indicePuntero >= len(hoja.NumeroMarco) {
		return -1
	}

	return hoja.NumeroMarco[indicePuntero]
}

// AsignarMarcosAProcesoTPPorPID reconstruye la tabla multinivel de un proceso tras una des-suspensión.
func AsignarMarcosAProcesoTPPorPID(pidBuscado int) {
	marcos := collectMarcos(pidBuscado)
	if len(marcos) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: no hay marcos asignados a PID %d", pidBuscado))
		return
	}
	procTP := getProcesoTP(pidBuscado)
	if procTP == nil {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: PID %d no encontrado en ProcesosTP", pidBuscado))
		return
	}
	tabla, _ := construirTablaNivel(marcos, 0, 1)
	procTP.TablaNivel1 = tabla
}
