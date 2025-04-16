package main

import (
	"fmt"
	"log"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func main() {
	var cfg config.CPUConfig
	err := config.CargarConfig("cpu/config/cpu.config.json", &cfg)
	if err != nil{
		log.Fatalf("Error en cargar config CPU %v", err)
	}
	fmt.Println("Config de CPU OK") // borrar
}