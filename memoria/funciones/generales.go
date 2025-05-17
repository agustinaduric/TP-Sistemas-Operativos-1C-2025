package fmemoria

import (
	"log"
	"net/http"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func LevantarServidorMemoria() {
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje) // borrar, 1er check
	mux.HandleFunc("/obtener-instruccion", HandlerObtenerInstruccion)

	puerto := config.IntToStringConPuntos(global.MemoriaConfig.PortMemory)

	log.Printf("Servidor de Memoria escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}

func HandlerObtenerInstruccion(w http.ResponseWriter, r *http.Request){
	var proceso struct {
		PID int `json:"pid"`
		PC int `json:"pc"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&proceso)
	if err != nil {
		log.Printf("Error al decodificar la instruccion del proceso: PID: %d, PC:%d, %s\n", proceso.PID, proceso.PC, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar la instruccion"))
		return
	}
	instruccion,err := BuscarInstruccion(proceso.PID, proceso.PC)
	if err != nil {
		log.Printf("Error al buscar instruccion PID: %d, PC:%d, %s\n", proceso.PID, proceso.PC, err.Error())
		http.Error(w, "No se pudo obbtener la instruccion", http.StatusNotFound)
		return
	}
	errCodif := json.NewEncoder(w).Encode(instruccion) // le respondo a cpu
	if errCodif != nil {
		log.Printf("Error al codificar la instruccion para CPU %s\n", errCodif.Error())
		return
	}
}