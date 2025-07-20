package fCpu

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/ciclo"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func IniciarConfiguracionCpu(filePath string) config.CPUConfig {
	var config config.CPUConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func LevantarServidorCPU() {
	global.WgCPU.Add(1)
	defer global.WgCPU.Done()
	mux := http.NewServeMux()
	mux.HandleFunc("/datoCPU", Recibir_Proceso_Kernel)
	mux.HandleFunc("/interrupcion", protocolos.Ocurrio_Interrupcion)
	mux.HandleFunc("/reconectar", protocolos.Reconectar_Proceso)

	puerto := config.IntToStringConPuntos(global.ConfigCargadito.PortCpu)

	log.Printf("Servidor de Kernel escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}

func Recibir_Proceso_Kernel(w http.ResponseWriter, r *http.Request) {
	global.CpuLogger.Debug("Recibi un proceso de kernel")
	var proceso structs.PIDyPC_Enviar_CPU
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&proceso)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}
	procesoejec := structs.PIDyPC_Enviar_CPU {
		PID: proceso.PID,
		PC: proceso.PC,
	}
	global.Proceso_Ejecutando = procesoejec
	resp := "OK"
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.CpuLogger.Error(
			fmt.Sprintf("Error codificando respuesta de recibir proceso: %s", err.Error()),
		)
	} else {
		global.CpuLogger.Debug(
			fmt.Sprintf("Confirmacion enviada"),
		)
	}
	go ciclo.Ciclo()
	return
}

func ConfigurarLog() *logger.LoggerStruct {
	logLevel, error1 := logger.ParseLevel(global.ConfigCargadito.LogLevel)
	if error1 != nil {
		fmt.Println("ERROR: El nivel de log ingresado no es valido")
		os.Exit(1)
	}
	logger, error2 := logger.NewLogger(global.Nombre, logLevel)
	if error2 != nil {
		fmt.Println("ERROR: No se pudo crear el logger")
		os.Exit(1)
	}
	return logger
}
