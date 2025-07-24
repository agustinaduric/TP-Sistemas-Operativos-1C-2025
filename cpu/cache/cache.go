package cache

import (
	"fmt"
	"time"

	algoritmoCache "github.com/sisoputnfrba/tp-golang/cpu/cache/algoritmos"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var cachePags = global.CachePaginas

func InicializarCachePaginas(tamanio int, algoritmo string) {
	global.EntradasMaxCache = tamanio
	global.AlgoritmoCache = algoritmo
	cachePags = make([]structs.EntradaCache, 0, global.EntradasMaxCache)
	global.PunteroClock = 0
	global.CpuLogger.Debug(fmt.Sprintf("Se inicializo CachePaginas con algoritmo: %s", algoritmo))
}

func BuscarEncache(pid int, dirLogica int) (bool, byte) {
	time.Sleep(time.Duration(global.ConfigCargadito.CacheDelay) * time.Millisecond)

	pagina := dirLogica / global.Page_size

	global.CpuLogger.Debug(fmt.Sprintf("Comenzo busqueda en CachePaginas PID: %d, Pag: %d", pid, pagina))
	for i := 0; i < len(cachePags); i++ {
		if cachePags[i].PID == pid && cachePags[i].Pagina == pagina {
			cachePags[i].BitUso = true
			global.CpuLogger.Info(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d ", pid, pagina))
			return true, cachePags[i].Contenido[0]
		}
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d ", pid, pagina))
	return false, 0
}

func EscribirEnCache(pid int, dirLogica int, datos []byte) {
	time.Sleep(time.Duration(global.ConfigCargadito.CacheDelay) * time.Millisecond)

	pagina := dirLogica / global.Page_size

	for i := range cachePags {
		if cachePags[i].PID == pid && cachePags[i].Pagina == pagina {
			cachePags[i].Contenido = datos
			cachePags[i].BitUso = true
			cachePags[i].BitModificado = true
			global.CpuLogger.Debug("Ya existe esa entrada")
			return
		}
	}
	global.CpuLogger.Debug("No existe la entrada")
	nuevaEntrada := structs.EntradaCache{
		PID:           pid,
		Pagina:        pagina,
		Contenido:     datos,
		BitUso:        true,
		BitModificado: true,
	}
	if len(cachePags) < global.EntradasMaxCache {
		global.CpuLogger.Debug("Hay espacio en la cache, no reemplazo")
		cachePags = append(cachePags, nuevaEntrada)
		global.CpuLogger.Info(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", pid, pagina))
		global.CpuLogger.Debug("Se escribio en la cache")
	} else {
		global.CpuLogger.Debug("No hay espacio en la cache, reemplazo")
		ReemplazarEntrada(nuevaEntrada)
	}
}

func ReemplazarEntrada(nuevaEntrada structs.EntradaCache) {
	switch global.AlgoritmoCache {
	case "CLOCK":
		algoritmoCache.Clock(nuevaEntrada)
	case "CLOCK-M":
		algoritmoCache.ClockM(nuevaEntrada)
	default:
		global.CpuLogger.Error(fmt.Sprintf("Algoritmo de reeemplazo no valido: %s", global.AlgoritmoCache))
	}
}

func LimpiarCacheDelProceso(pid int) {
	global.CpuLogger.Debug(fmt.Sprintf("Limpiando cache PID: %d", pid))
	for i := 0; i < len(cachePags); {
		if cachePags[i].PID == pid {
			if cachePags[i].BitModificado {
				algoritmoCache.EnviarEscribirAMemoria(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, &cachePags[i])
			}
			cachePags = append(cachePags[:i], cachePags[i+1:]...)
		} else {
			i++
		}
	}
	global.CpuLogger.Debug(fmt.Sprintf("Limpieza cache terminada PID: %d", pid))
}
