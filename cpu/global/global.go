package global

import (
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var ConfigCargadito config.CPUConfig

var Instruccion string
var Instruccion_ejecutando []string

var Proceso_Ejecutando structs.PIDyPC_Enviar_CPU

var InstruccionRecibida = make(chan int)

var Hubo_syscall bool
