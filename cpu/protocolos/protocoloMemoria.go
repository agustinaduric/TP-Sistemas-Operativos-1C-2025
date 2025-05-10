package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Solicitar_instruccion() {
	var Proceso structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: global.Proceso_Ejecutando.PID,
		PC:  global.Proceso_Ejecutando.PC,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/solictarInstruccion", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", global.Proceso_Ejecutando.PID, global.ConfigCargadito.PortMemory)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
}

func Recibir_instruccion(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&global.Instruccion)
	if err != nil {
		log.Printf("error al decodificar instruccion: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar instruccion"))
		os.Exit(1)
	}
	<-global.InstruccionRecibida
}
