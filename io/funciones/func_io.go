package fio

import (
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"encoding/json"
	"os"
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