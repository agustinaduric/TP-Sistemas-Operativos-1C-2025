package main

import (
	"fmt"

	fCpu "github.com/sisoputnfrba/tp-golang/cpu/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {

	configCargadito := fCpu.IniciarConfiguracionCpu("cpu/config/cpu.config.json")
	fmt.Println("Soy cpu")

	comunicacion.EnviarMensaje(configCargadito.IpMemory, configCargadito.PortMemory, "Soy cpu, hola memoria")

	comunicacion.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel, "Soy cpu,hola kernel")
}
