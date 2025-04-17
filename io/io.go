package main

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/io/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)


func main() {
	configCargadito := fio.IniciarConfiguracionIO("io/config/io.config.json")
	fmt.Println(configCargadito.Mensaje)
	 
	config.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel,configCargadito.Mensaje )
}