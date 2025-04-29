package main

import (
	fkernel "github.com/sisoputnfrba/tp-golang/kernel/funciones"
	"github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {

	configCargadito := fkernel.IniciarConfiguracionKernel("kernel/config/kernel.config.json")

	comunicacion.EnviarMensaje(configCargadito.IpMemory, configCargadito.PortMemory, "Soy kernel,hola memoria")
	fkernel.LevantarServidorKernel(configCargadito)
	planificacion.Iniciar_planificacion(configCargadito)
}
