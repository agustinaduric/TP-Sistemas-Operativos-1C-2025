package planificacion

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

type t_algoritmo int

const (
	FIFO t_algoritmo = iota // iota vale
	SJF
	SRT
	PMCP
	ERROR
)

// ----------- PLANIFICADOR CORTO PLAZO

func planificador_corto_plazo(configCargadito config.KernelConfig) {
	var pcb_execute structs.PCB
	var algoritmo_planificacion_corto t_algoritmo
	algoritmo_planificacion_corto = _chequear_algoritmo_corto(configCargadito)

	global.MutexPlanificadores.Lock()
	slog.Info("Algoritmo de planificación", "algoritmo", configCargadito.SchedulerAlgorithm)

	global.MutexPlanificadores.Unlock()

	for {

		global.ProcesoListo <- 0
		//var pidsito *int

		switch algoritmo_planificacion_corto {
		case FIFO:

			var respuesta int = protocolos.Enviar_datos_a_cpu(pcb_execute)
			if respuesta == 200 { // ==200 si memoria confirmo, !=200 si hubo algun error

				global.MutexREADY.Lock()
				pcb_execute = global.Pop_estado(&structs.ColaReady)
				global.MutexREADY.Unlock()

				global.MutexEXEC.Lock()
				global.Push_estado(&structs.ColaExecute, pcb_execute)
				pcb_execute.Estado = structs.EXEC
				//structs.ProcesoEjecutando = pcb_execute       esto creo que ya no lo vamos a usar

				global.MutexLog.Lock()
				slog.Info("Estado cambiado", "PID", pcb_execute.PID, "EstadoAnterior", "READY", "EstadoActual", "EXEC")
				global.MutexLog.Unlock()
			} else {
				log.Printf("hubo un error: no se mando bien a cpu o no hay cpu libres")
			}
		case SJF:
		// PROXIMAMENTE

		case SRT:
		// PROXIMAMENTE

		default:

		}
	}
}

// ----------- PLANIFICADOR LARGO PLAZO

func planificador_largo_plazo(configCargadito config.KernelConfig) { // DIVIDIDO EN 2 PARTES: UNA PARA LLEVAR PROCESOS A READY Y OTRA PARA SACARLOS DE LA MEMORIA

	esperarEnter() // no se si va aca o dentro de la funcion iniciar_planificacion

	global.MutexPlanificadores.Lock()
	slog.Info("Iniciado planificador de largo plazo")
	global.MutexPlanificadores.Unlock()

	go limpieza_cola_exit() //funcion que se encarga de finalizar los procesos

	var algoritmo_planificacion t_algoritmo = _chequear_algoritmo_largo(configCargadito)

	for {
		global.ProcesoCargado <- 0

		switch algoritmo_planificacion {
		case FIFO:
			var pcb_a_cargar structs.PCB = structs.ColaNew[0]
			protocolos.Enviar_proceso_a_memoria(pcb_a_cargar, configCargadito) // envio el proceso a memoria para preguntar si entra
			global.SemInicializacion <- 0
			if global.ConfirmacionProcesoCargado == 1 { // si memoria da el OK proceso, sino me salgo y espero
				global.ConfirmacionProcesoCargado = 0

				global.MutexNEW.Lock()
				pcb_a_cargar = global.Pop_estado(&structs.ColaNew) // saco de la cola NEW
				global.MutexNEW.Unlock()

				pcb_a_cargar.Estado = structs.READY // pongo el estado en READY
				global.MutexREADY.Lock()
				global.Push_estado(&structs.ColaReady, pcb_a_cargar) // meto en la cola READY
				global.MutexREADY.Unlock()

				global.MutexLog.Lock()
				slog.Info("Estado cambiado", "PID", pcb_a_cargar.PID, "EstadoAnterior", "NEW", "EstadoActual", "READY")
				global.MutexLog.Unlock()

				<-global.ProcesoListo // aviso al plani corto que tiene un proceso en ready
			}

		case PMCP:
		// PROXIMAMENTE

		default:

		}
	}
}

// ----------- FUNCIONES SECUNDARIAS

func Iniciar_planificacion(configCargadito config.KernelConfig) {
	// Creamos un hilo para el planificador de corto plazo y este va a mantener conexion con CPU

	go planificador_corto_plazo(configCargadito) //asi le ponemos un hilo

	// -------------------------------------------------
	global.MutexPlanificadores.Lock()
	slog.Info("Iniciado planificador de corto plazo")
	global.MutexPlanificadores.Unlock()
	// -------------------------------------------------
	// Creamos un hilo para el planificador de largo plazo
	go planificador_largo_plazo(configCargadito)
	// -------------------------------------------------
}

// ----------- FUNCIONES AUXILIARES

func _chequear_algoritmo_corto(configCargadito config.KernelConfig) t_algoritmo {
	if strings.EqualFold(configCargadito.SchedulerAlgorithm, "FIFO") {
		return FIFO
	}
	if strings.EqualFold(configCargadito.SchedulerAlgorithm, "SJF") {
		return SJF
	}
	if strings.EqualFold(configCargadito.SchedulerAlgorithm, "SRT") {
		return SRT
	}

	return ERROR
}

func _chequear_algoritmo_largo(configCargadito config.KernelConfig) t_algoritmo {
	if strings.EqualFold(configCargadito.ReadyIngressAlgorithm, "FIFO") {
		return FIFO
	}
	if strings.EqualFold(configCargadito.ReadyIngressAlgorithm, "PMCP") {
		return PMCP
	}

	return ERROR
}

func esperarEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Presione ENTER para iniciar el planificador de largo plazo...")
	_, _ = reader.ReadString('\n') // espera hasta que se ingrese ENTER
}

func limpieza_cola_exit() {
	for {

		global.ProcesoParaFinalizar <- 0 // ACTIVAR ESTE SEMAFORO CADA VEZ QUE SE METE UN PROCESO EN LA COLA EXIT

		global.MutexEXIT.Lock()
		ProcesoExit := global.Pop_estado(&structs.ColaExit)
		global.MutexEXIT.Unlock()

		protocolos.Enviar_P_Finalizado_memoria(ProcesoExit.PID)
		global.SemFinalizacion <- 0
		if global.ConfirmacionProcesoFinalizado == 1 {
			global.ConfirmacionProcesoFinalizado = 0
			//LOGGEAR METRICAS Y LOG OBLIGATORIO DE FINALIZACION DE PROCESO
			<-global.ProcesoCargado
		}
	}
}
