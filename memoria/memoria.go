package main

import (
	fmemoria "github.com/sisoputnfrba/tp-golang/memoria/funciones"
)

func main() {
	configCargadito := fmemoria.IniciarConfiguracionMemoria("memoria/config/memoria.config.json")
	fmemoria.LevantarServidorMemoria(configCargadito)

}
