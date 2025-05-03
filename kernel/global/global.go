package global

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

var MutexPlanificadores sync.Mutex //asi se crea un mutex jej
var MutexLog sync.Mutex
var ProcesoCargado = make(chan int)
var ProcesoListo = make(chan int)
var MutexCpuDisponible sync.Mutex
var ConfigCargadito config.KernelConfig

//-----------------------------------------------MUTEX COLAS-------------------------------------------------------

var MutexNEW sync.Mutex
var MutexREADY sync.Mutex
var MutexEXEC sync.Mutex
var MutexBLOCKED sync.Mutex
var MutexSUSP_BLOCKED sync.Mutex
var MutexSUSP_READY sync.Mutex
var MutexEXIT sync.Mutex
