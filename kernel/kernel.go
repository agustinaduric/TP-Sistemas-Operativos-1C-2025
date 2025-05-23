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

	chequeoParametros()

	global.ConfigCargadito = fkernel.IniciarConfiguracionKernel("kernel/config/kernel.config.json")
	go fkernel.LevantarServidorKernel(global.ConfigCargadito)
	comunicacion.EnviarMensaje(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, "Soy kernel,hola memoria")
	handshake := structs.Handshake{IP: global.ConfigCargadito.IpKernel, Puerto: global.ConfigCargadito.PortKernel}
	comunicacion.EnviarHandshake(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, handshake)

	planificacion.Iniciar_planificacion(global.ConfigCargadito)
	PrimerProceso()
	global.WgKernel.Wait()
}

func chequeoParametros() {
	if len(os.Args) < 3 {
		fmt.Println("ERROR: No se ingresaro la cantidad de parametros necesarios")
		os.Exit(1) //si hay error en los parametros termino la ejecucion
	}
}

func PrimerProceso() {
	var PATH string = os.Args[1]
	var tamanioSTR string = os.Args[2]
	var Tamanio int
	Tamanio, _ = strconv.Atoi(tamanioSTR)
	proceso := PCB.Crear(PATH, Tamanio)
	global.MutexNEW.Lock()
	global.Push_estado(&structs.ColaNew, proceso)
	global.MutexNEW.Unlock()

}
