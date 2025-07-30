package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/PCB"
	fkernel "github.com/sisoputnfrba/tp-golang/kernel/funciones"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func main() {

	comunicacion.VerificarParametros(3)
	configPath := os.Args[3]
	global.ConfigCargadito = fkernel.IniciarConfiguracionKernel(configPath)
	global.KernelLogger = fkernel.ConfigurarLog()
	go fkernel.LevantarServidorKernel(global.ConfigCargadito)
	handshake := structs.Handshake{IP: global.ConfigCargadito.IpKernel, Puerto: global.ConfigCargadito.PortKernel}
	comunicacion.EnviarHandshake(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, handshake)

	go planificacion.Iniciar_planificacion(global.ConfigCargadito)
	PrimerProceso()
	global.KernelLogger.Debug(fmt.Sprintf("se creo el primer proceso"))
	global.WgKernel.Wait()
	global.KernelLogger.Close()
}

func PrimerProceso() {
	var PATH string = os.Args[1]
	var tamanioSTR string = os.Args[2]
	var Tamanio int
	Tamanio, _ = strconv.Atoi(tamanioSTR)
	proceso := PCB.Crear(PATH, Tamanio)
	global.IniciarMetrica("", "NEW", &proceso)
}
