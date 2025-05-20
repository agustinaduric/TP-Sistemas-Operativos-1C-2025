package fkernel

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func definirLogLevel(config config.KernelConfig) {
	var level string = config.LogLevel
	switch level {
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		slog.SetLogLoggerLevel(slog.LevelInfo) // nivel por defecto
	}

	//hace que todos los logs se impriman
	log.SetFlags(log.Lmicroseconds | log.Lshortfile) // lo probamos y si queda muy gede lo comentamos y fue
}

func IniciarConfiguracionKernel(filePath string) config.KernelConfig {
	var config config.KernelConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	definirLogLevel(config)

	return config
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
	// registro el IO
	structs.IOsRegistrados[registro.Nombre] = &nuevoIO
	log.Printf("se registro el io: %s", registro.Nombre) // borrar desp
}

func HandlerFinalizarIO(w http.ResponseWriter, r *http.Request) {
	var respuestaFin structs.RespuestaIO
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&respuestaFin)
	if err != nil {
		http.Error(w, "Error en decodificar la respuesta de io: "+err.Error(), http.StatusBadRequest)
		return
	}
	dispositivo := structs.IOsRegistrados[respuestaFin.NombreIO]
	dispositivo.PIDActual = 0
	if respuestaFin.Desconexion {
		cola := structs.ColaBlockedIO[respuestaFin.NombreIO]
		proceso := PCB.Buscar_por_pid(respuestaFin.PID, &cola)
		structs.ColaBlockedIO[respuestaFin.NombreIO] = cola
		proceso.Estado = structs.EXIT
		global.Push_estado(&structs.ColaExit, proceso)
		return
	}
	if len(structs.ColaBlockedIO[respuestaFin.NombreIO]) > 0 {
		//saco el primero:
		siguiente := structs.ColaBlockedIO[respuestaFin.NombreIO][0] //
		structs.ColaBlockedIO[respuestaFin.NombreIO] = structs.ColaBlockedIO[respuestaFin.NombreIO][1:]
		dispositivo.PIDActual = siguiente.PID
		SolicitudParaIO := structs.Solicitud{PID: siguiente.PID, NombreIO: dispositivo.Nombre, Duracion: siguiente.IOPendienteDuracion}
		comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
	}
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
	mux.HandleFunc("/confirma-finalizado", protocolos.Recibir_confirmacion)
	mux.HandleFunc("/confirm-dumpmemory", protocolos.Recibir_confirmacion_DumpMemory)
	mux.HandleFunc("/conectarcpu", protocolos.Conectarse_con_CPU)

	puerto := config.IntToStringConPuntos(configCargadito.PortKernel)

	log.Printf("Servidor de Kernel escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}
