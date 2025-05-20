package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Conectarse_con_CPU(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var CPUnuevo structs.CPU_a_kernel
	err := decoder.Decode(&CPUnuevo)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Printf("se conecto una CPU")
	log.Printf("identificador:: %s", CPUnuevo.Identificador)
	structs.CPUs_Conectados = append(structs.CPUs_Conectados, CPUnuevo)
}

func Enviar_datos_a_cpu(pcb_a_cargar structs.PCB) int {
	var PIDyPC structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: pcb_a_cargar.PID,
		PC:  pcb_a_cargar.PC,
	}
	global.MutexCpuDisponible.Lock()
	var Cpu_disponible structs.CPU_a_kernel = Buscar_CPU_libre()
	global.MutexCpuDisponible.Unlock()
	if Cpu_disponible.Identificador == "" {
		return 0
	}
	body, err := json.Marshal(PIDyPC)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/datoCPU", Cpu_disponible.IP, Cpu_disponible.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, Cpu_disponible.Puerto)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	return resp.StatusCode
}

func Buscar_CPU_libre() structs.CPU_a_kernel {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Disponible {
			structs.CPUs_Conectados[i].Disponible = false
			return structs.CPUs_Conectados[i]
		}

	}
	log.Printf("No hay CPU's libres >:(")
	return structs.CPU_a_kernel{}
}

func Recibir_devolucion_CPU(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion structs.DevolucionCpu
	err := decoder.Decode(&Devolucion)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Printf("me llego una Devolucion del CPU")
	log.Printf("PID devuelto: %d", Devolucion.PID)
	switch Devolucion.Motivo {

	case structs.INIT_PROC:
		log.Println("El motivo es: Crear Proceso")
		syscalls.INIT_PROC(Devolucion.ArchivoInst, Devolucion.Tamaño)
		//HAY QUE HACER QUE VUELVA EL PROCESO A EJECUTAR EN CPU

	case structs.DUMP_MEMORY:
		log.Println("El motivo es: Hacer un Dump Memory")

		syscalls.DUMP_MEMORY(Devolucion.PID)
	case structs.IO:

		log.Println("El motivo es: Sycall IO ")

		syscalls.SolicitarSyscallIO((Devolucion.SolicitudIO))

	case structs.EXIT_PROC:
		log.Println("El motivo es: EXIT")
		var pcb structs.PCB = PCB.Buscar_por_pid(Devolucion.PID, &structs.ColaExecute)

		global.MutexEXEC.Lock()
		PCB.Extraer_estado(&structs.ColaExecute, pcb.PID)
		global.MutexEXEC.Unlock()

		global.MutexEXEC.Lock()
		global.Push_estado(&structs.ColaExit, pcb)
		global.MutexEXEC.Unlock()

		pcb.Estado = structs.EXIT

		syscalls.EXIT()

	case structs.REPLANIFICAR:

	}
	return
}
