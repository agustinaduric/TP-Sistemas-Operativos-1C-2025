package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Enviar_proceso_a_memoria(pcb_a_cargar structs.PCB) string {
	global.KernelLogger.Debug(fmt.Sprintf("Entre a la funcion Enviar_proceso_a_memoria"))
	// consulto si hay espacio:
	espacioMemoria := fmt.Sprintf("http://%s:%d/espacio-libre", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	respEspacio, errEspacio := http.Get(espacioMemoria) // el get es para pedir info
	if errEspacio != nil {
		global.KernelLogger.Debug(fmt.Sprintf("Error al consultar espacio libre en memoria: %s", errEspacio.Error()))
		return ""
	}
	defer respEspacio.Body.Close()
	// me responde y verifico:
	var espacioLibre structs.EspacioLibreRespuesta
	json.NewDecoder(respEspacio.Body).Decode(&espacioLibre)
	if espacioLibre.BytesLibres < pcb_a_cargar.Tamanio {
		global.KernelLogger.Debug(fmt.Sprintf("no hay espacio para cargar al proceso PID:%d", pcb_a_cargar.PID))
		return ""
	}
	// se manda el proceso si hay espacio:
	var Proceso structs.Proceso_a_enviar = structs.Proceso_a_enviar{
		PID:     pcb_a_cargar.PID,
		Tamanio: pcb_a_cargar.Tamanio,
		PATH:    pcb_a_cargar.PATH,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error codificando el proceso: %s", err.Error()))
	}
	url := fmt.Sprintf("http://%s:%d/cargar-proceso", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, global.ConfigCargadito.PortMemory))
	}
	defer resp.Body.Close()
	var respuesta string
	json.NewDecoder(resp.Body).Decode(&respuesta)
	return respuesta
}

func Enviar_P_Finalizado_memoria(PID int) string {
	var Proceso int = PID
	body, err := json.Marshal(Proceso)
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error codificando el proceso: %s", err.Error()))
	}
	url := fmt.Sprintf("http://%s:%d/finalizarproc", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", PID, global.ConfigCargadito.PortMemory))
	}
	defer resp.Body.Close()
	var respuesta string
	json.NewDecoder(resp.Body).Decode(&respuesta)
	return respuesta
}

func MandarProcesoADesuspension(PID int) string {
	var Proceso int = PID
	body, err := json.Marshal(Proceso)
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error codificando el proceso: %s", err.Error()))
	}
	url := fmt.Sprintf("http://%s:%d/desuspension-proceso", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", PID, global.ConfigCargadito.PortMemory))
	}
	defer resp.Body.Close()
	var respuesta string
	json.NewDecoder(resp.Body).Decode(&respuesta)
	return respuesta
}
