package main

import (
	"os"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	fmemoria "github.com/sisoputnfrba/tp-golang/memoria/funciones"
)

func main() {
	comunicacion.VerificarParametros(1)
	configPath := os.Args[1]
	fmemoria.IniciarConfiguracionMemoria(configPath)
	fmemoria.IniciarMemoriaUsuario()
	fmemoria.LevantarServidorMemoria()
}
