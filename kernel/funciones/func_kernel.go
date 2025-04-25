package fkernel

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
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

func LevantarServidorKernel(configCargadito config.KernelConfig) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje)

	puerto := config.IntToStringConPuntos(configCargadito.PortKernel)

	log.Printf("Servidor de Kernel escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}

func PlaniLargoFIFO (){
  //tiene que empezar frenado y esperar un ENTER
  while(1){
     //semaforo iniciar / finalizar procesos

    
  }
}
 
 
