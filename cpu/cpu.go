package main

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)


func main() {
   
   configCargadito := fCpu.IniciarConfiguracionCpu("cpu/config/cpu.config.json")
   fmt.Println(configCargadito.Mensaje)
	
   config.EnviarMensaje(configCargadito.IpMemory, configCargadito.PortMemory,configCargadito.Mensaje )
   
   config.EnviarMensaje(configCargadito.IpKernel, configCargadito.PortKernel,configCargadito.Mensaje )
}

