package main

import (
	"fmt"
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func main() {
	var cfg config.IOConfig
	err := config.CargarConfig("kernel/config/kernel.config.json", &cfg)
	if err != nil{
		log.Fatalf("Error en cargar config kernel %v", err)
	}
	fmt.Println("Config de kernel OK") // borrar
}
