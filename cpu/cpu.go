package main

import (
	"fmt"

	fCpu "github.com/sisoputnfrba/tp-golang/cpu/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {

	configCargadito := fCpu.IniciarConfiguracionCpu("cpu/config/cpu.config.json")
	fmt.Println(configCargadito.Mensaje)

	comunicacion.EnviarMensaje(configCargadito.IpMemory, configCargadito.PortMemory, configCargadito.Mensaje)

	comunicacion.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel, configCargadito.Mensaje)
}
