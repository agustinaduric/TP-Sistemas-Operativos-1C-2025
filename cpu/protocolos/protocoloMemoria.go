package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func SolicitarInstruccion() {
	var Proceso structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: global.Proceso_Ejecutando.PID,
		PC:  global.Proceso_Ejecutando.PC,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/obtener-instruccion", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", global.Proceso_Ejecutando.PID, global.ConfigCargadito.PortMemory)
		return
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	defer resp.Body.Close()

	var instruccion structs.Instruccion
	decoder := json.NewDecoder(resp.Body)
	errRecepcion := decoder.Decode(&instruccion)
	if errRecepcion != nil {
		log.Printf("error al decodificar instruccion: %s\n", errRecepcion.Error())
		os.Exit(1)
	}
	global.Instruccion = instruccion.Operacion + " " + strings.Join(instruccion.Argumentos, " ")
}

func Conectarse_con_Memoria(identificador string) {
	var Proceso structs.CPU_a_memoria = structs.CPU_a_memoria{
		IP:            global.ConfigCargadito.IpCPu,
		Puerto:        global.ConfigCargadito.PortCpu,
		Identificador: identificador,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/conectarcpumemoria", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando CPU:%s a puerto:%d", os.Args[1], global.ConfigCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	errRecepcion := decoder.Decode(&global.Datos_Memoria)
	if errRecepcion != nil {
		log.Printf("error al decodificar instruccion: %s\n", errRecepcion.Error())
		os.Exit(1)
	}

}