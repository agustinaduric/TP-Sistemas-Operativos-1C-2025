package comunicacion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func VerificarParametros(cantidad int) {
	if len(os.Args) < cantidad {
		fmt.Println("ERROR: No se ingresaro la cantidad de parametros necesarios")
		os.Exit(1) //si hay error en los parametros termino la ejecucion
	}
}

func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var mensaje structs.Mensaje
	err := decoder.Decode(&mensaje)
	if err != nil {
		log.Printf("Error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	log.Println("Me llego un mensaje de un cliente")
	log.Printf("%+v\n", mensaje)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func EnviarMensaje(ip string, puerto int, mensajeTxt string) {
	mensaje := structs.Mensaje{Mensaje: mensajeTxt}
	body, err := json.Marshal(mensaje)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/mensaje", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensaje a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func EnviarSolicitudIO(ip string, puerto int, soliUsoIO structs.Solicitud) {
	body, err := json.Marshal(soliUsoIO)
	if err != nil {
		log.Printf("error codificando solicitudIO: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/solicitud-io", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando solicitudIO:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s, EnviarSolicitudIO", resp.Status)
}

func EnviarFinIO(ip string, puerto int, respuestaIO structs.RespuestaIO) {
	body, err := json.Marshal(respuestaIO)
	if err != nil {
		log.Printf("error codificando la respuestaIO: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/finalizar-io", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando respuestaIO:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func EnviarHandshake(ip string, puerto int, PuertoIP structs.Handshake) {
	body, err := json.Marshal(PuertoIP)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/recibir-handshake", ip, puerto)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}

func EnviarDesconexion(ip string, puerto int, dispositivo structs.IODesconectado) {
	body, err := json.Marshal(dispositivo)
	if err != nil {
		log.Printf("error codificando la respuestaIO: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/desconexion-io", ip, puerto)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}
