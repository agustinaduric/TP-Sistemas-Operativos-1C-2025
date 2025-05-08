package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Enviar_proceso_a_memoria(pcb_a_cargar structs.PCB, configCargadito config.KernelConfig) {
	var Proceso structs.Proceso_a_enviar = structs.Proceso_a_enviar{
		PID:     pcb_a_cargar.PID,
		Tamanio: pcb_a_cargar.Tamanio,
		PATH:    pcb_a_cargar.PATH,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/proceso", configCargadito.IpMemory, configCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, configCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
}

func Recibir_confirmacion(w http.ResponseWriter, r *http.Request) {
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

}

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
