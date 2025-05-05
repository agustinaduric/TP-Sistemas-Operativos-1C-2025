package fio

import(
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// *** Globales ***
var ipKernel string
var puertoKernel int

func IniciarConfiguracionIO(filePath string) config.IOConfig {
	var config config.IOConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	ipKernel = config.IpKernel
	puertoKernel = config.PortKernel
	return config
}

func RegistrarEnKernel(nombre string, config config.IOConfig){
	ioARegistrar := structs.RegistroIO{
		Nombre: nombre,
		IP: config.IpIo,
		Puerto: config.PortIo,
	}
	body, err := json.Marshal(ioARegistrar)
	if err != nil {
		log.Printf("error codificando ioARegistrar: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/registrar-io", config.IpKernel, config.PortKernel)
	http.Post(url, "application/json", bytes.NewBuffer(body))
	}

func RealizarIO(w http.ResponseWriter, r *http.Request){
	var solicitud structs.Solicitud
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&solicitud)
	if err != nil {
		log.Printf("Error al decodificar la solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar la solicitud"))
		return
	}
	// log obligatorio 1/2
	log.Printf("## PID: %d - Inicio de IO - Tiempo %d", solicitud.PID, solicitud.Duracion)
	time.Sleep(time.Duration(solicitud.Duracion)*time.Second)
	//log obligatorio 2/2
	log.Printf("## PID: %d - Fin de IO", solicitud.PID)
	respuesta := structs.RespuestaIO{
		NombreIO: solicitud.NombreIO,
		PID: solicitud.PID,
	}
	comunicacion.EnviarFinIO(ipKernel,puertoKernel, respuesta)
}