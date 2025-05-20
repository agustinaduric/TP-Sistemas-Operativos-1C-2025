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
	PID                 int
	PATH                string
	Estado              Estado
	PC                  int
	Tamanio             int
	EstimadoRafaga      float64
	UltimaRafagaReal    float64
	IngresoEstado       time.Time
	MetricasEstado      map[Estado]int           // cant entradas estado
	TiemposEstado       map[Estado]time.Duration // tiempo en cada estado
	TiempoInicioEstado  time.Time                // para calcular cuanto tiempo estuvo en cada estado
	IOPendiente         string                   // nombre de IO en la que esta bloqueado
	IOPendienteDuracion time.Duration
	//Registros        RegistrosCPU
}

/*type RegistrosCPU struct {
	AX  uint8  // Registro Numérico de propósito general
	BX  uint8  // idem
	CX  uint8  // idem
	DX  uint8  // idem
	EAX uint32 // idem
	EBX uint32 // idem
	ECX uint32 // idem
	EDX uint32 // idem
	SI  uint32 // Dirección lógica de memoria de origen (string copy)
	DI  uint32 // Dirección lógica de memoria de destino (string copy)
}*/

type Proceso_a_enviar struct {
	PID     int
	Tamanio int
	PATH    string
}

type PIDyPC_Enviar_CPU struct {
	PID int
	PC  int
}

type Handshake struct {
	Puerto int
	IP     string
}
