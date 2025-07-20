package fio

import(
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	time.Sleep(time.Duration(solicitud.Duracion)*time.Millisecond)
	// log obligatorio 2/2
	globalIO.IOLogger.Info(fmt.Sprintf("PID: %d - Fin de IO", solicitud.PID))
	respuesta := structs.RespuestaIO{
		NombreIO: solicitud.NombreIO,
		PID: solicitud.PID,
	}
	comunicacion.EnviarFinIO(globalIO.IpKernel, globalIO.PuertoKernel, respuesta)
	globalIO.IOLogger.Debug(fmt.Sprintf("Se envio a kernel fin IO, Dispositivo: %s PID: %d", solicitud.NombreIO, solicitud.PID ))
}

func EsperarDesconexion(dispositivo string) {
	globalIO.IOLogger.Debug(fmt.Sprintf("%s ingreso a EsperarDesconexion", dispositivo))
	senial := make(chan os.Signal, 1)
	signal.Notify(senial, syscall.SIGINT, syscall.SIGTERM)
	<-senial
	globalIO.IOLogger.Debug(fmt.Sprintf("Dispositivo: %s se esta desconectando", dispositivo))
	ioDesconectado := structs.IODesconectado{Nombre: dispositivo}
	comunicacion.EnviarDesconexion(globalIO.IpKernel, globalIO.PuertoKernel,ioDesconectado)
	globalIO.IOLogger.Debug(fmt.Sprintf("AViso desconexion dispositivo: %s enviada a kernel", dispositivo))
}

func LevantarIO(configCargadito config.IOConfig) {
	mux := http.NewServeMux()
	mux.HandleFunc("/solicitud-io", RealizarIO)
	puerto := config.IntToStringConPuntos(configCargadito.PortIo)

	globalIO.IOLogger.Debug(fmt.Sprintf("IO escuchando solicitudes en %s", puerto))
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		globalIO.IOLogger.Error(fmt.Sprintf("Error al levantar IO: %v", err))
	}
}