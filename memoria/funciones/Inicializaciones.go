package fmemoria

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func IniciarConfiguracionMemoria(filePath string) {
	/*global.MemoriaLogger.Debug(
		fmt.Sprintf("Iniciando configuración de memoria desde: %s", filePath),
	)*/

	configFile, err := os.Open(filePath)
	if err != nil {
		/*global.MemoriaLogger.Error(
			fmt.Sprintf("Error al abrir config de memoria '%s': %s", filePath, err.Error()),
		)*/
		os.Exit(1)
	}

	defer configFile.Close()

	var config config.MemoriaConfig
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {

		/*global.MemoriaLogger.Error(
			fmt.Sprintf("Error al parsear config de memoria '%s': %s", filePath, err.Error()),
		)
*/
		os.Exit(1)
	}

	global.MemoriaConfig = config

	/*global.MemoriaLogger.Debug(
		fmt.Sprintf("Configuración de memoria cargada: MemorySize=%d, PageSize=%d",
			config.MemorySize, config.PageSize,
		),
	)*/
}

func IniciarMapaMemoriaDeUsuario() {
	global.MemoriaLogger.Debug("Inicializando mapa de memoria de usuario")

	totalMarcos := CantidadMarcos()
	global.MapMemoriaDeUsuario = make([]int, totalMarcos)
	for i := 0; i < totalMarcos; i++ {
		global.MapMemoriaDeUsuario[i] = -1
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Mapa de memoria de usuario inicializado con %d marcos libres", totalMarcos),
	)
}

func IniciarMemoriaUsuario() {
	global.MemoriaLogger.Debug("Reservando buffer de memoria de usuario")

	totalBytes := global.MemoriaConfig.MemorySize
	global.MemoriaUsuario = make([]byte, totalBytes)

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Buffer de memoria de usuario reservado: %d bytes", totalBytes),
	)
	
	IniciarMapaMemoriaDeUsuario()
}
