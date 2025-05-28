package main

import (
	"os"
	fio "github.com/sisoputnfrba/tp-golang/io/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {
	comunicacion.VerificarParametros(2)
	nombre := os.Args[1]
	configPath := os.Args[2]
	ConfigCargadito := fio.IniciarConfiguracionIO(configPath)
	fio.RegistrarEnKernel(nombre, ConfigCargadito)
	go fio.LevantarIO(ConfigCargadito)
	// comunicacion.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel, "SOy io, hola kernel")
}
