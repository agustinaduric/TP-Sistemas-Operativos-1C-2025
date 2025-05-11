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
	PID int
	NombreIO string
	Duracion time.Duration
}

type RegistroIO struct{
	Nombre       string
	IP           string
	Puerto       int
}

type RespuestaIO struct{
	NombreIO      string
	PID           int
	Desconexion bool
}