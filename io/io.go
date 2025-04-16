package main

import (
	"fmt"
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func main() {
	var cfg config.IOConfig
	err := config.CargarConfig("io/config/io.config.json", &cfg)
	if err != nil{
		log.Fatalf("Error en cargar config IO %v", err)
	}
	fmt.Println("Config de IO OK") // borrar
}