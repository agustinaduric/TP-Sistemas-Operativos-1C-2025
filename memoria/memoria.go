package main

import (
	fmemoria "github.com/sisoputnfrba/tp-golang/memoria/funciones"
)

func main() {
	fmemoria.IniciarConfiguracionMemoria("memoria/config/memoria.config.json")
	fmemoria.IniciarMemoriaUsuario()
	fmemoria.LevantarServidorMemoria()
}
