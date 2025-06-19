package algoritmoCache

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Clock(entrada structs.EntradaCache) {
	global.CpuLogger.Debug("Entro a clock")
	for i:=0; i<len(global.CachePaginas); i++{
		entradaActual := &global.CachePaginas[global.PunteroClock]
		if !entradaActual.BitUso {
			if entradaActual.BitModificado{
				global.CpuLogger.Debug(fmt.Sprintf("Entrada modificada, se la mando a memoria antes de reemplazar"))
				comunicacion.EnviarEscribirAMemoria(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, entradaActual)
				global.CpuLogger.Debug(fmt.Sprintf("Se envio la entrada a memoria antes de reemplazar"))
			}
			global.CpuLogger.Debug(fmt.Sprintf("Reemplazo CLOCK en posicion: %d", global.PunteroClock))
			global.CachePaginas[global.PunteroClock] = entrada
			avanzarPuntero()
			return
		}
		entradaActual.BitUso = false
		global.CpuLogger.Debug("Bit de uso pasa a 0 y continuo")
		avanzarPuntero()
	}
}

func ClockM(entrada structs.EntradaCache) {
	global.CpuLogger.Debug("Entro a clock-m")
	// primera vuelta (U = 0 y M = 0)
	for i:=0; i<len(global.CachePaginas); i++{
		entradaActual := &global.CachePaginas[global.PunteroClock]
		if !entradaActual.BitUso && !entradaActual.BitModificado {
			if entradaActual.BitModificado{
				global.CpuLogger.Debug(fmt.Sprintf("Reemplazo, U= 0 y M=0"))
				global.CachePaginas[global.PunteroClock] = entrada
				avanzarPuntero()
			}
			avanzarPuntero()
			return
		}
		avanzarPuntero()
	}
	//segunda vuelta
	for  i:=0; i<len(global.CachePaginas); i++{
		entradaActual := &global.CachePaginas[global.PunteroClock]
		if !entradaActual.BitUso && entradaActual.BitModificado {
			global.CpuLogger.Debug(fmt.Sprintf("Entrada modificada U=0 M=1, se la mando a memoria antes de reemplazar"))
			comunicacion.EnviarEscribirAMemoria(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, entradaActual)
			global.CpuLogger.Debug(fmt.Sprintf("Se envio la entrada a memoria antes de reemplazar"))
			global.CachePaginas[global.PunteroClock] = entrada
			avanzarPuntero()
			return
		}
		entradaActual.BitUso= false
		global.CpuLogger.Debug("Bit de uso pasa a 0 y continuo")
		avanzarPuntero()
	}
}

func avanzarPuntero(){
	global.PunteroClock++
	if global.PunteroClock >= global.EntradasMaxCache {
		global.PunteroClock = 0
		global.CpuLogger.Debug("El puntero de clock volvio a cero")
	}
	global.CpuLogger.Debug("El puntero de clock avanzo")
}