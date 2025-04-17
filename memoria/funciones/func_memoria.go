package fmemoria

import (
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"net/http"
	"os"
	"encoding/json"
)

func IniciarConfiguracionMemoria(filePath string) config.MemoriaConfig {
	var config config.MemoriaConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func LevantarServidorMemoria(configCargadito config.MemoriaConfig){
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", config.RecibirMensaje)

	puerto := config.IntToStringConPuntos(configCargadito.PortMemory)

	log.Printf("Servidor de Memoria escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}
 