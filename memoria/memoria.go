package main

import (
	"fmt"
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func main() {
	var cfg config.IOConfig
	err := config.CargarConfig("memoria/config/memoria.config.json", &cfg)
	if err != nil{
		log.Fatalf("Error en cargar config memoria %v", err)
	}
	fmt.Println("Config de memoria OK") // borrar
}

/* ***estaba en el archivo viejo de memoria***
var (
	memoriaUsuario []byte       // En vez de (void*) esto es mas god
	memoriaMutex   sync.Mutex   // para proteger la memoria en concurrencia
	swapFile       *os.File     // archivo de SWAP abierto globalmente
)
*/