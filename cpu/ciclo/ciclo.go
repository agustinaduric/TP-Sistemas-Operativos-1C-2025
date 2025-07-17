package ciclo

import (
	"fmt"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/cache"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Ciclo() {
	global.Hubo_interrupcion = false
	global.Hubo_syscall = false
	global.CpuLogger.Debug("Inicio de ciclo")

	for {
		fetch() // busca la siguiente instrucción.
		decode_and_execute()
		CheckInterrupt() // se fija si hay interrupciones.
		if global.Hubo_syscall || global.Hubo_interrupcion {
			break
		}
	}
}

func fetch() {
	global.CpuLogger.Info(fmt.Sprintf("## PID: %d - FETCH - Program Counter: %d", global.Proceso_Ejecutando.PID, global.Proceso_Ejecutando.PC))
	protocolos.SolicitarInstruccion()
	// global.InstruccionRecibida <- 0 // espero que la memoria me mande la instruccion // creo que aca no es necesario
}

func decode_and_execute() {
	global.Instruccion_ejecutando = strings.Fields(global.Instruccion) // Dividimos las partes de la instruccion en un slice global.

	switch global.Instruccion_ejecutando[0] {
	case "WRITE":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: WRITE - %s - %s", global.Proceso_Ejecutando.PID, global.Instruccion_ejecutando[1], global.Instruccion_ejecutando[2]))
		var direccion int = global.String_a_int(global.Instruccion_ejecutando[1])
		WRITE(direccion, global.Instruccion_ejecutando[2])
		global.Proceso_Ejecutando.PC++
	case "READ":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: READ - %s - %s", global.Proceso_Ejecutando.PID, global.Instruccion_ejecutando[1], global.Instruccion_ejecutando[2]))
		var direccion int = global.String_a_int(global.Instruccion_ejecutando[1])
		var tamanio int = global.String_a_int(global.Instruccion_ejecutando[2])
		READ(direccion, tamanio)
		global.Proceso_Ejecutando.PC++
	case "NOOP":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: NOOP", global.Proceso_Ejecutando.PID))
		global.Proceso_Ejecutando.PC++
	case "GOTO":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: GOTO - %s", global.Proceso_Ejecutando.PID, global.Instruccion_ejecutando[1]))
		var numero int = global.String_a_int(global.Instruccion_ejecutando[1])
		global.Proceso_Ejecutando.PC = numero
	case "IO":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: IO - %s - %s", global.Proceso_Ejecutando.PID, global.Instruccion_ejecutando[1], global.Instruccion_ejecutando[2]))
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:    global.Proceso_Ejecutando.PID,
			PC:     global.Proceso_Ejecutando.PC,
			Motivo: "IO",
		}
		global.Proceso_Ejecutando.PC++
		protocolos.Enviar_syscall(devolucion)
		global.Hubo_syscall = true
	case "INIT_PROC":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: INIT_PROC - %s - %s", global.Proceso_Ejecutando.PID, global.Instruccion_ejecutando[1], global.Instruccion_ejecutando[2]))
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:           global.Proceso_Ejecutando.PID,
			PC:            global.Proceso_Ejecutando.PC,
			Motivo:        "INIT_PROC",
			ArchivoInst:   global.Instruccion_ejecutando[1],
			Tamaño:        global.String_a_int(global.Instruccion_ejecutando[2]),
			Identificador: global.Nombre,
		}
		protocolos.Enviar_syscall(devolucion)
		global.Proceso_Ejecutando.PC++
		<-global.Proceso_reconectado
		//global.Hubo_syscall = true no va porque tiene que volver el proceso
	case "DUMP_MEMORY":
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: DUMP_MEMORY", global.Proceso_Ejecutando.PID))
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:    global.Proceso_Ejecutando.PID,
			PC:     global.Proceso_Ejecutando.PC,
			Motivo: "DUMP_MEMORY",
		}
		global.Proceso_Ejecutando.PC++
		protocolos.Enviar_syscall(devolucion)
		global.Hubo_syscall = true
	case "EXIT":
		cache.LimpiarCacheDelProceso(global.Proceso_Ejecutando.PID)
		global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Ejecutando: EXIT", global.Proceso_Ejecutando.PID))
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:    global.Proceso_Ejecutando.PID,
			PC:     global.Proceso_Ejecutando.PC,
			Motivo: "EXIT",
		}
		global.Proceso_Ejecutando.PC++
		protocolos.Enviar_syscall(devolucion)
		global.Hubo_syscall = true
	}

}

func CheckInterrupt() {
	if global.Hayinterrupcion {
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:    global.Proceso_Ejecutando.PID,
			PC:     global.Proceso_Ejecutando.PC,
			Motivo: "REPLANIFICAR"}
		protocolos.Enviar_syscall(devolucion)
		global.Hubo_interrupcion = true
		return

	}
	return
}
