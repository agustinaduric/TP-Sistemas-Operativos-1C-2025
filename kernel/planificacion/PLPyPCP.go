package planificacion

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"time"

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

func planificador_corto_plazo() {
	var pcb_execute structs.PCB
	algoritmo_planificacion_corto := global.ConfigCargadito.SchedulerAlgorithm

	global.KernelLogger.Debug(fmt.Sprintf("Algoritmo de planificación: %s", algoritmo_planificacion_corto))

	for {

		<-global.ProcesoListo

		switch algoritmo_planificacion_corto {
		case "FIFO":
			pcb_execute = structs.ColaReady[0]
			var respuesta int = protocolos.Enviar_datos_a_cpu(pcb_execute)
			if respuesta == 200 { // ==200 si memoria confirmo, !=200 si hubo algun error

				global.IniciarMetrica("READY", "EXEC", &pcb_execute)
				//structs.ProcesoEjecutando = pcb_execute       esto creo que ya no lo vamos a usar

			} else {

				global.KernelLogger.Debug(fmt.Sprintf("hubo un error: no se mando bien a cpu o no hay cpu libres"))
			}
		case "SJF":
		// PROXIMAMENTE

		case "SRT":
		// PROXIMAMENTE

		default:

		}
	}
}

// ----------- PLANIFICADOR LARGO PLAZO

func planificador_largo_plazo() { // DIVIDIDO EN 2 PARTES: UNA PARA LLEVAR PROCESOS A READY Y OTRA PARA SACARLOS DE LA MEMORIA

	esperarEnter() // no se si va aca o dentro de la funcion iniciar_planificacion

	global.KernelLogger.Debug("Planificador de largo plazo iniciado")

	go limpieza_cola_exit() //funcion que se encarga de finalizar los procesos

	algoritmo_planificacion := global.ConfigCargadito.ReadyIngressAlgorithm
	global.KernelLogger.Debug(fmt.Sprintf("el algoritmo de largo plazo es: %s", algoritmo_planificacion))

	for {
		<-global.ProcesoCargado
		global.KernelLogger.Debug("Llego un proceso al planificador largo")
		switch algoritmo_planificacion {
		case "FIFO":
			var pcb_a_cargar structs.PCB = structs.ColaNew[0]

			// envio el proceso a memoria para preguntar si entra
			if respuesta := protocolos.Enviar_proceso_a_memoria(pcb_a_cargar); respuesta == "OK" { // si memoria da el OK proceso, sino me salgo y espero
				global.KernelLogger.Debug("El proceso fue aceptado en memoria")
				global.IniciarMetrica("NEW", "READY", &pcb_a_cargar)

				global.ProcesoListo <- 0 // aviso al plani corto que tiene un proceso en ready
				global.KernelLogger.Debug("Se envia aviso desde plani largo a plani corto")
			}

		case "PMCP":
			OrdenarColaPorTamanio(structs.ColaNew)
			var pcb_a_cargar structs.PCB = structs.ColaNew[0]
			if respuesta := protocolos.Enviar_proceso_a_memoria(pcb_a_cargar); respuesta == "OK" { // si memoria da el OK proceso, sino me salgo y espero
				global.KernelLogger.Debug("El proceso fue aceptado en memoria")
				global.IniciarMetrica("NEW", "READY", &pcb_a_cargar)

				global.ProcesoListo <- 0 // aviso al plani corto que tiene un proceso en ready
				global.KernelLogger.Debug("Se envia aviso desde plani largo a plani corto")
			}

		default:

		}
	}
}

// ----------- PLANIFICADOR MEDIANO PLAZO

