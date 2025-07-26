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

	global.MemoriaLogger.Info(
		fmt.Sprintf("## PID: %d - Lectura - Dir. Física: %d - Tamaño: %d",
			pid, direccionFisica, len(valor)),
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
	global.MemoriaLogger.Debug(
		fmt.Sprintf("[PrepararEscritura] PID=%d, bytes a escribir en %d..%d: %v",
			pid, direccionFisica, limite-1, bytesTexto,
		),
	)

	for i := 0; i < longitud; i++ {
		global.MemoriaUsuario[direccionFisica+i] = bytesTexto[i]
	}

	global.MemoriaLogger.Info(
		fmt.Sprintf("## PID: %d - Escritura - Dir. Física: %d - Tamaño: %d",
			pid, direccionFisica, longitud),
	)
	// 3) Leer de nuevo el bloque escrito y mostrarlo en Debug
	escrito := make([]byte, longitud)
	copy(escrito, global.MemoriaUsuario[direccionFisica:limite])
	global.MemoriaLogger.Debug(
		fmt.Sprintf("[VerificarEscritura] PID=%d, bytes leídos en %d..%d: %v",
			pid, direccionFisica, limite-1, escrito,
		),
	)

	global.MemoriaMutex.Unlock()
}

func ObtenerPagina(direccionFisica int) []byte {
	// 1) Obtener el tamaño de página
	tamanioPagina := int(global.MemoriaConfig.PageSize)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"[ObtenerPagina] Dirección física recibida: %d, Tamaño de página: %d",
		direccionFisica, tamanioPagina,
	))

	// 2) Calcular número de página e índices de inicio y fin
	numeroPagina := int(direccionFisica / tamanioPagina)
	inicio := numeroPagina * tamanioPagina
	fin := inicio + tamanioPagina
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"[ObtenerPagina] Página calculada: %d (bytes %d a %d)",
		numeroPagina, inicio, fin-1,
	))

	// 3) Verificar que no se excedan los límites de memoria
	totalMemoria := len(global.MemoriaUsuario)
	if fin > totalMemoria {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"[ObtenerPagina] Índice final %d excede tamaño de memoria (%d)",
			fin, totalMemoria,
		))
		return []byte{}
	}

	// 4) Copiar los datos de la página
	datosPagina := make([]byte, tamanioPagina)
	copy(datosPagina, global.MemoriaUsuario[inicio:fin])
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"[ObtenerPagina] Datos de la página %d extraídos: %v",
		numeroPagina, datosPagina,
	))

	return datosPagina
}
