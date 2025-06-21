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

// construirTablaNivel recorre recursivamente la jerarquía multinivel y asigna marcos desde el índice idx.
// Retorna la tabla creada y el nuevo valor de idx después de consumir marcos en este subárbol.
func construirTablaNivel(marcos []int, idx int, nivel int) ([]structs.Tp, int) {
	entries := global.MemoriaConfig.EntriesPerPage
	levels := global.MemoriaConfig.NumberOfLevels

	tps := make([]structs.Tp, entries)
	for i := range tps {
		// Si no es el último nivel, construyo el subárbol
		if nivel < levels {
			subtabla, newIdx := construirTablaNivel(marcos, idx, nivel+1)
			tps[i].TablaSiguienteNivel = subtabla
			tps[i].EsUltimoNivel = false
			// NumeroMarco queda nil
			idx = newIdx
		} else {
			// Nivel hoja: inicializo NumeroMarco con longitud entries
			tps[i].EsUltimoNivel = true
			tps[i].TablaSiguienteNivel = nil
			tps[i].NumeroMarco = make([]int, entries)
			// Sólo en la posición i consumo un marco si sobra
			if idx < len(marcos) {
				tps[i].NumeroMarco[0] = marcos[idx]
				idx++
			} else {
				tps[i].NumeroMarco[0] = -1
			}
			// para el resto dejo -1
			for j := 1; j < entries; j++ {
				tps[i].NumeroMarco[j] = -1
			}
		}
	}
	return tps, idx
}

// InicializarProcesoTP recoge marcos y construye TablaNivel1 llamando a construirTablaNivel.
func InicializarProcesoTP(pid int) {
	marcos := collectMarcos(pid)
	tabla, _ := construirTablaNivel(marcos, 0, 1)
	procTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: tabla,
	}
	global.ProcesosTP = append(global.ProcesosTP, procTP)
}

// Marco recorre la jerarquía multinivel según nivelesIndices y devuelve el primer NúmeroMarco en hoja o -1.
func Marco(pid int, nivelesIndices []int) int {
	procTP := getProcesoTP(pid)
	if procTP == nil {
		return -1
	}
	if len(nivelesIndices) != global.MemoriaConfig.NumberOfLevels {
		return -1
	}
	nivelActual := procTP.TablaNivel1
	for _, idx := range nivelesIndices {
		if idx < 0 || idx >= len(nivelActual) {
			return -1
		}
		entrada := nivelActual[idx]
		if entrada.EsUltimoNivel {
			// devolvemos el primer marco de esa hoja
			return entrada.NumeroMarco[0]
		}
		nivelActual = entrada.TablaSiguienteNivel
	}
	return -1
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
