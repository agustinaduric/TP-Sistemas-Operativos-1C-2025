package planificacion

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

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

var MutexPlanificadores sync.Mutex //asi se crea un mutex jej
var MutexLog sync.Mutex
var MutexCola sync.Mutex
var procesoCargado = make(chan int)
var procesoListo = make(chan int)
var MutexCpuDisponible sync.Mutex

// ----------- PLANIFICADOR CORTO PLAZO

func planificador_corto_plazo(configCargadito config.KernelConfig) {
	var pcb_execute structs.PCB
	var algoritmo_planificacion_corto t_algoritmo
	algoritmo_planificacion_corto = _chequear_algoritmo_corto(configCargadito)

	MutexPlanificadores.Lock()
	slog.Info("Algoritmo de planificación", "algoritmo", configCargadito.SchedulerAlgorithm)

	MutexPlanificadores.Unlock()

	for {

		<-procesoListo
		//var pidsito *int

		switch algoritmo_planificacion_corto {
		case FIFO:
			pcb_execute = pop_estado(&structs.ColaReady)
			structs.ProcesoEjecutando = pcb_execute
			//pcb_execute.Estado = structs.EXECT
			//push_estado(, pcb_execute)

			MutexLog.Lock()
			slog.Info("Estado cambiado", "PID", pcb_execute.PID, "EstadoAnterior", "READY", "EstadoActual", "EXEC")
			MutexLog.Unlock()

			//enviar_contexto_ejecucion(pcb_execute) FUNCIONES A CREAR PARA COMUNICACION CON CPU
			//recibir_contexto_ejecucion(pcb_execute)

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

	MutexPlanificadores.Lock()
	slog.Info("Iniciado planificador de largo plazo")
	MutexPlanificadores.Unlock()

	//go limpieza_cola_exit()  //funcion A CREAR para que se encargue de finalizar los procesos

	var algoritmo_planificacion t_algoritmo = _chequear_algoritmo_largo(configCargadito)

	for {
		<-procesoCargado

		switch algoritmo_planificacion {
		case FIFO:
			var pcb_a_cargar structs.PCB = structs.ColaNew[0]
			var respuesta int = enviar_proceso_a_memoria(pcb_a_cargar, configCargadito) // envio el proceso a memoria para preguntar si entra
			if respuesta == 200 {                                                       // ==200 si memoria confirmo, !=200 si hubo algun error
				pcb_a_cargar = pop_estado(&structs.ColaNew)   // saco de la cola NEW
				pcb_a_cargar.Estado = structs.READY           // pongo el estado en READY
				push_estado(&structs.ColaReady, pcb_a_cargar) // meto en la cola READY
				procesoListo <- 0                             // aviso al plani corto que tiene un proceso en ready
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
	MutexPlanificadores.Lock()
	slog.Info("Iniciado planificador de corto plazo")
	MutexPlanificadores.Unlock()
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

func pop_estado(Cola *structs.ColaProcesos) structs.PCB {
	MutexCola.Lock()
	defer MutexCola.Unlock()

	if len(*Cola) == 0 {
		return structs.PCB{} // o manejar el error como prefieras
	}

	pcb := (*Cola)[0]
	*Cola = (*Cola)[1:] // directamente recortás el slice

	return pcb
}

func push_estado(Cola *structs.ColaProcesos, pcb structs.PCB) {
	MutexCola.Lock()
	*Cola = append(*Cola, pcb) // Usamos el puntero a Cola para modificar el slice
	MutexCola.Unlock()
}

func enviar_proceso_a_memoria(pcb_a_cargar structs.PCB, configCargadito config.KernelConfig) int {
	var Proceso structs.Proceso_a_enviar = structs.Proceso_a_enviar{
		PID:     pcb_a_cargar.PID,
		Tamanio: pcb_a_cargar.Tamanio,
		PATH:    pcb_a_cargar.PATH,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/proceso", configCargadito.IpMemory, configCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, configCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	return resp.StatusCode
}

func enviar_datos_a_cpu(pcb_a_cargar structs.PCB, configCargadito config.KernelConfig) structs.DevolucionCpu {
	var PIDyPC structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: pcb_a_cargar.PID,
		PC:  pcb_a_cargar.PC,
	}
	MutexCpuDisponible.Lock()
	var Cpu_disponible structs.CPU = Buscar_CPU_libre()
	MutexCpuDisponible.Unlock()
	body, err := json.Marshal(PIDyPC)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/proceso", Cpu_disponible.Config.IpCPu, Cpu_disponible.Config.PortCpu)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, Cpu_disponible.Config.PortCpu)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	return resp.StatusCode
}

func esperarEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Presione ENTER para iniciar el planificador de largo plazo...")
	_, _ = reader.ReadString('\n') // espera hasta que se ingrese ENTER
}

func Buscar_CPU_libre() structs.CPU {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Disponible {
			return structs.CPUs_Conectados[i]
		}

	}
	log.Printf("No hay CPU's libres >:(")
	return structs.CPU{}
}
