package main

import (
	fmemoria "github.com/sisoputnfrba/tp-golang/memoria/funciones"
)

var MemoriaUsuario []byte

func main() {
	configCargadito := fmemoria.IniciarConfiguracionMemoria("memoria/config/memoria.config.json")
	fmemoria.LevantarServidorMemoria(configCargadito)

	MemoriaUsuario = fmemoria.IniciarMemoriaUsuario(configCargadito)

}
