package fkernel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

func LevantarServidorKernel(configCargadito config.KernelConfig) {
	global.WgKernel.Add(1)
	defer global.WgKernel.Done()
	structs.IOsRegistrados = make(map[string]*structs.DispositivoIO)
	structs.ColaBlockedIO = make(map[string]structs.ColaProcesos)
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje)
	mux.HandleFunc("/devolucion", protocolos.Recibir_devolucion_CPU)
	mux.HandleFunc("/registrar-io", protocolos.HandlerRegistrarIO)
	mux.HandleFunc("/finalizar-io", protocolos.HandlerFinalizarIO)
	mux.HandleFunc("/conectarcpu", protocolos.Conectarse_con_CPU)
	mux.HandleFunc("/desconexion-io", protocolos.HandlerDesconexionIO)

	puerto := config.IntToStringConPuntos(configCargadito.PortKernel)

	global.KernelLogger.Debug(fmt.Sprintf("Servidor de Kernel escuchando en %s", puerto))
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		global.KernelLogger.Error(fmt.Sprintf("Error al levantar el servidor: %v", err))
	}
}
