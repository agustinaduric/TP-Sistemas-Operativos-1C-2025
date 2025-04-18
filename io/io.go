package main

import (
	"fmt"

	fio "github.com/sisoputnfrba/tp-golang/io/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {
	configCargadito := fio.IniciarConfiguracionIO("io/config/io.config.json")
	fmt.Println(configCargadito.Mensaje)

	comunicacion.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel, configCargadito.Mensaje)
}
