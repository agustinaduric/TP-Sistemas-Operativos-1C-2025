package global

import (
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// var ConfirmacionProcesoCargado int
var ConfirmacionProcesoFinalizado int

var WgKernel sync.WaitGroup

var MutexPlanificadores sync.Mutex //asi se crea un mutex jej
var MutexLog sync.Mutex
var ProcesoCargado = make(chan int)
var ProcesoListo = make(chan int)
var ProcesoParaFinalizar = make(chan int)
var MutexCpuDisponible sync.Mutex
var ConfigCargadito config.KernelConfig

var SemFinalizacion = make(chan int)

//-----------------------------------------------MUTEX COLAS-------------------------------------------------------

var MutexNEW sync.Mutex
var MutexREADY sync.Mutex
var MutexEXEC sync.Mutex
var MutexBLOCKED sync.Mutex
var MutexSUSP_BLOCKED sync.Mutex
var MutexSUSP_READY sync.Mutex
var MutexEXIT sync.Mutex

//-----------------------------------------------FUNCIONES AUXILIARES-------------------------------------------------------

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

func Extraer_estado(Cola *structs.ColaProcesos, PID int) structs.PCB {
	var extraido structs.PCB
	for i := 0; i < len(*Cola); i++ {
		if (*Cola)[i].PID == PID {
			extraido = (*Cola)[i]
			*Cola = append((*Cola)[:i], (*Cola)[i+1:]...)

			return extraido
		}
	}
	return structs.PCB{}
}

func IniciarMetrica(estadoViejo string, estadoNuevo string, proceso *structs.PCB) {
	switch estadoNuevo {
	case "NEW":
		proceso.MetricasEstado[structs.NEW] = proceso.MetricasEstado[structs.NEW] + 1
		proceso.Estado = structs.NEW
		proceso.TiempoInicioEstado = time.Now()
		MutexNEW.Lock()
		Push_estado(&structs.ColaNew, *proceso)
		MutexNEW.Unlock()
		ProcesoCargado <- 0
	case "READY":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.READY
		proceso.MetricasEstado[structs.READY] = proceso.MetricasEstado[structs.READY] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexREADY.Lock()
		Push_estado(&structs.ColaReady, *proceso)
		MutexREADY.Unlock()
		ProcesoListo <- 0
		// PONER LOG OBLIGATORIOS DE CAMBIO DE ESTADO
	case "EXEC":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.EXEC
		proceso.MetricasEstado[structs.EXEC] = proceso.MetricasEstado[structs.EXEC] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexEXEC.Lock()
		Push_estado(&structs.ColaExecute, *proceso)
		MutexEXEC.Unlock()
		// PONER LOG OBLIGATORIOS DE CAMBIO DE ESTADO
	case "BLOCKED":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.BLOCKED
		proceso.MetricasEstado[structs.BLOCKED] = proceso.MetricasEstado[structs.BLOCKED] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexBLOCKED.Lock()
		Push_estado(&structs.ColaBlocked, *proceso)
		MutexBLOCKED.Unlock()
		// PONER LOG OBLIGATORIOS DE CAMBIO DE ESTADO
	case "SUSP_BLOCKED":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.SUSP_BLOCKED
		proceso.MetricasEstado[structs.SUSP_BLOCKED] = proceso.MetricasEstado[structs.SUSP_BLOCKED] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexSUSP_BLOCKED.Lock()
		Push_estado(&structs.ColaSuspBlocked, *proceso)
		MutexSUSP_BLOCKED.Unlock()
		// PONER LOG OBLIGATORIOS DE CAMBIO DE ESTADO
	case "SUSP_READY":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.SUSP_READY
		proceso.MetricasEstado[structs.SUSP_READY] = proceso.MetricasEstado[structs.SUSP_READY] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexSUSP_READY.Lock()
		Push_estado(&structs.ColaSuspReady, *proceso)
		MutexSUSP_READY.Unlock()
		// PONER LOG OBLIGATORIOS DE CAMBIO DE ESTADO
	case "EXIT":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.EXIT
		proceso.MetricasEstado[structs.EXIT] = proceso.MetricasEstado[structs.EXIT] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexEXIT.Lock()
		Push_estado(&structs.ColaExit, *proceso)
		MutexEXIT.Unlock()
		ProcesoParaFinalizar <- 0
		// PONER LOG OBLIGATORIOS DE CAMBIO DE ESTADO
	case "FINALIZADO":
		DetenerMetrica(estadoViejo, proceso)
		//LOGGEAR METRICAS Y LOG OBLIGATORIO DE FINALIZACION DE PROCESO
	}
}

func DetenerMetrica(estadoViejo string, proceso *structs.PCB) {
	switch estadoViejo {
	case "NEW":
		Extraer_estado(&structs.ColaNew, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.NEW] += duracion
	case "READY":
		Extraer_estado(&structs.ColaReady, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.READY] += duracion
	case "EXEC":
		Extraer_estado(&structs.ColaExecute, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.EXEC] += duracion
	case "BLOCKED":
		Extraer_estado(&structs.ColaBlocked, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.BLOCKED] += duracion
	case "SUSP_BLOCKED":
		Extraer_estado(&structs.ColaSuspBlocked, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.SUSP_BLOCKED] += duracion
	case "SUSP_READY":
		Extraer_estado(&structs.ColaSuspReady, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.SUSP_READY] += duracion
	case "EXIT":
		Extraer_estado(&structs.ColaExit, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.EXIT] += duracion
	}
}
