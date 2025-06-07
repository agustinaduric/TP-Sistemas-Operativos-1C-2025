package main

import (
	"os"
	fio "github.com/sisoputnfrba/tp-golang/io/funciones"
	"github.com/sisoputnfrba/tp-golang/io/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {
	comunicacion.VerificarParametros(2)
	nombre := os.Args[1]
	configPath := os.Args[2]
	ConfigCargadito := fio.IniciarConfiguracionIO(configPath)
	globalIO.IOLogger = fio.ConfigurarLog()
	fio.RegistrarEnKernel(nombre, ConfigCargadito)
	go fio.EsperarDesconexion(nombre)
	fio.LevantarIO(ConfigCargadito)
}