package fCpu

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/ciclo"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/config"
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
	mux.HandleFunc("/interrupcion", Recibir_Proceso_Kernel)
	mux.HandleFunc("/Reconectar", protocolos.Reconectar_Proceso)

	puerto := config.IntToStringConPuntos(global.ConfigCargadito.PortCpu)

	log.Printf("Servidor de Kernel escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}

func Recibir_Proceso_Kernel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&global.Proceso_Ejecutando)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}
	ciclo.Ciclo()
	return
}
