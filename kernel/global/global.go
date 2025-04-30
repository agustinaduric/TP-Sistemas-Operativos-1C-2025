package global

import (
	"sync"
)

var MutexPlanificadores sync.Mutex //asi se crea un mutex jej
var MutexLog sync.Mutex
var MutexCola sync.Mutex
var ProcesoCargado = make(chan int)
var ProcesoListo = make(chan int)
var MutexCpuDisponible sync.Mutex
