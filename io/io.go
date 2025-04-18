package main

import (
	"fmt"

	fio "github.com/sisoputnfrba/tp-golang/io/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {
	configCargadito := fio.IniciarConfiguracionIO("io/config/io.config.json")
	fmt.Println("Soy io desde un println")

	comunicacion.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel, "SOy io, hola kernel")
}
