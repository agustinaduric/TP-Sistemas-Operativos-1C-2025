package structs

import (
	"time"
)

type DispositivoIO struct {
	Nombre  string
	IP      string
	Puerto  int
	Ocupado bool
}

type SolicitudIO struct {
	PID      int
	Duracion time.Duration
	Servicio string
}
