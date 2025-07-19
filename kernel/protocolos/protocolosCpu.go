package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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
		global.KernelLogger.Error(fmt.Sprintf("error al decodificar mensaje: %s\n", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	global.KernelLogger.Debug(fmt.Sprintf("se conecto una CPU"))
	global.KernelLogger.Debug(fmt.Sprintf("identificador: %s", CPUnuevo.Identificador))
	global.MutexCpuDisponible.Lock()
	structs.CPUs_Conectados = append(structs.CPUs_Conectados, CPUnuevo)
	global.MutexCpuDisponible.Unlock()

	global.MutexSemaforosCPU.Lock()
	if _, exists := global.SemaforosCPU[CPUnuevo.Identificador]; !exists {
		sem := make(chan struct{}, 1)
		global.SemaforosCPU[CPUnuevo.Identificador] = sem
	}
	global.MutexSemaforosCPU.Unlock()

}

func Enviar_datos_a_cpu(pcb_a_cargar structs.PCB) int {
	var PIDyPC structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: pcb_a_cargar.PID,
		PC:  pcb_a_cargar.PC,
	}
	global.KernelLogger.Debug(fmt.Sprintf("Entre a la funcion Enviar_datos_a_cpu"))
	global.MutexCpuDisponible.Lock()
	var Cpu_disponible structs.CPU_a_kernel = Buscar_CPU_libre()
	global.MutexCpuDisponible.Unlock()
	global.KernelLogger.Debug(fmt.Sprintf("Se encontro la cpu libre: %s", Cpu_disponible.Identificador))
	if Cpu_disponible.Identificador == "" {

		return 0
	}
	body, err := json.Marshal(PIDyPC)
	if err != nil {
		global.KernelLogger.Error("error codificando el proceso")
	}
	global.KernelLogger.Error("llegue a la parte 1")
	url := fmt.Sprintf("http://%s:%d/datoCPU", Cpu_disponible.IP, Cpu_disponible.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	global.KernelLogger.Error("llegue a la parte 2")
	if err != nil {
		global.KernelLogger.Error("llegue a la parte 3")
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, Cpu_disponible.Puerto))
	}
	log.Printf("respuesta del servidor: %s , ENVIAR_DATOS_A_CPU", resp.Status)
	global.KernelLogger.Error("llegue a la parte 5")
	/* var CPUocupado structs.CPU_nodisponible = structs.CPU_nodisponible{
		CPU:     Cpu_disponible,
		Proceso: pcb_a_cargar,
	}                                                    CREO QUE NO VA PORQUE ESTO NO LO VA A USAR SRT
	global.MutexCpuNoDisponibles.Lock()
	structs.CPUs_Nodisponibles = append(structs.CPUs_Nodisponibles, CPUocupado)
	global.MutexCpuNoDisponibles.Unlock()*/
	return resp.StatusCode
}

func Enviar_datos_SRT_a_cpu(pcb_a_cargar structs.PCB, Cpu_disponible structs.CPU_a_kernel) int {
	var PIDyPC structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: pcb_a_cargar.PID,
		PC:  pcb_a_cargar.PC,
	}
	body, err := json.Marshal(PIDyPC)
	if err != nil {
		global.KernelLogger.Error("error codificando el proceso")
	}
	url := fmt.Sprintf("http://%s:%d/datoCPU", Cpu_disponible.IP, Cpu_disponible.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, Cpu_disponible.Puerto))
	}
	log.Printf("respuesta del servidor: %s", resp.Status)

	var CPUocupado structs.CPU_nodisponible = structs.CPU_nodisponible{
		CPU:     Cpu_disponible,
		Proceso: pcb_a_cargar,
	}
	global.MutexCpuNoDisponibles.Lock()
	structs.CPUs_Nodisponibles = append(structs.CPUs_Nodisponibles, CPUocupado)
	global.MutexCpuNoDisponibles.Unlock()
	return resp.StatusCode
}

func Mandar_interrupcion(Cpu structs.CPU_a_kernel) {
	var Interrupcion string = "Interrupcion"
	body, err := json.Marshal(Interrupcion)
	if err != nil {
		global.KernelLogger.Error("error codificando la interrupcion")
	}
	url := fmt.Sprintf("http://%s:%d/interupcion", Cpu.IP, Cpu.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando interrupcion al puerto:%d", Cpu.Puerto))
	}
	global.KernelLogger.Debug(fmt.Sprintf("respuesta del servidor: %s", resp.Status))
	return
}

