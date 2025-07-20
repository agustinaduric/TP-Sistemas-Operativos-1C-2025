package syscalls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func SolicitarSyscallIO(NuevaSolicitudIO structs.Solicitud) {
	pcbSolicitante := PCB.Buscar_por_pid(NuevaSolicitudIO.PID, &structs.ColaExecute) // esto es una copia(?
	global.KernelLogger.Debug(fmt.Sprintf("El proceso: %d solicita io: %s", NuevaSolicitudIO.PID, NuevaSolicitudIO.NombreIO))
	_, hayMatch := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
	if !hayMatch {
		global.KernelLogger.Debug(fmt.Sprintf("No existe IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		global.IniciarMetrica("EXEC", "EXIT", &pcbSolicitante)
		structs.ProcesoEjecutando = structs.PCB{}
		return
	}
	global.KernelLogger.Debug(fmt.Sprintf("Existe IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
	dispositivo := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
	pcbSolicitante.IOPendiente = dispositivo.Nombre
	pcbSolicitante.IOPendienteDuracion = NuevaSolicitudIO.Duracion
	if dispositivo.PIDActual != -1 { // ocupado
		global.KernelLogger.Debug(fmt.Sprintf("IO ocupado: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		global.IniciarMetrica("EXEC", "BLOCKED", &pcbSolicitante)
		global.MutexBLOCKED.Lock()
		colaDeBloqueados := structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO]
		global.Push_estado(&colaDeBloqueados, pcbSolicitante)
		structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO] = colaDeBloqueados
		global.MutexBLOCKED.Unlock()
	} else { // libre
		global.KernelLogger.Debug(fmt.Sprintf("IO libre: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		global.MutexBLOCKED.Lock()
		cola := structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO]
		global.Push_estado(&cola, pcbSolicitante)
		structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO] = cola
		global.MutexBLOCKED.Unlock()
		dispositivo.PIDActual = pcbSolicitante.PID
		SolicitudParaIO := structs.Solicitud{PID: pcbSolicitante.PID, NombreIO: NuevaSolicitudIO.NombreIO, Duracion: NuevaSolicitudIO.Duracion}
		global.KernelLogger.Debug(fmt.Sprintf("Se intenta enviar solicitud a IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
		global.IniciarMetrica("EXEC", "BLOCKED", &pcbSolicitante)
		global.KernelLogger.Debug(fmt.Sprintf("Se envio solicitud a IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID)) //este mensaje capaz va adentro de la funcion de arriba
	}
	structs.ProcesoEjecutando = structs.PCB{}
}

func INIT_PROC(pseudocodigo string, tamanio int) {
	var nuevo_pcb structs.PCB = PCB.Crear(pseudocodigo, tamanio)

	global.IniciarMetrica("", "NEW", &nuevo_pcb)
}

func DUMP_MEMORY(PID int) {

	body, err := json.Marshal(PID)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/memory-dump", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", PID, global.ConfigCargadito.PortMemory))
	}

	defer resp.Body.Close()
	var Devolucion_DumpMemory structs.Devolucion_DumpMemory
	json.NewDecoder(resp.Body).Decode(&Devolucion_DumpMemory)

	var Proceso structs.PCB
	Proceso = PCB.Buscar_por_pid(Devolucion_DumpMemory.PID, &structs.ColaBlocked)
	global.KernelLogger.Debug(fmt.Sprintf("me llego una Devolucion de Memoria"))
	switch Devolucion_DumpMemory.Respuesta {
	case "CONFIRMACION":
		global.IniciarMetrica("BLOCKED", "READY", &Proceso)

	case "ERROR":
		global.IniciarMetrica("BLOCKED", "EXIT", &Proceso)

	}

}
