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

// para el proceso que quiere usar la IO segun CPU:
func SolicitarSyscallIO(NuevaSolicitudIO structs.Solicitud) {
	// procedo a ver si existe la io
	_, hayMatch := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
	pcbSolicitante := structs.ProcesoEjecutando
	if hayMatch {
		dispositivo := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
		pcbSolicitante.Estado = structs.BLOCKED // lo mando a blocked: por esperar o por estar usando la io
		pcbSolicitante.IOPendiente = dispositivo.Nombre
		if dispositivo.PIDActual != 0 { // ocupado
			structs.ColaBlocked[NuevaSolicitudIO.NombreIO] = append(structs.ColaBlocked[NuevaSolicitudIO.NombreIO], pcbSolicitante) // agrego a cola de bloqueados por IO
		} else { // libre
			dispositivo.PIDActual = pcbSolicitante.PID // lo ocupo
			// cargo el struct y lo mando
			SolicitudParaIO := structs.Solicitud{PID: pcbSolicitante.PID, Duracion: NuevaSolicitudIO.Duracion, NombreIO: NuevaSolicitudIO.NombreIO}
			comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
		}
	} else {
		pcbSolicitante.IOPendiente = ""
		pcbSolicitante.Estado = structs.EXIT // no existe, se va a exit
		// cola exit para guardar los que terminaron ¿?
	}
}

func INIT_PROC(pseudocodigo string, tamanio int) {
	var nuevo_pcb structs.PCB = PCB.Crear(pseudocodigo, tamanio)
	nuevo_pcb.Estado = structs.NEW
	global.MutexNEW.Lock()
	PCB.Push_estado(&structs.ColaNew, nuevo_pcb)
	global.MutexNEW.Unlock()

	global.MutexLog.Lock()
	slog.Info("PID", ": ", nuevo_pcb.PID, "Se crea el proceso - Estado: NEW", "")
	global.MutexLog.Unlock()
}

func EXIT() {
	panic("unimplemented")
}

func Recibir_confirmacion_DumpMemory(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion_DumpMemory string
	err := decoder.Decode(&Devolucion_DumpMemory)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Printf("me llego una Devolucion de Memoria")
	switch Devolucion_DumpMemory {
	case "CONFIRMACION":

		//llamar al init_proceso
	case "ERROR":

		log.Println("El motivo es: Hacer un Dump Memory")
		//llamar al dump memory

	}
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

}
