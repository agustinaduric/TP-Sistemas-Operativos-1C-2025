package fkernel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func IniciarConfiguracionKernel(filePath string) config.KernelConfig {
	var config config.KernelConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func ConfigurarLog() *logger.LoggerStruct {
	logLevel, error1 := logger.ParseLevel(global.ConfigCargadito.LogLevel)
	if error1 != nil {
		fmt.Println("ERROR: El nivel de log ingresado no es valido")
		os.Exit(1)
	}
	logger, error2 := logger.NewLogger("kernel.log", logLevel)
	if error2 != nil {
		fmt.Println("ERROR: No se pudo crear el logger")
		os.Exit(1)
	}
	return logger
}

func HandlerRegistrarIO(w http.ResponseWriter, r *http.Request) {
	var registro structs.RegistroIO
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&registro)
	if err != nil {
		http.Error(w, "Error en decodificar mje: "+err.Error(), http.StatusBadRequest)
		return
	}
	nuevoIO := structs.DispositivoIO{
		Nombre:       registro.Nombre,
		IP:           registro.IP,
		Puerto:       registro.Puerto,
		PIDActual:    0,
		ColaEsperaIO: []*structs.PCB{},
	}
	structs.IOsRegistrados[registro.Nombre] = &nuevoIO
	global.KernelLogger.Debug(fmt.Sprintf("se registro el io: %s en kernel", registro.Nombre))
}

func HandlerFinalizarIO(w http.ResponseWriter, r *http.Request) {
	var respuestaFin structs.RespuestaIO
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&respuestaFin)
	if err != nil {
		http.Error(w, "Error en decodificar la respuesta de io: "+err.Error(), http.StatusBadRequest)
		return
	}
	cola := structs.ColaBlockedIO[respuestaFin.NombreIO]
	proceso := PCB.Buscar_por_pid(respuestaFin.PID, &cola) // esto es una copia(?
	structs.ColaBlockedIO[respuestaFin.NombreIO] = cola
	if respuestaFin.Desconexion {
		global.KernelLogger.Debug(fmt.Sprintf("Se desconecto io: %s", respuestaFin.NombreIO))
		global.IniciarMetrica("BLOCKED", "EXIT", &proceso)
		return
	}
	global.KernelLogger.Debug(fmt.Sprintf("NO se desconecto io: %s", respuestaFin.NombreIO))
	global.IniciarMetrica("BLOCKED", "READY", &proceso)

	dispositivo := structs.IOsRegistrados[respuestaFin.NombreIO]
	dispositivo.PIDActual = 0
	global.KernelLogger.Debug(fmt.Sprintf("El dispositivo %s esta libre", respuestaFin.NombreIO))
	if len(structs.ColaBlockedIO[respuestaFin.NombreIO]) > 0 {
		siguiente := structs.ColaBlockedIO[respuestaFin.NombreIO][0]
		structs.ColaBlockedIO[respuestaFin.NombreIO] = structs.ColaBlockedIO[respuestaFin.NombreIO][1:]
		dispositivo.PIDActual = siguiente.PID
		global.KernelLogger.Debug(fmt.Sprintf("PID: %d ocupo el dispositivo: %s", siguiente.PID, respuestaFin.NombreIO))
		SolicitudParaIO := structs.Solicitud{PID: siguiente.PID, NombreIO: dispositivo.Nombre, Duracion: siguiente.IOPendienteDuracion}
		comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
		global.KernelLogger.Debug(fmt.Sprintf("Solcitud IO: %s enviada, PID: %d", respuestaFin.NombreIO, siguiente.PID))
	}
	global.KernelLogger.Debug(fmt.Sprintf("NO hay procesos en espera para: %s", respuestaFin.NombreIO))
}

func LevantarServidorKernel(configCargadito config.KernelConfig) {
	global.WgKernel.Add(1)
	defer global.WgKernel.Done()
	structs.IOsRegistrados = make(map[string]*structs.DispositivoIO)
	structs.ColaBlockedIO = make(map[string]structs.ColaProcesos)
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje)
	mux.HandleFunc("/devolucion", protocolos.Recibir_devolucion_CPU)
	mux.HandleFunc("/registrar-io", HandlerRegistrarIO)
	mux.HandleFunc("/finalizar-io", HandlerFinalizarIO)
	mux.HandleFunc("/confirmacion", protocolos.Recibir_confirmacionFinalizado)
	//mux.HandleFunc("/confirma-finalizado", protocolos.Recibir_confirmacion)
	mux.HandleFunc("/confirm-dumpmemory", protocolos.Recibir_confirmacion_DumpMemory)
	mux.HandleFunc("/conectarcpu", protocolos.Conectarse_con_CPU)

	puerto := config.IntToStringConPuntos(configCargadito.PortKernel)

	global.KernelLogger.Debug(fmt.Sprintf("Servidor de Kernel escuchando en %s", puerto))
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("Error al levantar el servidor: %v", err))
	}
}
