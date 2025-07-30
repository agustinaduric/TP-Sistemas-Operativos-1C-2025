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

func SolicitarSyscallIO(NuevaSolicitudIO structs.Solicitud, identificador_cpu string) {
	pcbSolicitante, existe := PCB.Buscar_por_pid(NuevaSolicitudIO.PID, &structs.ColaExecute)
	pcbSolicitante.Desalojado=false
	if !existe {
		global.KernelLogger.Error(fmt.Sprintf("No existe PID %d en excute", NuevaSolicitudIO.PID))
		return
	}
	global.KernelLogger.Debug(fmt.Sprintf("El proceso: %d solicita io: %s", NuevaSolicitudIO.PID, NuevaSolicitudIO.NombreIO)) // muestra PID Solicitud
	_, hayMatch := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]

	if !hayMatch { // no esta registrado
		global.KernelLogger.Debug(fmt.Sprintf("No existe IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID)) // muestra PID encontrado
		global.IniciarMetrica("EXEC", "EXIT", &pcbSolicitante)
		go global.Habilitar_CPU_con_plani_corto(identificador_cpu)
		structs.ProcesoEjecutando = structs.PCB{}
		return
	}
	global.KernelLogger.Debug(fmt.Sprintf("Existe IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
	pcbSolicitante.IOPendiente = NuevaSolicitudIO.NombreIO
	pcbSolicitante.IOPendienteDuracion = NuevaSolicitudIO.Duracion

	dispositivoLibre := BuscarIOLibre(NuevaSolicitudIO.NombreIO)

	if dispositivoLibre == nil { // estan todos ocupados
		global.KernelLogger.Debug(fmt.Sprintf("IO ocupado: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		global.IniciarMetrica("EXEC", "BLOCKED", &pcbSolicitante)
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Bloqueado por IO: %s", pcbSolicitante.PID, NuevaSolicitudIO.NombreIO))
		go global.Habilitar_CPU_con_plani_corto(identificador_cpu)
		global.MutexBLOCKED_IO.Lock()
		colaDeBloqueados := structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO]
		global.Push_estado(&colaDeBloqueados, pcbSolicitante)
		structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO] = colaDeBloqueados
		global.MutexBLOCKED_IO.Unlock()
	} else { // libre
		global.KernelLogger.Debug(fmt.Sprintf("IO libre: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		dispositivoLibre.PIDActual = pcbSolicitante.PID
		SolicitudParaIO := structs.Solicitud{PID: pcbSolicitante.PID, NombreIO: NuevaSolicitudIO.NombreIO, Duracion: NuevaSolicitudIO.Duracion}
		global.KernelLogger.Debug(fmt.Sprintf("Se intenta enviar solicitud a IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID))
		comunicacion.EnviarSolicitudIO(dispositivoLibre.IP, dispositivoLibre.Puerto, SolicitudParaIO) //esto puede ser un go
		global.KernelLogger.Debug(fmt.Sprintf("Se envio solicitud a IO: %s, PID: %d", NuevaSolicitudIO.NombreIO, pcbSolicitante.PID)) //este mensaje capaz va adentro de la funcion de arriba
		global.IniciarMetrica("EXEC", "BLOCKED", &pcbSolicitante)
		global.KernelLogger.Info(fmt.Sprintf("## (%d) - Bloqueado por IO: %s", pcbSolicitante.PID, NuevaSolicitudIO.NombreIO))
		go global.Habilitar_CPU_con_plani_corto(identificador_cpu)
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
	Proceso, _ = PCB.Buscar_por_pid(Devolucion_DumpMemory.PID, &structs.ColaBlocked)
	global.KernelLogger.Debug(fmt.Sprintf("me llego una Devolucion de Memoria: %s", Devolucion_DumpMemory.Respuesta))
	switch Devolucion_DumpMemory.Respuesta {
	case "OK":
		global.IniciarMetrica("BLOCKED", "READY", &Proceso)

	default:
		global.IniciarMetrica("BLOCKED", "EXIT", &Proceso)

	}

}

//----------------------------------- funcion axuliar para syscall io -----------------------------------//

func BuscarIOLibre(nombreIO string) *structs.DispositivoIO {
	dispositivos, existe := structs.IOsRegistrados[nombreIO]
	if !existe {
		return nil
	}
	for _, dispositivo := range dispositivos {
		if dispositivo.PIDActual == -1 {
			return dispositivo
		}
	}
	return nil
}
