package structs

type TipoDeSyscall string

const (
	INIT_PROC   TipoDeSyscall = "INIT_PROC"
	IO_SYSCALL  TipoDeSyscall = "IO"
	DUMP_MEMORY TipoDeSyscall = "DUMP_MEMORY"
)

type Syscall struct {
	Tipo TipoDeSyscall
	Args []string // a checkear
	PID  int
}

type ColaProcesos []PCB

var (
	ColaNew         ColaProcesos
	ColaReady       ColaProcesos
	ColaBlocked     map[string]ColaProcesos // a checkear, el strin es por el io
	ColaSuspBlocked ColaProcesos
	ColaSuspReady   ColaProcesos
)

var IOsRegistrados map[string]*DispositivoIO
var ProcesoEjecutando PCB // chequear

// nota para desp -> despachar: ProcesoEjecutando = ... ¿?

var CPUs_Conectados []CPU



const (
	HANDSHAKE int = iota + 1
	CREAR_PROCESO
	ELIMINAR_PROCESO
	OBTENER_INSTRUCCION
	INSTRUCCION
	CONTEXTO_EJECUCION
	OPERACION_COMPLETADA
	AJUSTAR_TAMANIO
	OBTENER_NRO_MARCO
	MARCO
	LECTURA
	VALOR_LEIDO
	ESCRITURA
)