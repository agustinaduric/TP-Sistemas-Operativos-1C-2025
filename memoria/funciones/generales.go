package fmemoria

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func LevantarServidorMemoria() {
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje) // borrar, 1er check
	mux.HandleFunc("/obtener-instruccion", HandlerObtenerInstruccion)
	mux.HandleFunc("/espacio-libre", HandlerEspacioLibre)
	mux.HandleFunc("/cargar-proceso", HandlerCargarProceso)

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
		log.Printf("Error al decodificar la solicitud de instruccion PID: %d, PC:%d, %s\n", proceso.PID, proceso.PC, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar la solicitud de instruccion"))
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

func HandlerEspacioLibre(w http.ResponseWriter, r *http.Request){
	espacio := espacioDisponible()
	respuestaEspacio := structs.EspacioLibreRespuesta{ BytesLibres: espacio,}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(respuestaEspacio)
	if err != nil {
		log.Printf("Error al codificar la respuesta de espacio memoria %s\n", err.Error())
		return
	}
}

func HandlerCargarProceso(w http.ResponseWriter, r *http.Request){
	var proceso structs.Proceso_a_enviar
	jsonParser := json.NewDecoder(r.Body)
	err := jsonParser.Decode(&proceso)
	if err != nil {
		http.Error(w, "Error en decodificar el proceso recibido: "+err.Error(), http.StatusBadRequest)
		return
	}
	instrucciones, errCargar := CargarInstrucciones(proceso.PATH)
	if errCargar != nil {
		http.Error(w, "Error en cargar las instrucciones: "+errCargar.Error(), http.StatusInternalServerError)
		return
	}
	global.Procesos = append(global.Procesos, structs.ProcesoMemoria{PID: proceso.PID,Tamanio: proceso.Tamanio, EnSwap: false, Path: proceso.PATH, Instrucciones: instrucciones})
	// le confirmo a kernel que se cargo:
	urlKernel := fmt.Sprintf("http:/%s:%d/confirmacion", global.MemoriaConfig.IpMemory, global.MemoriaConfig.PortMemory)
	// *****NOTA***** 
	/*Acá hay un problema, necesito la ip y el puerto del kernel para confirmar que cargué
	el proceso en memoria y hacerle un post, pero pero pero el .json de memoria no tiene
	la ip ni el puerto de kernel.
	Para que no tire error, deje los de memoria. Hay que sacarlos y poner kernel(?
	*/
	body,errCodificacion := json.Marshal("OK")
	if errCodificacion != nil{
		log.Printf("Error al codificar la confirmacion: %s",errCodificacion.Error())
		return
	}
	_,errEnvio := http.Post(urlKernel, "application/json", bytes.NewBuffer(body))
	if errEnvio != nil{
		log.Printf("Error al enviar la confirmacion: %s",errEnvio.Error())
		return
	}
}