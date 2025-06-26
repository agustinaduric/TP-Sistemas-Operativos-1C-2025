package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

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

func LeerMemoriaUsuario(pid int, direccionFisica int, tamanio int) []byte {
	global.MemoriaMutex.Lock()
	IncrementarLecturasMem(pid)
	if direccionFisica < 0 || direccionFisica >= len(global.MemoriaUsuario) {
		global.MemoriaLogger.Error(
			fmt.Sprintf("leerMemoriaUsuario: dirección fuera de rango: %d", direccionFisica),
		)
		return nil
	}

	fin := direccionFisica + tamanio
	if fin > len(global.MemoriaUsuario) {
		fin = len(global.MemoriaUsuario)
	}

	valor := global.MemoriaUsuario[direccionFisica:fin]
	global.MemoriaLogger.Debug(
		fmt.Sprintf("leerMemoriaUsuario: leyó byte en dirección %d → valor: %q", direccionFisica, valor),
	)
	global.MemoriaMutex.Unlock()
	return valor
}

func EscribirMemoriaUsuario(pid int, direccionFisica int, bytesTexto []byte) {
	global.MemoriaMutex.Lock()
	IncrementarEscriturasMem(pid)
	longitud := len(bytesTexto)
	limite := direccionFisica + longitud
	if direccionFisica < 0 || limite > len(global.MemoriaUsuario) {
		global.MemoriaLogger.Error(
			fmt.Sprintf("escribirMemoriaUsuario: intento de escribir fuera de rango. "+
				"DirInicio=%d, TamañoTexto=%d, TamañoMemoria=%d",
				direccionFisica, longitud, len(global.MemoriaUsuario)),
		)
		return
	}

	for i := 0; i < longitud; i++ {
		global.MemoriaUsuario[direccionFisica+i] = bytesTexto[i]
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("escribirMemoriaUsuario: escrito texto de tamaño %d en dirección %d. Datos: %q",
			longitud, direccionFisica, bytesTexto),
	)

	global.MemoriaMutex.Unlock()
}
