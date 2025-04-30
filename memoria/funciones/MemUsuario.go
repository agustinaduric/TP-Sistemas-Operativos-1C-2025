package fmemoria

import (
	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func CantidadMarcos() int {
	return global.MemoriaConfig.MemorySize / global.MemoriaConfig.PageSize
}

func MarcosDisponibles() int {
	var contador int = 0
	for i := 0; i < CantidadMarcos(); i++ {
		if global.MapMemoriaDeUsuario[i] == -1 {
			contador++
		}
	}
	return contador
}
func MarcosNecesitados(tamanioProceso int) int {
	if tamanioProceso%global.MemoriaConfig.PageSize == 0 {
		return tamanioProceso / global.MemoriaConfig.PageSize
	}
	return (tamanioProceso / global.MemoriaConfig.PageSize) + 1
}

// trankilo compilador idiota, la vamos a usar
func hayEspacio(tamanioProceso int) bool {
	if MarcosNecesitados(tamanioProceso) <= MarcosDisponibles() {
		return true
	}
	return false
}

func espacioDisponible() int {
	return MarcosDisponibles() * global.MemoriaConfig.PageSize
}
