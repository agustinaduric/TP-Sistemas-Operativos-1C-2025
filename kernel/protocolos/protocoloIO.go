package protocolos

import(
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/syscalls"
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
		PIDActual:    -1,
		ColaEsperaIO: []*structs.PCB{},
	}
	structs.IOsRegistrados[registro.Nombre] = append(structs.IOsRegistrados[registro.Nombre], &nuevoIO)
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
	global.KernelLogger.Info(fmt.Sprintf("## %d finalizó IO y pasa a READY", respuestaFin.PID))
    proceso, existe := PCB.Buscar_por_pid(respuestaFin.PID, &structs.ColaBlocked)
	if !existe {
		proceso, existe = PCB.Buscar_por_pid(respuestaFin.PID, &structs.ColaSuspBlocked)
		if existe {
			global.IniciarMetrica("BLOCKED", "READY", &proceso)
		} else {
			global.KernelLogger.Error(fmt.Sprintf("No existe PID %d en bloqueados ni en bloqueadosSUSP", respuestaFin.PID))
			return
		}
	}else {global.IniciarMetrica("BLOCKED", "READY", &proceso)}
	

	dispositivos := structs.IOsRegistrados[respuestaFin.NombreIO]

	for _,dispositivoIO:= range dispositivos{
		if dispositivoIO.PIDActual == respuestaFin.PID{
			dispositivoIO.PIDActual = -1
			global.KernelLogger.Debug(fmt.Sprintf("El dispositivo: %s que ocupo PID: %d esta libre", respuestaFin.NombreIO, respuestaFin.PID))
			
			if len(structs.ColaBlockedIO[respuestaFin.NombreIO]) > 0 {
				siguiente := Buscar_Siguiente_IO(respuestaFin.NombreIO)
				if siguiente.PID == -1 { return }
				dispositivo := syscalls.BuscarIOLibre(respuestaFin.NombreIO)
				dispositivo.PIDActual = siguiente.PID
				global.KernelLogger.Debug(fmt.Sprintf("PID: %d ocupo el dispositivo: %s", siguiente.PID, respuestaFin.NombreIO))
				SolicitudParaIO := structs.Solicitud{PID: siguiente.PID, NombreIO: dispositivo.Nombre, Duracion: siguiente.IOPendienteDuracion}
				global.MutexBLOCKED.Lock()
				structs.ColaBlockedIO[respuestaFin.NombreIO] = structs.ColaBlockedIO[respuestaFin.NombreIO][1:]
				global.MutexBLOCKED.Unlock()
				comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
				global.KernelLogger.Debug(fmt.Sprintf("Solcitud IO: %s enviada, PID: %d", respuestaFin.NombreIO, siguiente.PID))
			} else {
				global.KernelLogger.Debug(fmt.Sprintf("NO hay procesos en espera para: %s", respuestaFin.NombreIO))
			}
			break
		}
	}

}

func HandlerDesconexionIO(w http.ResponseWriter, r *http.Request){
	var ioDesconectado structs.IODesconectado
	err := json.NewDecoder(r.Body).Decode(&ioDesconectado)
	if err != nil {
		http.Error(w, "Error en decodificar la desconexion de io: "+err.Error(), http.StatusBadRequest)
		return
	}
	global.KernelLogger.Debug(fmt.Sprintf("Recibi la desconexion de: %s", ioDesconectado.Nombre))

	dispositivos := structs.IOsRegistrados[ioDesconectado.Nombre]
	instancias := make([]*structs.DispositivoIO, 0, len(dispositivos))

	for _, dispositivo := range dispositivos {
        if dispositivo.IP == ioDesconectado.IP && dispositivo.Puerto == ioDesconectado.Puerto {
			if dispositivo.PIDActual !=-1 { // si hay uno ejecutando, lo mato
			proceso, _ := PCB.Buscar_por_pid(dispositivo.PIDActual, &structs.ColaBlocked)
			global.IniciarMetrica("BLOCKED", "EXIT", &proceso)
			global.KernelLogger.Debug(fmt.Sprintf("EXIT PID: %d por estar ejecutando en io desconectado", dispositivo.PIDActual))
			}
		}else {
            instancias = append(instancias, dispositivo)
        }
	}

	if len(instancias) == 0{
		cola := structs.ColaBlockedIO[ioDesconectado.Nombre]
			global.KernelLogger.Debug("Eliminando procesos en cola de espera...")
			delete(structs.IOsRegistrados, ioDesconectado.Nombre)
			delete(structs.ColaBlockedIO, ioDesconectado.Nombre)
			for _, pcb := range cola {
				global.IniciarMetrica("BLOCKED", "EXIT", &pcb)
				global.KernelLogger.Debug(fmt.Sprintf("EXIT PID %d", pcb.PID))
			}
		global.KernelLogger.Debug(fmt.Sprintf("todos los procesos esperando %s se fueron a EXIT", ioDesconectado.Nombre))
	} else {
        structs.IOsRegistrados[ioDesconectado.Nombre] = instancias
        global.KernelLogger.Debug(fmt.Sprintf("Instancias restantes de %s: %d", ioDesconectado.Nombre, len(instancias)))
    }
} 

func Buscar_Siguiente_IO(NombreIO string) structs.PCB {
	for len(structs.ColaBlockedIO[NombreIO]) > 0 {
		if structs.ColaBlockedIO[NombreIO][0].Estado == "BLOCKED" {
			siguiente := structs.ColaBlockedIO[NombreIO][0]
			return siguiente
		} else {
			global.MutexBLOCKED.Lock()
			structs.ColaBlockedIO[NombreIO] = structs.ColaBlockedIO[NombreIO][1:]
			global.MutexBLOCKED.Unlock()
		}
	}
	global.KernelLogger.Debug(fmt.Sprintf("NO hay procesos en espera para: %s", NombreIO))
	return structs.PCB{PID: -1}
}