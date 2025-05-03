package PCB

import (
	"strings"
	"sync"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var MutexPID sync.Mutex
var contPID int = 0

func Crear(PATH string, Tamanio int) structs.PCB {
	var proceso structs.PCB

	MutexPID.Lock()
	proceso.PID = contPID
	contPID++
	MutexPID.Unlock()

	proceso.PATH = PATH
	proceso.Tamanio = Tamanio
	proceso.PC = 0

	if strings.EqualFold(global.ConfigCargadito.SchedulerAlgorithm, "SJF") || strings.EqualFold(global.ConfigCargadito.SchedulerAlgorithm, "SRT") {
		// HACER CUANDO HAGA SJF Y SRT
		return proceso
	}

	proceso.EstimadoRafaga = -1
	proceso.UltimaRafagaReal = -1

	//el resto de campos se inicializan en su "cero" por default (0, "", nil)

	return proceso
}

func Pop_estado(Cola *structs.ColaProcesos) structs.PCB {

	if len(*Cola) == 0 {
		return structs.PCB{} // o manejar el error como prefieras
	}

	pcb := (*Cola)[0]
	*Cola = (*Cola)[1:] // directamente recortás el slice

	return pcb
}

func Push_estado(Cola *structs.ColaProcesos, pcb structs.PCB) {
	*Cola = append(*Cola, pcb) // Usamos el puntero a Cola para modificar el slice
}
