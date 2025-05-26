package ciclo

import (
	"strings"

	mmu "github.com/sisoputnfrba/tp-golang/cpu/MMU"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Ciclo() {
	//se_devolvio_contexto = 0
	global.Hubo_syscall = false

	for {
		fetch() // busca la siguiente instrucción.
		decode_and_execute()
		if global.Hubo_syscall {
			break
		}
		CheckInterrupt() // se fija si hay interrupciones.
		//if(se_devolvio_contexto){ break}

	}
}

func fetch() {
	protocolos.SolicitarInstruccion()
	// global.InstruccionRecibida <- 0 // espero que la memoria me mande la instruccion // creo que aca no es necesario
}

func decode_and_execute() {
	global.Instruccion_ejecutando = strings.Fields(global.Instruccion) // Dividimos las partes de la instruccion en un slice global.
	if global.Instruccion_ejecutando[0] == "WRITE" || global.Instruccion_ejecutando[0] == "READ" {
		var direccionfisica string = mmu.DL_a_DF(global.Instruccion_ejecutando[1])
		global.Instruccion_ejecutando[1] = direccionfisica
	}

	switch global.Instruccion_ejecutando[0] {
	case "WRITE":
		var direccion int = global.String_a_int(global.Instruccion_ejecutando[1])
		WRITE(direccion, global.Instruccion_ejecutando[2])
		global.Proceso_Ejecutando.PC++
	case "READ":
		var direccion int = global.String_a_int(global.Instruccion_ejecutando[1])
		var tamanio int = global.String_a_int(global.Instruccion_ejecutando[2])
		READ(direccion, tamanio)
		global.Proceso_Ejecutando.PC++
	case "NOOP":
		// no hace nada
		global.Proceso_Ejecutando.PC++
	case "GOTO":
		var numero int = global.String_a_int(global.Instruccion_ejecutando[1])
		global.Proceso_Ejecutando.PC = numero
	case "IO":
		//
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:    global.Proceso_Ejecutando.PID,
			PC:     global.Proceso_Ejecutando.PC,
			Motivo: "IO",
		}
		global.Proceso_Ejecutando.PC++
		protocolos.Enviar_syscall(devolucion)
		global.Hubo_syscall = true
	case "INIT_PROC":
		//
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
		//
		var devolucion structs.DevolucionCpu = structs.DevolucionCpu{
			PID:    global.Proceso_Ejecutando.PID,
			PC:     global.Proceso_Ejecutando.PC,
			Motivo: "DUMP_MEMORY",
		}
		global.Proceso_Ejecutando.PC++
		protocolos.Enviar_syscall(devolucion)
		global.Hubo_syscall = true
	case "EXIT":
		//
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
		return
		// aca el tp dice que deberia esperar la cpu xd
	}
	return
}
