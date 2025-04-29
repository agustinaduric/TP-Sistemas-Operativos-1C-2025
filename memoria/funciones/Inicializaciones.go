package fmemoria

import (
	"encoding/json"
	"log"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func IniciarConfiguracionMemoria(filePath string) {
	var config config.MemoriaConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
}

func IniciarMapaMemoriaDeUsuario() {
	global.MapMemoriaDeUsuario = make([]int, global.MemoriaConfig.MemorySize/global.MemoriaConfig.PageSize)

	for i := 0; i < CantidadMarcos(); i++ {
		global.MapMemoriaDeUsuario[i] = -1
	}
}

func IniciarMemoriaUsuario() {
	global.MemoriaUsuario = make([]byte, global.MemoriaConfig.MemorySize)
	IniciarMapaMemoriaDeUsuario()
}
