package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func InicializarProcesoTP(pid int) {
	// 1) Recoger todos los marcos asignados al pid
	var marcos []int
	for marco, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, marco)
		}
	}
	// 2) Índice para asignar cada marco
	idxMarco := 0

	// 3) Función recursiva
	var construir func(nivel int) []structs.Tp
	construir = func(nivel int) []structs.Tp {
		tps := make([]structs.Tp, global.MemoriaConfig.EntriesPerPage)
		for i := range tps {
			if nivel < global.MemoriaConfig.NumberOfLevels {
				tps[i].TablaSiguienteNivel = construir(nivel + 1)
			} else {
				// nivel hoja: asignar numeroMarco
				if idxMarco < len(marcos) {
					tps[i].NumeroMarco = marcos[idxMarco]
					idxMarco++
				} else {
					tps[i].NumeroMarco = -1
				}
				tps[i].EsUltimoNivel = true
			}
		}
		return tps
	}

	// 4) Crear ProcesoTP y añadir
	procesoTP := structs.ProcesoTP{
		PID:         pid,
		TablaNivel1: construir(1),
	}
	global.ProcesosTP = append(global.ProcesosTP, procesoTP)
}

func Marco(pid int, nivelesIndices []int) int {
	// 1) Buscar la tabla del proceso
	var procTP *structs.ProcesoTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			procTP = &global.ProcesosTP[i]
			break
		}
	}
	if procTP == nil {
		return -1
	}

	// 2) Validar longitud de índices
	if len(nivelesIndices) != global.MemoriaConfig.NumberOfLevels {
		return -1
	}

	// 3) Recorrer niveles
	nivelActual := procTP.TablaNivel1
	for _, idx := range nivelesIndices {
		if idx < 0 || idx >= len(nivelActual) {
			return -1
		}
		entrada := nivelActual[idx]
		if entrada.EsUltimoNivel {
			// retornar el numero de marco en hoja
			return entrada.NumeroMarco
		}
		nivelActual = entrada.TablaSiguienteNivel
	}

	return -1
}

func AsignarMarcosAProcesoTPPorPID(pidBuscado int) {
	// 1) Recolectar índices de marcos en MapMemoriaDeUsuario
	var marcos []int
	for idx, pid := range global.MapMemoriaDeUsuario {
		if pid == pidBuscado {
			marcos = append(marcos, idx)
		}
	}
	if len(marcos) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: no hay marcos asignados a PID %d", pidBuscado))
		return
	}

	// 2) Encontrar ProcesoTP correspondiente
	var procTP *structs.ProcesoTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pidBuscado {
			procTP = &global.ProcesosTP[i]
			break
		}
	}
	if procTP == nil {
		global.MemoriaLogger.Error(fmt.Sprintf("AsignarMarcos: PID %d no encontrado en ProcesosTP", pidBuscado))
		return
	}

	// 3) Iterador para asignar marcos en orden
	idxMarco := 0
	total := len(marcos)

	// 4) Función recursiva para recorrer niveles
	var asignar func(nivel []structs.Tp) bool
	asignar = func(nivel []structs.Tp) bool {
		for i := range nivel {
			entry := &nivel[i]
			if !entry.EsUltimoNivel {
				if asignar(entry.TablaSiguienteNivel) {
					return true
				}
			} else {
				if idxMarco < total {
					entry.NumeroMarco = marcos[idxMarco]
					idxMarco++
				} else {
					return true
				}
			}
		}
		return false
	}

	// 5) Iniciar asignación desde el nivel 1
	asignar(procTP.TablaNivel1)
}
