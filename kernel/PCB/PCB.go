package PCB

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
	proceso.TiemposEstado = make(map[structs.Estado]time.Duration)
	proceso.MetricasEstado = make(map[structs.Estado]int)
	proceso.Desalojado = false

	if strings.EqualFold(global.ConfigCargadito.SchedulerAlgorithm, "SJF") || strings.EqualFold(global.ConfigCargadito.SchedulerAlgorithm, "SRT") {
		proceso.EstimadoRafaga = global.ConfigCargadito.InitialEstimate
		proceso.UltimaRafagaReal = 0
		global.KernelLogger.Debug(fmt.Sprintf("Se crea el PCB del proceso con PID: %d", proceso.PID))
		return proceso
	}

	proceso.EstimadoRafaga = -1
	proceso.UltimaRafagaReal = -1
	global.KernelLogger.Debug(fmt.Sprintf("Se crea el PCB del proceso con PID: %d", proceso.PID))

	//el resto de campos se inicializan en su "cero" por default (0, "", nil)

	return proceso
}

func Buscar_por_pid(PID int, Cola *structs.ColaProcesos) (structs.PCB, bool) {
	aux := make(structs.ColaProcesos, len(*Cola))
	copy(aux, *Cola)
	for i := 0; i < len(aux); i++ {
		if aux[i].PID == PID {
			global.KernelLogger.Debug(fmt.Sprintf("Se encontro el proceso por pid"))
			return aux[i], true
		}

	}
	global.KernelLogger.Debug(fmt.Sprintf("No Se encontro el proceso por pid"))
	return structs.PCB{}, false

}

func Actualizar_PC(PID int, PC int) {

	for i := 0; i < len(structs.ColaExecute); i++ {
		if (structs.ColaExecute)[i].PID == PID {
			//global.KernelLogger.Debug(fmt.Sprintf("Se Actualizo PC de PID: %d", PID))
			(structs.ColaExecute)[i].PC = PC
			return
		}

	}
	//global.KernelLogger.Debug(fmt.Sprintf("No Se PID para actualiza PC, este mensaje no deberia salir"))
	return

}
