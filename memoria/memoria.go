package main

import (

	"github.com/sisoputnfrba/tp-golang/memoria/funciones"
)


func main() {
	configCargadito := fmemoria.IniciarConfiguracionMemoria("memoria/config/memoria.config.json")
	fmemoria.LevantarServidorMemoria(configCargadito)
 	
}

/* ***estaba en el archivo viejo de memoria***
var (
	memoriaUsuario []byte       // En vez de (void*) esto es mas god
	memoriaMutex   sync.Mutex   // para proteger la memoria en concurrencia
	swapFile       *os.File     // archivo de SWAP abierto globalmente
)
*/