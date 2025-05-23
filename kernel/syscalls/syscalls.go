package syscalls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func SolicitarSyscallIO(NuevaSolicitudIO structs.Solicitud) {
	pcbSolicitante := global.Extraer_estado(&structs.ColaExecute, structs.ProcesoEjecutando.PID)
	_, hayMatch := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
	dispositivo := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
	if !hayMatch {
		pcbSolicitante.Estado = structs.EXIT
		pcbSolicitante.IOPendiente = ""
		global.MutexEXIT.Lock()
		global.Push_estado(&structs.ColaExit, pcbSolicitante)
		global.MutexEXIT.Unlock()
		structs.ProcesoEjecutando = structs.PCB{}
		return
	}
	pcbSolicitante.Estado = structs.BLOCKED // lo mando a blocked: por esperar o por estar usando la io
	pcbSolicitante.IOPendiente = dispositivo.Nombre
	pcbSolicitante.IOPendienteDuracion = NuevaSolicitudIO.Duracion
	if dispositivo.PIDActual != 0 { // ocupado
		global.MutexBLOCKED.Lock()
		colaDeBloqueados := structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO]
		global.Push_estado(&colaDeBloqueados, pcbSolicitante)
		structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO] = colaDeBloqueados
		global.MutexBLOCKED.Unlock()
	} else { // libre
		dispositivo.PIDActual = pcbSolicitante.PID
		SolicitudParaIO := structs.Solicitud{PID: pcbSolicitante.PID, NombreIO: NuevaSolicitudIO.NombreIO, Duracion: NuevaSolicitudIO.Duracion}
		comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
	}
	structs.ProcesoEjecutando = structs.PCB{}
}

func INIT_PROC(pseudocodigo string, tamanio int) {
	var nuevo_pcb structs.PCB = PCB.Crear(pseudocodigo, tamanio)

	global.IniciarMetrica("", "NEW", &nuevo_pcb)

	global.MutexLog.Lock()
	slog.Info("PID", ": ", nuevo_pcb.PID, "Se crea el proceso - Estado: NEW", "")
	global.MutexLog.Unlock()
}

func DUMP_MEMORY(PID int) {

	body, err := json.Marshal(PID)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/memorydump", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", PID, global.ConfigCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	//return resp.StatusCode
	return
}

func EXIT() {
	<-global.ProcesoParaFinalizar
}
