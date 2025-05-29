package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func CantidadMarcos() int {
	total := global.MemoriaConfig.MemorySize / global.MemoriaConfig.PageSize

	global.MemoriaLogger.Debug(
		fmt.Sprintf("CantidadMarcos: MemorySize=%d, PageSize=%d, TotalMarcos=%d",
			global.MemoriaConfig.MemorySize,
			global.MemoriaConfig.PageSize,
			total,
		),
	)
	return total
}

func MarcosDisponibles() int {

	global.MemoriaLogger.Debug("MarcosDisponibles: contando marcos libres")

	var contador int
	for i := 0; i < CantidadMarcos(); i++ {
		if global.MapMemoriaDeUsuario[i] == -1 {
			contador++
		}
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosDisponibles: TotalDisponibles=%d", contador),
	)
	return contador
}

func MarcosNecesitados(tamanioProceso int) int {

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosNecesitados: TamanioProceso=%d", tamanioProceso),
	)

	var necesarios int
	if rem := tamanioProceso % global.MemoriaConfig.PageSize; rem == 0 {
		necesarios = tamanioProceso / global.MemoriaConfig.PageSize
	} else {
		necesarios = (tamanioProceso / global.MemoriaConfig.PageSize) + 1
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosNecesitados: Necesarios=%d (para TamanioProceso=%d)", necesarios, tamanioProceso),
	)
	return necesarios
}

// trankilo, ya lo vamos a usar, compilador idiota
func hayEspacio(tamanioProceso int) bool {
	necesarios := MarcosNecesitados(tamanioProceso)
	disponibles := MarcosDisponibles()
	hay := necesarios <= disponibles

	global.MemoriaLogger.Debug(
		fmt.Sprintf("hayEspacio: Necesarios=%d, Disponibles=%d, HayEspacio=%t",
			necesarios, disponibles, hay,
		),
	)

	return hay
}

// espacioDisponible devuelve el total de bytes disponibles en memoria.
func espacioDisponible() int {
	disponibles := MarcosDisponibles() * global.MemoriaConfig.PageSize

	global.MemoriaLogger.Debug(
		fmt.Sprintf("espacioDisponible: MarcosDisponibles=%d, PageSize=%d, BytesDisponibles=%d",
			MarcosDisponibles(),
			global.MemoriaConfig.PageSize,
			disponibles,
		),
	)

	return disponibles
}
