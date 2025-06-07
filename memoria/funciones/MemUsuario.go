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

func leerMemoriaUsuario(direccionFisica int) string {
	global.MemoriaMutex.Lock()

	if direccionFisica < 0 || direccionFisica >= len(global.MemoriaUsuario) {
		global.MemoriaLogger.Error(
			fmt.Sprintf("leerMemoriaUsuario: dirección fuera de rango: %d", direccionFisica),
		)
		return ""
	}

	valor := global.MemoriaUsuario[direccionFisica]
	global.MemoriaLogger.Debug(
		fmt.Sprintf("leerMemoriaUsuario: leyó byte en dirección %d → valor: %q", direccionFisica, valor),
	)
	global.MemoriaMutex.Unlock()
	return string(valor)
}

func escribirMemoriaUsuario(direccionFisica int, texto string) {
	global.MemoriaMutex.Lock()

	longitud := len(texto)
	limite := direccionFisica + longitud
	if direccionFisica < 0 || limite > len(global.MemoriaUsuario) {
		global.MemoriaLogger.Error(
			fmt.Sprintf("escribirMemoriaUsuario: intento de escribir fuera de rango. "+
				"DirInicio=%d, TamañoTexto=%d, TamañoMemoria=%d",
				direccionFisica, longitud, len(global.MemoriaUsuario)),
		)
		return
	}

	bytesTexto := []byte(texto)
	for i := 0; i < longitud; i++ {
		global.MemoriaUsuario[direccionFisica+i] = bytesTexto[i]
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("escribirMemoriaUsuario: escrito texto de tamaño %d en dirección %d. Datos: %q",
			longitud, direccionFisica, texto),
	)

	//TODO
	//ENVIAR EL OK

	global.MemoriaMutex.Unlock()
}
