package fmemoria

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func LevantarServidorMemoria() {
	mux := http.NewServeMux()
	mux.HandleFunc("/recibir-handshake", HandlerRecibirHandshake)
	mux.HandleFunc("/obtener-instruccion", HandlerObtenerInstruccion)
	mux.HandleFunc("/espacio-libre", HandlerEspacioLibre)
	mux.HandleFunc("/cargar-proceso", HandlerCargarProceso)
	mux.HandleFunc("/conectarcpumemoria", HandshakeCpu)

	puerto := config.IntToStringConPuntos(global.MemoriaConfig.PortMemory)
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Servidor de Memoria iniciándose en %s", puerto),
	)

	if err := http.ListenAndServe(puerto, mux); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error al levantar el servidor: %s", err.Error()),
		)
		os.Exit(1)
	}
}

func HandlerRecibirHandshake(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug("HandlerRecibirHandshake: entrada")

	var handshake structs.Handshake
	if err := json.NewDecoder(r.Body).Decode(&handshake); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando handshake: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar mensaje", http.StatusBadRequest)
		return
	}
	global.IPkernel = handshake.IP
	global.PuertoKernel = handshake.Puerto

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Handshake recibido: IP=%s, Puerto=%d", handshake.IP, handshake.Puerto),
	)
}

func HandlerObtenerInstruccion(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug("HandlerObtenerInstruccion: entrada")

	var req struct {
		PID int `json:"pid"`
		PC  int `json:"pc"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando solicitud PID=%d, PC=%d: %s",
				req.PID, req.PC, err.Error(),
			),
		)
		http.Error(w, "Error al decodificar la solicitud de instrucción", http.StatusBadRequest)
		return
	}

	instr, err := BuscarInstruccion(req.PID, req.PC)
	if err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error buscando instrucción PID=%d, PC=%d: %s",
				req.PID, req.PC, err.Error(),
			),
		)
		http.Error(w, "No se pudo obtener la instrucción", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(instr); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error codificando instrucción para CPU: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("Instrucción enviada: PID=%d, PC=%d", req.PID, req.PC),
		)
	}
}

func HandshakeCpu(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug("HandshakeCpu: entrada")

	var cpu structs.CPU_a_memoria
	if err := json.NewDecoder(r.Body).Decode(&cpu); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando handshake CPU: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar mensaje", http.StatusBadRequest)
		return
	}

	datos := structs.Datos_memoria{
		Tamaño_pagina:    global.MemoriaConfig.PageSize,
		Cant_entradas:    global.MemoriaConfig.EntriesPerPage,
		Numeros_de_nivel: global.MemoriaConfig.NumberOfLevels,
	}
	if err := json.NewEncoder(w).Encode(datos); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error codificando datos para CPU: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("Handshake CPU completado: %+v", datos),
		)
	}
}

func HandlerEspacioLibre(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug("HandlerEspacioLibre: entrada")

	espacio := espacioDisponible()
	resp := structs.EspacioLibreRespuesta{BytesLibres: espacio}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error codificando respuesta espacio libre: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("Espacio libre reportado: %d bytes", espacio),
		)
	}
}

func HandlerCargarProceso(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug("HandlerCargarProceso: entrada")

	var proc structs.Proceso_a_enviar
	if err := json.NewDecoder(r.Body).Decode(&proc); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando proceso: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar el proceso recibido", http.StatusBadRequest)
		return
	}

	instrucciones, err := CargarInstrucciones(proc.PATH)
	if err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error cargando instrucciones para PID=%d: %s", proc.PID, err.Error()),
		)
		http.Error(w, "Error al cargar las instrucciones", http.StatusInternalServerError)
		return
	}

	global.Procesos = append(global.Procesos, structs.ProcesoMemoria{
		PID:           proc.PID,
		Tamanio:       proc.Tamanio,
		EnSwap:        false,
		Path:          proc.PATH,
		Instrucciones: instrucciones,
	})
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Proceso cargado: PID=%d, Tamanio=%d, Instrucciones=%d",
			proc.PID, proc.Tamanio, len(instrucciones),
		),
	)

	url := fmt.Sprintf("http://%s:%d/confirmacion", global.IPkernel, global.PuertoKernel)
	body, _ := json.Marshal("OK")
	if _, err := http.Post(url, "application/json", bytes.NewBuffer(body)); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error enviando confirmación a Kernel: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug("Confirmación enviada a Kernel")
	}
}