func Reconectarse_CPU(Cpu structs.CPU_a_kernel) {
	var Reconectarse string = "Reconectarse"
	body, err := json.Marshal(Reconectarse)
	if err != nil {
		global.KernelLogger.Error("error codificando el mensaje")
	}
	url := fmt.Sprintf("http://%s:%d/Reconectar", Cpu.IP, Cpu.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando mensaje al puerto:%d", Cpu.Puerto))
	}
	global.KernelLogger.Debug(fmt.Sprintf("respuesta del servidor: %s", resp.Status))
	return
}

func Buscar_CPU_libre() structs.CPU_a_kernel {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Disponible {
			structs.CPUs_Conectados[i].Disponible = false
			return structs.CPUs_Conectados[i]
		}

	}

	global.KernelLogger.Debug("No hay CPU's libres >:(")
	return structs.CPU_a_kernel{}
}

func Buscar_CPU_Para_Desalojar(Estimado float64) structs.CPU_a_kernel {
	longitud := len(structs.CPUs_Nodisponibles)
	var devuelvo structs.CPU_a_kernel = structs.CPU_a_kernel{}
	for i := 0; i < longitud; i++ {
		aux := time.Since(structs.CPUs_Nodisponibles[i].Proceso.TiempoInicioEstado)
		tiempo_restante := structs.CPUs_Nodisponibles[i].Proceso.EstimadoRafaga - (float64(aux) - float64(structs.CPUs_Nodisponibles[i].Proceso.Auxiliar))

		if tiempo_restante > Estimado {

			devuelvo = structs.CPUs_Nodisponibles[i].CPU
			Estimado = tiempo_restante
		}

	}

	global.KernelLogger.Debug("Ninguna cpu con menor estimado")
	return devuelvo
}

func Buscar_CPU(identificador string) structs.CPU_a_kernel {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Identificador == identificador {
			global.KernelLogger.Debug(fmt.Sprintf("Se encontro cpu con ese id :) id: %s", identificador))
			return structs.CPUs_Conectados[i]
		}

	}
	global.KernelLogger.Debug(fmt.Sprintf("No se encontro cpu con ese id >:( %s:", identificador))
	return structs.CPU_a_kernel{}
}

func Habilitar_CPU(identificador string) {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Identificador == identificador {
			structs.CPUs_Conectados[i].Disponible = true
			global.KernelLogger.Debug(fmt.Sprintf("Se habilita la cpu: %s", identificador))
			return
		}

	}
	return
}

func Recibir_devolucion_CPU(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion structs.DevolucionCpu
	err := decoder.Decode(&Devolucion)
	if err != nil {
		global.KernelLogger.Error("error al decodificar mensaje")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	global.KernelLogger.Debug("me llego una Devolucion del CPU")
	log.Printf("PID devuelto: %d", Devolucion.PID)
	global.KernelLogger.Debug(fmt.Sprintf("PID devuelto: %d", Devolucion.PID))

	proceso := PCB.Buscar_por_pid(Devolucion.PID, &structs.ColaExecute)
	proceso.PC = Devolucion.PC
	Cpu := Buscar_CPU(Devolucion.Identificador)
	switch Devolucion.Motivo {

	case structs.INIT_PROC:
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: INIT_PROC", proceso.PID))
		syscalls.INIT_PROC(Devolucion.ArchivoInst, Devolucion.Tamaño)
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: INIT_PROC", proceso.PID))
		global.KernelLogger.Debug(fmt.Sprintf("intentando reconectarse al cpu id: %s", Cpu.Identificador))
		Reconectarse_CPU(Cpu)
		global.KernelLogger.Debug(fmt.Sprintf("se logro reconectar"))

	case structs.DUMP_MEMORY:
		Habilitar_CPU(Cpu.Identificador)
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: DUMP_MEMORY", proceso.PID))
		global.IniciarMetrica("EXEC", "BLOCKED", &proceso)
		syscalls.DUMP_MEMORY(Devolucion.PID)

	case structs.IO:
		Habilitar_CPU(Cpu.Identificador)
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: IO", proceso.PID))
		syscalls.SolicitarSyscallIO((Devolucion.SolicitudIO))

	case structs.EXIT_PROC:
		Habilitar_CPU(Cpu.Identificador)
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: EXIT", proceso.PID))
		global.IniciarMetrica("EXEC", "EXIT", &proceso)

	case structs.REPLANIFICAR:
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT", proceso.PID))

		global.MutexSemaforosCPU.Lock()
		sem := global.SemaforosCPU[Cpu.Identificador]
		global.MutexSemaforosCPU.Unlock()

		sem <- struct{}{}
	}
	return
}
