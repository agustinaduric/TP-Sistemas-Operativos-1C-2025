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
	"github.com/sisoputnfrba/tp-golang/io/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func ConfigurarLog() *logger.LoggerStruct {
	logLevel, error1 := logger.ParseLevel(globalIO.IOConfig.LogLevel)
	if error1 != nil {
		fmt.Println("ERROR: El nivel de log ingresado no es valido")
		os.Exit(1)
	}
	logger, error2 := logger.NewLogger("io.log", logLevel)
	if error2 != nil {
		fmt.Println("ERROR: No se pudo crear el logger")
		os.Exit(1)
	}
	return logger
}

func IniciarConfiguracionIO(filePath string) config.IOConfig {
	var config config.IOConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	globalIO.IOConfig = config
	globalIO.IpKernel = config.IpKernel
	globalIO.PuertoKernel = config.PortKernel
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
		globalIO.IOLogger.Error(fmt.Sprintf("error codificando ioARegistrar: %s", err.Error()))
	}
	url := fmt.Sprintf("http://%s:%d/registrar-io", config.IpKernel, config.PortKernel)
	http.Post(url, "application/json", bytes.NewBuffer(body))
	globalIO.IOLogger.Debug(fmt.Sprintf("IO envio al kernel el dispositivo: %s", nombre))
	}

func RealizarIO(w http.ResponseWriter, r *http.Request){
	var solicitud structs.Solicitud
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&solicitud)
	if err != nil {
		globalIO.IOLogger.Error(fmt.Sprintf("Error al decodificar la solicitud: %s\n", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar la solicitud"))
		return
	}
	// log obligatorio 1/2
	globalIO.IOLogger.Info(fmt.Sprintf("PID: %d - Inicio de IO - Tiempo %d", solicitud.PID, solicitud.Duracion))
	time.Sleep(time.Duration(solicitud.Duracion)*time.Second)
	// log obligatorio 2/2
	globalIO.IOLogger.Info(fmt.Sprintf("PID: %d - Fin de IO", solicitud.PID))
	respuesta := structs.RespuestaIO{
		NombreIO: solicitud.NombreIO,
		PID: solicitud.PID,
		Desconexion: false,
	}
	comunicacion.EnviarFinIO(globalIO.IpKernel, globalIO.PuertoKernel, respuesta)
	globalIO.IOLogger.Debug(fmt.Sprintf("Se envio a kernel fin IO, Dispositivo: %s PID: %d", solicitud.NombreIO, solicitud.PID ))
}

func LevantarIO(configCargadito config.IOConfig) {
	mux := http.NewServeMux()
	mux.HandleFunc("/solicitar-io", RealizarIO)
	puerto := config.IntToStringConPuntos(configCargadito.PortIo)

	globalIO.IOLogger.Debug(fmt.Sprintf("IO escuchando solicitudes en %s", puerto))
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		globalIO.IOLogger.Error(fmt.Sprintf("Error al levantar IO: %v", err))
	}
}