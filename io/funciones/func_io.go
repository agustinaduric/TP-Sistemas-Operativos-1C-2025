package fio

import (
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
	"encoding/json"
	"os"
	"fmt"
	"net/http"
	"bytes"
)

func IniciarConfiguracionIO(filePath string) config.IOConfig {
	var config config.IOConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
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
		log.Printf("error codificando ioARegistrar: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/registrar-io", config.IpKernel, config.PortKernel)
	http.Post(url, "application/json", bytes.NewBuffer(body))
	}
