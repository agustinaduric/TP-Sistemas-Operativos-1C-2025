package main

import (
	"os"

	fCpu "github.com/sisoputnfrba/tp-golang/cpu/funciones"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {

	comunicacion.VerificarParametros(2)
	nombre := os.Args[1]
	configPath := os.Args[2]
	global.ConfigCargadito = fCpu.IniciarConfiguracionCpu(configPath)
	go fCpu.LevantarServidorCPU()
	protocolos.Conectarse_con_Kernel(nombre)
	global.WgCPU.Wait()
	//comunicacion.EnviarMensaje(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, "Soy cpu, hola memoria")
	//comunicacion.EnviarMensaje(global.ConfigCargadito.IpKernel, global.ConfigCargadito.PortKernel, "Soy cpu,hola kernel")
}
