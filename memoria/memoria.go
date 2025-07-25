package main

import (
	"os"

	fmemoria "github.com/sisoputnfrba/tp-golang/memoria/funciones"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
)

func main() {
	comunicacion.VerificarParametros(1)
	configPath := os.Args[1]
	fmemoria.IniciarConfiguracionMemoria(configPath)
	global.MemoriaLogger = fmemoria.ConfigurarLog()
	fmemoria.IniciarMemoriaUsuario()
	fmemoria.LimpiarSwap()
	fmemoria.LevantarServidorMemoria()
}
