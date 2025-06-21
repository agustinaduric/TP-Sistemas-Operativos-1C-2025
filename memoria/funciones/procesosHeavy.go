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
	entriesPerPage := global.MemoriaConfig.EntriesPerPage
	numberOfLevels := global.MemoriaConfig.NumberOfLevels

	tps := make([]structs.Tp, entriesPerPage)
	for i := range tps {
		if nivel < numberOfLevels {
			subtabla, newIdx := construirTablaNivel(marcos, idx, nivel+1)
			tps[i].TablaSiguienteNivel = subtabla
			tps[i].EsUltimoNivel = false
			tps[i].NumeroMarco = -1
			idx = newIdx
		} else {
			if idx < len(marcos) {
				tps[i].NumeroMarco = marcos[idx]
				idx++
			} else {
				tps[i].NumeroMarco = -1
			}
			tps[i].EsUltimoNivel = true
			tps[i].TablaSiguienteNivel = nil
		}
	}
	return tps, idx
}

// InicializarProcesoTP recoge marcos y construye TablaNivel1 llamando a construirTablaNivel.
func InicializarProcesoTP(pid int) {
	marcos := collectMarcos(pid)
	idxMarco := 0
	tablaNivel1, _ := construirTablaNivel(marcos, idxMarco, 1)
	procesoTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: tablaNivel1,
	}
	global.ProcesosTP = append(global.ProcesosTP, procesoTP)
}

// Marco recorre la jerarquía multinivel según nivelesIndices y devuelve el NúmeroMarco en hoja o -1.
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
			return entrada.NumeroMarco
		}
		nivelActual = entrada.TablaSiguienteNivel
		if nivelActual == nil {
			return -1
		}
	}
	return -1
}

// AsignarMarcosAProcesoTPPorPID simplemente reutiliza InicializarProcesoTP para reasignar según MapMemoriaDeUsuario.
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
	idxMarco := 0
	tablaNivel1, _ := construirTablaNivel(marcos, idxMarco, 1)
	procTP.TablaNivel1 = tablaNivel1
}
