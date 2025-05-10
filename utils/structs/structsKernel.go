package structs

type TipoDeSyscall string

/*const (
	INIT_PROC   TipoDeSyscall = "INIT_PROC"
	IO_SYSCALL  TipoDeSyscall = "IO"
	DUMP_MEMORY TipoDeSyscall = "DUMP_MEMORY"
) */

type Syscall struct {
	Tipo TipoDeSyscall
	Args []string // a checkear
	PID  int
}

type ColaProcesos []PCB

var (
	ColaNew         ColaProcesos
	ColaReady       ColaProcesos
	ColaBlockedIO   map[string]ColaProcesos // a checkear, el strin es por el io
	ColaBlocked     ColaProcesos            // a checkear, el strin es por el io
	ColaSuspBlocked ColaProcesos
	ColaSuspReady   ColaProcesos
	ColaExecute     ColaProcesos
	ColaExit        ColaProcesos
)

var IOsRegistrados map[string]*DispositivoIO
var ProcesoEjecutando PCB // chequear

// nota para desp -> despachar: ProcesoEjecutando = ... ¿?

var CPUs_Conectados []CPU_a_kernel