func planificador_mediano_plazo() {

	algoritmo_planificacion := global.ConfigCargadito.ReadyIngressAlgorithm
	global.KernelLogger.Debug(fmt.Sprintf("el algoritmo de mediano plazo es: %s", algoritmo_planificacion))

	for {
		<-global.ProcesoEnSuspReady
		global.KernelLogger.Debug("Llego un proceso al planificador mediano")
		switch algoritmo_planificacion {
		case "FIFO":
			var pcb_a_cargar structs.PCB = structs.ColaSuspReady[0]

			// envio el proceso a memoria para preguntar si entra
			if respuesta := protocolos.MandarProcesoADesuspension(pcb_a_cargar.PID); respuesta == "OK" { // si memoria da el OK proceso, sino me salgo y espero
				global.KernelLogger.Debug("El proceso que estaba suspendido fue aceptado en memoria")
				global.IniciarMetrica("SUSP_READY", "READY", &pcb_a_cargar)

				global.ProcesoListo <- 0 // aviso al plani corto que tiene un proceso en ready
				global.KernelLogger.Debug("Se envia aviso desde plani mediano a plani corto")
			}

		case "PMCP":
			OrdenarColaPorTamanio(structs.ColaSuspReady)
			var pcb_a_cargar structs.PCB = structs.ColaSuspReady[0]
			if respuesta := protocolos.MandarProcesoADesuspension(pcb_a_cargar.PID); respuesta == "OK" { // si memoria da el OK proceso, sino me salgo y espero
				global.KernelLogger.Debug("El proceso que estaba suspendido fue aceptado en memoria")
				global.IniciarMetrica("SUSP_READY", "READY", &pcb_a_cargar)

				global.ProcesoListo <- 0 // aviso al plani corto que tiene un proceso en ready
				global.KernelLogger.Debug("Se envia aviso desde plani mediano a plani corto")
			}

		default:

		}
	}
}

// ----------- FUNCIONES SECUNDARIAS

func Iniciar_planificacion(configCargadito config.KernelConfig) {
	// Creamos un hilo para el planificador de corto plazo y este va a mantener conexion con CPU

	go planificador_corto_plazo() //asi le ponemos un hilo
	// -------------------------------------------------
	global.KernelLogger.Debug("Planificador de corto plazo iniciado")
	// -------------------------------------------------

	// Creamos un hilo para el planificador de largo plazo
	go planificador_largo_plazo()
	// -------------------------------------------------

	// Creamos un hilo para el planificador de mediano plazo
	go planificador_mediano_plazo()
	// -------------------------------------------------
	global.KernelLogger.Debug("Planificador de mediano plazo iniciado")
}

// ----------- FUNCIONES AUXILIARES

func esperarEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Presione ENTER para iniciar el planificador de largo plazo...")
	_, _ = reader.ReadString('\n') // espera hasta que se ingrese ENTER
}

func limpieza_cola_exit() {
	for {

		<-global.ProcesoParaFinalizar // ACTIVAR ESTE SEMAFORO CADA VEZ QUE SE METE UN PROCESO EN LA COLA EXIT
		global.KernelLogger.Debug("Llega un proceso a la limpieza de cola exit")

		ProcesoExit := structs.ColaExit[0]

		protocolos.Enviar_P_Finalizado_memoria(ProcesoExit.PID)
		<-global.SemFinalizacion
		global.KernelLogger.Debug("memoria finalizo el proceso")
		if global.ConfirmacionProcesoFinalizado == 1 {
			global.ConfirmacionProcesoFinalizado = 0
			global.IniciarMetrica("EXIT", "FINALIZADO", &ProcesoExit)
			if len(structs.ColaSuspReady) == 0 {
				global.ProcesoCargado <- 0
				global.KernelLogger.Debug("Se envia aviso desde la limpieza de cola exit a plani largo, la cola de SUSP_READY estaba vacia")
				return
			}
			global.ProcesoEnSuspReady <- 0
			global.KernelLogger.Debug("Se envia aviso desde la limpieza de cola exit a plani mediano")

		}
	}
}

func IniciarContadorDeSuspension(proceso *structs.PCB) {
	global.KernelLogger.Debug("Entro a la funcion IniciarContadorDeSuspension")
	time.Sleep(time.Duration(global.ConfigCargadito.SuspensionTime) * time.Millisecond)
	global.KernelLogger.Debug(fmt.Sprintf("Termino el conteo de suspension del proceso de PID: %d", proceso.PID))
	if proceso.Estado != structs.BLOCKED {
		return
	} //si el estado ya no es BLOCKED no hago nada y finalizo el hilo contador

	//si el estado sigue siendo blocked lo mando a memoria para que lo swapee y cambio su estado a SUSP_BLOCKED
	global.IniciarMetrica("BLOCKED", "SUSP_BLOCKED", proceso)
	protocolos.MandarProcesoASuspension(proceso.PID)

}

func OrdenarColaPorTamanio(cola structs.ColaProcesos) {
	sort.Slice(cola, func(i, j int) bool {
		if cola[i].Tamanio != cola[j].Tamanio {
			return cola[i].Tamanio < cola[j].Tamanio
		}
		return cola[i].IngresoEstado.Before(cola[j].IngresoEstado)
	})
}
