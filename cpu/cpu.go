package main

import (
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/cache"
	fCpu "github.com/sisoputnfrba/tp-golang/cpu/funciones"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {

	comunicacion.VerificarParametros(2)
	global.Nombre = os.Args[1]
	configPath := os.Args[2]
	global.ConfigCargadito = fCpu.IniciarConfiguracionCpu(configPath)
	global.CpuLogger = fCpu.ConfigurarLog()
	go fCpu.LevantarServidorCPU()
	protocolos.Conectarse_con_Kernel(global.Nombre)
	protocolos.Conectarse_con_Memoria(global.Nombre)
	cache.InicializarCachePaginas(global.ConfigCargadito.CacheReplacement)
	global.WgCPU.Wait()
	global.CpuLogger.Close()
}
