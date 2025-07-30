package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	var Cpu_disponible structs.CPU_a_kernel = Buscar_CPU_libre(pcb_a_cargar)
	global.MutexCpuDisponible.Unlock()
	
	if Cpu_disponible.Identificador == "" {

		return 0
	}
	global.KernelLogger.Debug(fmt.Sprintf("Se encontro la cpu libre: %s", Cpu_disponible.Identificador))
	body, err := json.Marshal(PIDyPC)
	if err != nil {
		global.KernelLogger.Error("error codificando el proceso")
	}
	url := fmt.Sprintf("http://%s:%d/datoCPU", Cpu_disponible.IP, Cpu_disponible.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, Cpu_disponible.Puerto))
	}
	global.KernelLogger.Debug(fmt.Sprintf("respuesta del servidor: %s , ENVIAR_DATOS_A_CPU", resp.Status))
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
	global.KernelLogger.Debug(fmt.Sprintf("respuesta del servidor: %s, Enviar_datos_SRT_a_cpu", resp.Status))

	return resp.StatusCode
}

func Mandar_interrupcion(Cpu structs.CPU_a_kernel) {
	var Interrupcion string = "interrupcion"
	body, err := json.Marshal(Interrupcion)
	if err != nil {
		global.KernelLogger.Error("error codificando la interrupcion")
	}
	url := fmt.Sprintf("http://%s:%d/interrupcion", Cpu.IP, Cpu.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando interrupcion al puerto:%d", Cpu.Puerto))
	}
	global.KernelLogger.Debug(fmt.Sprintf("respuesta del servidor: %s, Mandar Interrupcion", resp.Status))
	return
}

func Reconectarse_CPU(Cpu structs.CPU_a_kernel) {
	var Reconectarse string = "Reconectar" //aca esta el error
	body, err := json.Marshal(Reconectarse)
	if err != nil {
		global.KernelLogger.Error("error codificando el mensaje")
	}
	url := fmt.Sprintf("http://%s:%d/reconectar", Cpu.IP, Cpu.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando mensaje al puerto:%d", Cpu.Puerto))
	}
	global.KernelLogger.Debug(fmt.Sprintf("respuesta del servidor: %s", resp.Status))
	return
}

func Buscar_CPU_libre(Proceso structs.PCB) structs.CPU_a_kernel {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		indice := (i + global.Contador) % longitud // siempre da un numero entre 0 y longitud-1
		if structs.CPUs_Conectados[indice].Disponible {
			structs.CPUs_Conectados[indice].Disponible = false
			structs.CPUs_Conectados[indice].Proceso= Proceso
			global.Contador = (indice + 1) % longitud // la siguiente busqueda empezaria desde el que le sigue al seleccionado
			return structs.CPUs_Conectados[indice]
		}

	}

	global.KernelLogger.Debug("No hay CPU's libres >:(")
	return structs.CPU_a_kernel{}
}

