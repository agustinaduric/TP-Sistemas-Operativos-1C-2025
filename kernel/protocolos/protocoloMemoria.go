package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Enviar_proceso_a_memoria(pcb_a_cargar structs.PCB, configCargadito config.KernelConfig) int {
	// consulto si hay espacio:
	espacioMemoria := fmt.Sprintf("http:/%s:%d/espacio-libre", configCargadito.IpMemory, configCargadito.PortMemory)
	respEspacio, errEspacio := http.Get(espacioMemoria) // el get es para pedir info
	if errEspacio != nil {
		log.Printf("Error al consultar espacio libre en memoria: %s", errEspacio.Error())
		return 0
	}
	defer respEspacio.Body.Close()
	// me responde y verifico:
	var espacioLibre structs.EspacioLibreRespuesta
	json.NewDecoder(respEspacio.Body).Decode(&espacioLibre)
	if espacioLibre.BytesLibres < pcb_a_cargar.Tamanio {
		log.Printf("no hay espacio para cargar al proceso PID:%d", pcb_a_cargar.PID)
		return 0
	}
	// se manda el proceso si hay espacio:
	var Proceso structs.Proceso_a_enviar = structs.Proceso_a_enviar{
		PID:     pcb_a_cargar.PID,
		Tamanio: pcb_a_cargar.Tamanio,
		PATH:    pcb_a_cargar.PATH,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/cargar-proceso", configCargadito.IpMemory, configCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, configCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	return resp.StatusCode
}

//LA FUNCION DE ARRIBA YA SOLUCIONA Y NO NECESITAMOS ESTA
/*func Recibir_confirmacion(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion string
	err := decoder.Decode(&Devolucion)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		<-global.SemFinalizacion
		return
	}
	if Devolucion == "OK" {
		global.ConfirmacionProcesoCargado = 1
		<-global.SemFinalizacion
		return

	}
	<-global.SemFinalizacion
	return

}*/

func Enviar_P_Finalizado_memoria(PID int) {
	var Proceso int = PID
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/finalizarproc", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", PID, global.ConfigCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
}

func Recibir_confirmacionFinalizado(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion string
	err := decoder.Decode(&Devolucion)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		<-global.SemFinalizacion
		return
	}
	if Devolucion == "OK" {
		global.ConfirmacionProcesoFinalizado = 1
		<-global.SemFinalizacion
		return
	}
	<-global.SemFinalizacion
	return

}

func Recibir_confirmacion_DumpMemory(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var Devolucion_DumpMemory structs.Devolucion_DumpMemory
	err := decoder.Decode(&Devolucion_DumpMemory)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}
	var Proceso structs.PCB
	Proceso = PCB.Buscar_por_pid(Devolucion_DumpMemory.PID, &structs.ColaBlocked)
	log.Printf("me llego una Devolucion de Memoria")
	switch Devolucion_DumpMemory.Respuesta {
	case "CONFIRMACION":
		global.MutexREADY.Lock()
		global.Push_estado(&structs.ColaReady, Proceso)
		global.MutexREADY.Unlock()
	case "ERROR":
		global.MutexEXIT.Lock()
		global.Push_estado(&structs.ColaExit, Proceso)
		global.MutexREADY.Unlock()
		<-global.ProcesoParaFinalizar

	}
}
