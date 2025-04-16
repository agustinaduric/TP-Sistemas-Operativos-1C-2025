package config

import(
	"encoding/json"
	"os"
)
// ** CargarConfig**
// recibe el path de .json segun el modulo y un puntero al struct para cargarle los datos
func CargarConfig [T any] (path string, cfg *T) error{ 
	file, err := os.Open(path)
	if err != nil{
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(file)
}