func Buscar_CPU_Para_Desalojar(Proceso structs.PCB) structs.CPU_a_kernel {
	var Estimado float64
	if Proceso.Desalojado{Estimado = Proceso.TiempoRestante
	}else {Estimado = Proceso.EstimadoRafaga}
	global.KernelLogger.Debug(fmt.Sprintf("(%d) ingreso a Buscar_CPU_ParaDesalojar, Estimado: %f, Tiempo Restante: %f", Proceso.PID,Proceso.EstimadoRafaga, Proceso.TiempoRestante))

	
	global.KernelLogger.Debug(fmt.Sprintf("estoy comparando con : %f", Estimado))

	longitud := len(structs.CPUs_Conectados)
	var devuelvo structs.CPU_a_kernel = structs.CPU_a_kernel{}
	for i := 0; i < longitud; i++ {
		if !structs.CPUs_Conectados[i].Disponible{
		 if structs.CPUs_Conectados[i].Proceso.Desalojado{		
			aux := time.Since(structs.CPUs_Conectados[i].Proceso.TiempoInicioEstado)
			tiempo_restante := structs.CPUs_Conectados[i].Proceso.TiempoRestante -	(float64(aux))
			global.KernelLogger.Debug(fmt.Sprintf("VICTIMA Tiempo restante: %f", tiempo_restante))
			
			
			if tiempo_restante > Estimado {
				devuelvo = structs.CPUs_Conectados[i]
				Estimado = tiempo_restante
			}
		 }else{
			
		aux := time.Since(structs.CPUs_Conectados[i].Proceso.TiempoInicioEstado)
		global.KernelLogger.Debug(fmt.Sprintf("VICTIMA aux: %f", (float64(aux))))
		tiempo_restante := structs.CPUs_Conectados[i].Proceso.EstimadoRafaga - (float64(aux))
		global.KernelLogger.Debug(fmt.Sprintf("VICTIMA Tiempo restante: %f", tiempo_restante))
		if tiempo_restante > Estimado {

			devuelvo = structs.CPUs_Conectados[i]
			Estimado = tiempo_restante
		}
		}
		}	
	} 
	if devuelvo.Identificador != ""{
		
	}else{ global.KernelLogger.Debug("Ninguna cpu con menor estimado") }
	
	return devuelvo
}

func ActualizarCPU_Proceso(CPU structs.CPU_a_kernel,Proceso structs.PCB) {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Identificador == CPU.Identificador{ 
			structs.CPUs_Conectados[i].Disponible= false
			structs.CPUs_Conectados[i].Proceso = Proceso
		}
	}
	return
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
	global.KernelLogger.Debug(fmt.Sprintf("PID devuelto: %d", Devolucion.PID))
	proceso,_ := PCB.Buscar_por_pid(Devolucion.PID, &structs.ColaExecute)
	PCB.Actualizar_PC(proceso.PID, Devolucion.PC)
	proceso.PC = Devolucion.PC
	Cpu := Buscar_CPU(Devolucion.Identificador)
	switch Devolucion.Motivo {

	case structs.INIT_PROC:
		//proceso.Desalojado = false
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: INIT_PROC", proceso.PID))
		syscalls.INIT_PROC(Devolucion.ArchivoInst, Devolucion.Tamaño)
		global.KernelLogger.Debug(fmt.Sprintf("intentando reconectarse al cpu id: %s", Cpu.Identificador))
		go Reconectarse_CPU(Cpu) //intento 1 (bien, era el error)
		global.KernelLogger.Debug(fmt.Sprintf("se manda intento de reconectar"))

	case structs.DUMP_MEMORY:
		proceso.Desalojado = false
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: DUMP_MEMORY", proceso.PID))
		global.IniciarMetrica("EXEC", "BLOCKED", &proceso)
		go global.Habilitar_CPU_con_plani_corto(Cpu.Identificador)
		syscalls.DUMP_MEMORY(Devolucion.PID)

	case structs.IO:
		//proceso.Desalojado = false
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: IO", proceso.PID))
		syscalls.SolicitarSyscallIO((Devolucion.SolicitudIO) ,Cpu.Identificador)

	case structs.EXIT_PROC:
		proceso.Desalojado = false
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: EXIT", proceso.PID))
		global.IniciarMetrica("EXEC", "EXIT", &proceso)
		go global.Habilitar_CPU_con_plani_corto(Cpu.Identificador)

	case structs.REPLANIFICAR:
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT", proceso.PID))
		proceso.Desalojado = true
		sem := global.SemaforosCPU[Cpu.Identificador]
		global.KernelLogger.Debug(fmt.Sprintf("Mandando señal de cpu desalojada"))
		sem <- struct{}{}
		
		global.IniciarMetrica("EXEC", "READY", &proceso)
		
	case structs.REPLANIFICARPLUS:
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT", Devolucion.PID))
		sem := global.SemaforosCPU[Cpu.Identificador]
		global.KernelLogger.Debug(fmt.Sprintf("Mandando señal de cpu desalojada"))
		go global.Deshabilitar_CPU_con_plani_corto(Cpu.Identificador)
		sem <- struct{}{}
		
	}
	return
}
