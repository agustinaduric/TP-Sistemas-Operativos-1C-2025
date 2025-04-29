package planificacion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"

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

var MutexPlanificadores sync.Mutex //asi se crea un mutex jej
var MutexLog sync.Mutex
var MutexCola sync.Mutex

// ----------- PLANIFICADOR CORTO PLAZO

func planificador_corto_plazo() {
	var pcb_execute *structs.PCB
	var algoritmo_planificacion_corto t_algoritmo
	algoritmo_planificacion_corto = _chequear_algoritmo_corto()

	MutexPlanificadores.Lock()
	slog.Info("Algoritmo de planificación", "algoritmo", configKernel.AlgoritmoPlanificacion)
	MutexPlanificadores.Unlock()

	for {

		proceso_listo.Wait()
		var pidsito *int

		switch algoritmo_planificacion_corto; {
		case FIFO:
			pcb_execute = pop_estado(ColaReady)
			push_estado(exec, pcb_execute)

			MutexLog.Lock()
			// informar en un log
			MutexLog.Unlock()

			enviar_contexto_ejecucion(pcb_execute)
			recibir_contexto_ejecucion(pcb_execute)

		case SFJ:
		// PROXIMAMENTE

		case SRT:
		// PROXIMAMENTE

		default:

		}
	}
}

// ----------- PLANIFICADOR LARGO PLAZO

func planificador_largo_plazo() { // DIVIDIDO EN 2 PARTES: UNA PARA LLEVAR PROCESOS A READY Y OTRA PARA SACARLOS DE LA MEMORIA

	go limpieza_cola_exit() // funcion a crear para que se encargue de finalizar los procesos

	var algoritmo_planificacion t_algoritmo = _chequear_algoritmo_largo()

	for {
		<-procesoCargado

		switch algoritmo_planificacion; {
		case FIFO:
			var pcb_a_cargar *structs.PCB = structs.ColaNew[0]
			enviar_tamanio_proceso(pcb_a_cargar.Tamanio) // le envio el tamaño del proceso a memoria para saber si tiene lugar
			if recibir_confirmacion {                    // funcion tipo int que imita un bool (1=confirmo, 0=no confirmo)
				enviar_proceso_a_memoria(pcb_a_cargar.path, pcb_a_cargar.pid) // si tiene lugar le mando el path y el pid para que lo guarde
				pop_estado(structs.ColaNew)                                   // saco la cola new
				push_estado(pcb_a_cargar)                                     // meto en ready
				// falta avisar al semaforo de procesoListo que no se como hacerlo
			}

		case PMCP:
		// PROXIMAMENTE

		default:

		}
	}
}

// ----------- FUNCIONES SECUNDARIAS

func iniciar_planificacion() {
	// Creamos un hilo para el planificador de corto plazo y este va a mantener conexion con CPU

	go planificador_corto_plazo() //asi le ponemos un hilo

	// -------------------------------------------------
	MutexPlanificadores.Lock()
	slog.Info("Iniciado planificador de corto plazo")
	MutexPlanificadores.Unlock()
	// -------------------------------------------------
	// Creamos un hilo para el planificador de largo plazo
	go planificador_largo_plazo()
	// -------------------------------------------------
	MutexPlanificadores.Lock()
	slog.Info("Iniciado planificador de largo plazo")
	MutexPlanificadores.Unlock()
	// -------------------------------------------------
}

/*func inicializar_semaforos(){

    // ------- SEMAFOROS DE LOGS
    var MutexPlanificadores sync.Mutex //asi se crea un mutex jej
    var MutexLog sync.Mutex
	var MutexCola sync.Mutex

    // ------- SEMAFOROS DEL PLANIFICADOR DE CORTO PLAZO
	/*var proceso_listo = make(chan int, 0)
	var proceso_listo sync.WaitGroup
    var procesoCargado = make(chan struct{})

}*/

// ----------- FUNCIONES AUXILIARES

func _chequear_algoritmo_corto() t_algoritmo {
	if strings.EqualFold(configCargadito.scheduler_algorithm, "FIFO") {
		return FIFO
	}
	if strings.EqualFold(configCargadito.scheduler_algorithm, "SJF") {
		return SJF
	}
	if strings.EqualFold(configCargadito.scheduler_algorithm, "SRT") {
		return SRT
	}

	return ERROR
}

func _chequear_algoritmo_largo() t_algoritmo {
	if strings.EqualFold(configCargadito.ready_ingress_algorithm, "FIFO") {
		return FIFO
	}
	if strings.EqualFold(configCargadito.ready_ingress_algorithm, "PMCP") {
		return PMCP
	}

	return ERROR
}

func pop_estado(Cola *structs.ColaProcesos) *structs.PCB { ////////////go version
	MutexCola.Lock()
	var pcb *structs.PCB
	pcb = Cola[0]
	Cola = slice_remove(Cola, 0)
	MutexCola.Unlock()

	return pcb
}

func slice_remove(slice []int, index int) []int {
	return append(slice[:index], slice[index+1:]...)
}

func enviar_tamanio_proceso(tamanio int) {
	body, err := json.Marshal(tamanio)
	if err != nil {
		log.Printf("error codificando tamanio del proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/mensaje", configCargadito.IpMemory, configCargadito.PortMemory) // TODO: cambiar ruta. Esto no es un mensaje
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando tamanio del proceso:%s puerto:%d", configCargadito.IpMemory, configCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
}
