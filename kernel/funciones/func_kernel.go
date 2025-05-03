package fkernel

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

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

// recibe IO y lo agrega a  IOsRegistrados
func HandlerRegistrarIO(w http.ResponseWriter, r *http.Request) {
	var registro structs.RegistroIO
	// me llego un json y lo decodifico para tener los datos del io
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&registro)
	// pregunto si hay error
	if err != nil {
		http.Error(w, "Error en decodificar mje: "+err.Error(), http.StatusBadRequest)
		return
	}
	nuevoIO := structs.DispositivoIO{
		Nombre: registro.Nombre,
		IP: registro.IP,
		Puerto: registro.Puerto,
		PIDActual: 0,
		ColaEsperaIO: []*structs.PCB{},
	}
	// registro el IO
	structs.IOsRegistrados[registro.Nombre] = &nuevoIO
}

// para el proceso que quiere usar la IO segun CPU:
func SolicitarSyscallIO(NuevaSolicitudIO structs.Solicitud) {
	// procedo a ver si existe la io
	_, hayMatch := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
	pcbSolicitante := structs.ProcesoEjecutando
	if hayMatch {
		dispositivo := structs.IOsRegistrados[NuevaSolicitudIO.NombreIO]
		pcbSolicitante.Estado = structs.BLOCKED // lo mando a blocked: por esperar o por estar usando la io
		pcbSolicitante.IOPendiente = dispositivo.Nombre
		if dispositivo.PIDActual != 0 { // ocupado
			structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO] = append(structs.ColaBlockedIO[NuevaSolicitudIO.NombreIO], pcbSolicitante) // agrego a cola de bloqueados por IO
		} else { // libre
			dispositivo.PIDActual = pcbSolicitante.PID // lo ocupo
			// cargo el struct y lo mando
			SolicitudParaIO := structs.Solicitud{PID: pcbSolicitante.PID, Duracion: NuevaSolicitudIO.Duracion, NombreIO: NuevaSolicitudIO.NombreIO}
			comunicacion.EnviarSolicitudIO(dispositivo.IP, dispositivo.Puerto, SolicitudParaIO)
		}
	} else {
		pcbSolicitante.IOPendiente = ""
		pcbSolicitante.Estado = structs.EXIT // no existe, se va a exit
		// cola exit para guardar los que terminaron ¿?
	}
}

func LevantarServidorKernel(configCargadito config.KernelConfig) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje) // ver NewServeMux usar en este check -> goroutines/hilos
	mux.HandleFunc("/devolucion", protocolos.Recibir_devolucion_CPU)
	puerto := config.IntToStringConPuntos(configCargadito.PortKernel)

	log.Printf("Servidor de Kernel escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}
