package main

import (

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/kernel/funciones"
)

func main() {
	
	configCargadito := fkernel.IniciarConfiguracionKernel("kernel/config/kernel.config.json")
	
	config.EnviarMensaje(configCargadito.IpMemory, configCargadito.PortMemory,configCargadito.Mensaje)
	fkernel.LevantarServidorKernel(configCargadito)
}
 