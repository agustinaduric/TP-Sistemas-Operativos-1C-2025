package cache

import (
	"fmt"
	"time"

	algoritmoCache "github.com/sisoputnfrba/tp-golang/cpu/cache/algoritmos"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func InicializarCachePaginas(algoritmo string) {
	global.AlgoritmoCache = algoritmo
	global.CachePaginas = make([]structs.EntradaCache, 0, global.ConfigCargadito.CacheEntries)
	global.PunteroClock = 0
	global.CpuLogger.Debug(fmt.Sprintf("Se inicializo CachePaginas con algoritmo: %s", algoritmo))
}

func BuscarEncache(pid int, dirLogica int, tamanio int) (bool, []byte) {
	time.Sleep(time.Duration(global.ConfigCargadito.CacheDelay) * time.Millisecond)
	dato := make([]byte, tamanio)
	var desplazamiento int = dirLogica % global.Page_size
	global.CpuLogger.Debug(fmt.Sprintf("Desplazamiento: %d", desplazamiento))
	pagina := dirLogica / global.Page_size

	global.CpuLogger.Debug(fmt.Sprintf("Comenzo busqueda en CachePaginas PID: %d, Pag: %d", pid, pagina))
	for i := 0; i < len(global.CachePaginas); i++ {
		if global.CachePaginas[i].PID == pid && global.CachePaginas[i].Pagina == pagina {
			global.CachePaginas[i].BitUso = true
			//longitudContenido := len(global.CachePaginas[i].Contenido)-1
			global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Cache Hit - Pagina: %d ", pid, pagina))
			for j := desplazamiento; j <= (desplazamiento + tamanio -1); j++ {
				//dato[j] = append(dato, global.CachePaginas[i].Contenido[j])
				dato[j-desplazamiento] = global.CachePaginas[i].Contenido[j]
				//return true, global.CachePaginas[i].Contenido[longitudContenido]
				
			}
			return true, dato
		}
	}
	global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Cache Miss - Pagina: %d ", pid, pagina))
	return false, []byte{}
}

func EscribirEnCache(pid int, dirLogica int, datos []byte, bitModificado bool) {
	time.Sleep(time.Duration(global.ConfigCargadito.CacheDelay) * time.Millisecond)
	var desplazamiento int = dirLogica % global.Page_size
	pagina := dirLogica / global.Page_size
	for i := range global.CachePaginas {
		if global.CachePaginas[i].PID == pid && global.CachePaginas[i].Pagina == pagina {
			for j := desplazamiento; j <= ((len(datos) - 1) + desplazamiento); j++ {
				global.CachePaginas[i].Contenido[j] = datos[j-desplazamiento]
			}
			global.CachePaginas[i].BitUso = true
			global.CachePaginas[i].BitModificado = bitModificado
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
		BitModificado: bitModificado,
	}
	if len(global.CachePaginas) < global.ConfigCargadito.CacheEntries {
		global.CpuLogger.Debug("Hay espacio en la cache, no reemplazo")
		global.CachePaginas = append(global.CachePaginas, nuevaEntrada)
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
	for i := 0; i < len(global.CachePaginas); {
		if global.CachePaginas[i].PID == pid {
			if global.CachePaginas[i].BitModificado {
				algoritmoCache.EnviarEscribirAMemoria(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, &global.CachePaginas[i])
			}
			global.CachePaginas = append(global.CachePaginas[:i], global.CachePaginas[i+1:]...)
		} else {
			i++
		}
	}
	global.CpuLogger.Debug(fmt.Sprintf("Limpieza cache terminada PID: %d", pid))
}

func LeerEncache(pid int, dirLogica int, tamanio int)  []byte {
	time.Sleep(time.Duration(global.ConfigCargadito.CacheDelay) * time.Millisecond)
	dato := make([]byte, tamanio)
	var desplazamiento int = dirLogica % global.Page_size
	global.CpuLogger.Debug(fmt.Sprintf("Desplazamiento: %d", desplazamiento))
	pagina := dirLogica / global.Page_size

	global.CpuLogger.Debug(fmt.Sprintf("Comenzo busqueda en CachePaginas PID: %d, Pag: %d", pid, pagina))
	for i := 0; i < len(global.CachePaginas); i++ {
		if global.CachePaginas[i].PID == pid && global.CachePaginas[i].Pagina == pagina {
			global.CachePaginas[i].BitUso = true
			//longitudContenido := len(global.CachePaginas[i].Contenido)-1
			//global.CpuLogger.Info(fmt.Sprintf("## PID: %d - Cache Hit - Pagina: %d ", pid, pagina))
			for j := desplazamiento; j <=  (desplazamiento + tamanio - 1); j++ {
				//dato[j] = append(dato, global.CachePaginas[i].Contenido[j])
				dato[j-desplazamiento] = global.CachePaginas[i].Contenido[j]
				//return true, global.CachePaginas[i].Contenido[longitudContenido]
				
			}
			return  dato
		}
	}
	global.CpuLogger.Info(fmt.Sprintf("No se encontro en LeerEnCache, no se deberia ver esto"))
	return []byte{}
}