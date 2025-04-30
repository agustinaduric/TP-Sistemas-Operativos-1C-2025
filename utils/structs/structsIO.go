package structs

import (
	"time"
)

type DispositivoIO struct {
	Nombre       string
	IP           string
	Puerto       int
	PIDActual    int // 0 si no hay procesos usando el io
	ColaEsperaIO []*PCB
}

type Solicitud struct {
	PID      int
	Duracion time.Duration
	NombreIO string
}
