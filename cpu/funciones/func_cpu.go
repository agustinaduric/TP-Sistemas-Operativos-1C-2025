package fCpu

import (
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"encoding/json"
	"os"
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

