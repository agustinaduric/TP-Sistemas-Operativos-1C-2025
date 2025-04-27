package structs

import (
	"time"
)

type Estado string

const (
	NEW          Estado = "NEW"
	READY        Estado = "READY"
	EXEC         Estado = "EXEC"
	BLOCKED      Estado = "BLOCKED"
	SUSP_BLOCKED Estado = "SUSP_BLOCKED"
	SUSP_READY   Estado = "SUSP_READY"
	EXIT         Estado = "EXIT"
)

type Instruccion struct {
	Operacion  string
	Argumentos []string
}
type PCB struct {
	PID              int
	Estado           Estado
	PC               int
	Tamanio          int
	EstimadoRafaga   float64
	UltimaRafagaReal float64
	IngresoEstado    time.Time
	MetricasEstado   map[Estado]int           // cant entradas estado
	TiemposEstado    map[Estado]time.Duration // tiempo en cada estado
	IOPendiente      string                   // nombre de IO en la que esta bloqueado
}
