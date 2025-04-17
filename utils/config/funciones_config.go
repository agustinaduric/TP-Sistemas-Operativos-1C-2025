package config

import(
	"encoding/json"
	"os"
	"fmt"
	"log"
	"net/http"
	"bytes"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)
// ** CargarConfig**
// recibe el path de .json segun el modulo y un puntero al struct para cargarle los datos
func CargarConfig [T any] (path string, cfg *T) error{ //Comentario de nardo: para q pinga esta ese cfg 
	file, err := os.Open(path)
	if err != nil{
		log.Printf("error")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(file)
	//jsonParser.Decode(&config)
	//jsonParser := json.NewDecoder(configFile)
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

func IntToStringConPuntos(valor int) string{
	return fmt.Sprintf(":%d", valor) 
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
