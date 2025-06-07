package protocolos

import(
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
)

func HandlerRegistrarIO(w http.ResponseWriter, r *http.Request) {
	var registro structs.RegistroIO
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&registro)
	if err != nil {
		http.Error(w, "Error en decodificar mje: "+err.Error(), http.StatusBadRequest)
		return
	}
	nuevoIO := structs.DispositivoIO{
		Nombre:       registro.Nombre,
		IP:           registro.IP,
		Puerto:       registro.Puerto,
		PIDActual:    0,
		ColaEsperaIO: []*structs.PCB{},
	}
	structs.IOsRegistrados[registro.Nombre] = &nuevoIO
	global.KernelLogger.Debug(fmt.Sprintf("se registro el io: %s en kernel", registro.Nombre))
}

func HandlerFinalizarIO(w http.ResponseWriter, r *http.Request) {
	var respuestaFin structs.RespuestaIO
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&respuestaFin)
	if err != nil {
		http.Error(w, "Error en decodificar la respuesta de io: "+err.Error(), http.StatusBadRequest)
		return
	}
	cola := structs.ColaBlockedIO[respuestaFin.NombreIO]
	proceso := PCB.Buscar_por_pid(respuestaFin.PID, &cola) // esto es una copia(?
	structs.ColaBlockedIO[respuestaFin.NombreIO] = cola
	global.IniciarMetrica("BLOCKED", "READY", &proceso)

	dispositivo := structs.IOsRegistrados[respuestaFin.NombreIO]
	dispositivo.PIDActual = 0
	global.KernelLogger.Debug(fmt.Sprintf("El dispositivo %s esta libre", respuestaFin.NombreIO))
	if len(structs.ColaBlockedIO[respuestaFin.NombreIO]) > 0 {
		siguiente := structs.ColaBlockedIO[respuestaFin.NombreIO][0]
		structs.ColaBlockedIO[respuestaFin.NombreIO] = structs.ColaBlockedIO[respuestaFin.NombreIO][1:]
		dispositivo.PIDActual = siguiente.PID
		global.KernelLogger.Debug(fmt.Sprintf("PID: %d ocupo el dispositivo: %s", siguiente.PID, respuestaFin.NombreIO))
		SolicitudParaIO := structs.Solicitud{PID: siguiente.PID, NombreIO: dispositivo.Nombre, Duracion: siguiente.IOPendienteDuracion}
		comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
		global.KernelLogger.Debug(fmt.Sprintf("Solcitud IO: %s enviada, PID: %d", respuestaFin.NombreIO, siguiente.PID))
	}
	global.KernelLogger.Debug(fmt.Sprintf("NO hay procesos en espera para: %s", respuestaFin.NombreIO))
}

func HandlerDesconexionIO(w http.ResponseWriter, r *http.Request){
	var ioDesconectado structs.IODesconectado
	err := json.NewDecoder(r.Body).Decode(&ioDesconectado)
	if err != nil {
		http.Error(w, "Error en decodificar la desconexion de io: "+err.Error(), http.StatusBadRequest)
		return
	}
	global.KernelLogger.Debug(fmt.Sprintf("Recibi la desconexion de: %s", ioDesconectado.Nombre))
	dispositivo := structs.IOsRegistrados[ioDesconectado.Nombre]
	// si hay uno ejecutando, lo mato
	if dispositivo.PIDActual !=0 {
		proceso := PCB.Buscar_por_pid(dispositivo.PIDActual, &structs.ColaBlocked)
		global.IniciarMetrica("BLOCKED", "EXIT", &proceso)
		global.KernelLogger.Debug(fmt.Sprintf("EXIT PID: %d por estar ejecutando en io desconectado", dispositivo.PIDActual))
	}
	// mato a los procesos que quedaron esperando
	cola := structs.ColaBlockedIO[ioDesconectado.Nombre]
	global.KernelLogger.Debug("Eliminando procesos en cola de espera...")
	for _, pcb := range cola {
		global.IniciarMetrica("BLOCKED", "EXIT", &pcb)
		global.KernelLogger.Debug(fmt.Sprintf("EXIT PID %d", pcb.PID))
	}
	delete(structs.IOsRegistrados, ioDesconectado.Nombre)
	delete(structs.ColaBlockedIO, ioDesconectado.Nombre)
	global.KernelLogger.Debug(fmt.Sprintf("todos los procesos esperando %s se fueron a EXIT", ioDesconectado.Nombre))
}