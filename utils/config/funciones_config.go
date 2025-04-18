package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ** CargarConfig**
// recibe el path de .json segun el modulo y un puntero al struct para cargarle los datos
func CargarConfig[T any](path string, cfg *T) error { //Comentario de nardo: para q pinga esta ese cfg
	file, err := os.Open(path)
	if err != nil {
		log.Printf("error")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(file)
	//jsonParser.Decode(&config)
	//jsonParser := json.NewDecoder(configFile)
}

func IntToStringConPuntos(valor int) string {
	return fmt.Sprintf(":%d", valor)
}
