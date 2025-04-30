package planificacion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
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

		<-global.ProcesoListo
		//var pidsito *int

		switch algoritmo_planificacion_corto {
		case FIFO:

			var respuesta int = protocolos.Enviar_datos_a_cpu(pcb_execute)
			if respuesta == 200 { // ==200 si memoria confirmo, !=200 si hubo algun error
				pcb_execute = pop_estado(&structs.ColaReady)
				pcb_execute.Estado = structs.EXEC
				structs.ProcesoEjecutando = pcb_execute
				push_estado(&structs.ColaReady, pcb_execute)
				global.MutexLog.Lock()
				slog.Info("Estado cambiado", "PID", pcb_execute.PID, "EstadoAnterior", "READY", "EstadoActual", "EXEC")
				global.MutexLog.Unlock()
			} else {
				log.Printf("hubo un error: no se mando bien a cpu o no hay cpu libres")
			}

			go recibir_devolucion_CPU() // crep que esto no va aca

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

	//go limpieza_cola_exit()  //funcion A CREAR para que se encargue de finalizar los procesos

	var algoritmo_planificacion t_algoritmo = _chequear_algoritmo_largo(configCargadito)

	for {
		<-global.ProcesoCargado

		switch algoritmo_planificacion {
		case FIFO:
			var pcb_a_cargar structs.PCB = structs.ColaNew[0]
			protocolos.Enviar_proceso_a_memoria(pcb_a_cargar, configCargadito) // envio el proceso a memoria para preguntar si entra
			if protocolos.Recibir_confirmacion() {                             // ==200 si memoria confirmo, !=200 si hubo algun error
				pcb_a_cargar = pop_estado(&structs.ColaNew)   // saco de la cola NEW
				pcb_a_cargar.Estado = structs.READY           // pongo el estado en READY
				push_estado(&structs.ColaReady, pcb_a_cargar) // meto en la cola READY
				global.ProcesoListo <- 0                      // aviso al plani corto que tiene un proceso en ready
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

func pop_estado(Cola *structs.ColaProcesos) structs.PCB {
	global.MutexCola.Lock()
	defer global.MutexCola.Unlock()

	if len(*Cola) == 0 {
		return structs.PCB{} // o manejar el error como prefieras
	}

	pcb := (*Cola)[0]
	*Cola = (*Cola)[1:] // directamente recortás el slice

	return pcb
}

func push_estado(Cola *structs.ColaProcesos, pcb structs.PCB) {
	global.MutexCola.Lock()
	*Cola = append(*Cola, pcb) // Usamos el puntero a Cola para modificar el slice
	global.MutexCola.Unlock()
}

func esperarEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Presione ENTER para iniciar el planificador de largo plazo...")
	_, _ = reader.ReadString('\n') // espera hasta que se ingrese ENTER
}

func recibir_devolucion_CPU(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion structs.DevolucionCpu
	err := decoder.Decode(&Devolucion)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Println("me llego una Devolucion del CPU")
	log.Println("PID devuelto: %d", Devolucion.PID)
	switch Devolucion.Motivo {
	case structs.INICIAR_PROCESO:
		log.Println("El motivo es: Crear Proceso")

	case structs.ELIMINAR_PROCESO:

		log.Println("El motivo es: Eliminar Proceso")

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	return
}
