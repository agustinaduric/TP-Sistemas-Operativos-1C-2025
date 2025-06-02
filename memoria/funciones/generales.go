package fmemoria

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func ConfigurarLog() *logger.LoggerStruct {
	logLevel, error1 := logger.ParseLevel(global.MemoriaConfig.LogLevel)
	if error1 != nil {
		fmt.Println("ERROR: El nivel de log ingresado no es valido")
		os.Exit(1)
	}
	logger, error2 := logger.NewLogger("memoria.log", logLevel)
	if error2 != nil {
		fmt.Println("ERROR: No se pudo crear el logger")
		os.Exit(1)
	}
	return logger
}

func LevantarServidorMemoria() {
	mux := http.NewServeMux()
	mux.HandleFunc("/recibir-handshake", HandshakeKernel)
	mux.HandleFunc("/obtener-instruccion", HandlerObtenerInstruccion)
	mux.HandleFunc("/espacio-libre", HandlerEspacioLibre)
	mux.HandleFunc("/cargar-proceso", HandlerCargarProceso)
	mux.HandleFunc("/conectarcpumemoria", HandshakeCpu)
	mux.HandleFunc("/escribir", HandlerEscribirMemoria)
	mux.HandleFunc("/leer", HandlerLeerMemoria)

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

func HandshakeKernel(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug("HandshakeKernel: entrada")

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

	resp := "OK"
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error codificando respuesta de proceso cargado: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("Confirmacion enviada"),
		)
	}
}

func HandlerEscribirMemoria(w http.ResponseWriter, r *http.Request){
	global.MemoriaLogger.Debug("Entre a HandlerEscribirMemoria")
	var escritura structs.Escritura
	if err := json.NewDecoder(r.Body).Decode(&escritura); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando la solicitud de escritura: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar la solicitud de escritura", http.StatusBadRequest)
		return
	}
	if escritura.DirFisica < 0 || escritura.DirFisica+len(escritura.Datos)> len(global.MemoriaUsuario){
		http.Error(w, "Error: rango fuera de memoria", http.StatusBadRequest)
		return
	}
	copy(global.MemoriaUsuario[escritura.DirFisica:], escritura.Datos)
	global.MemoriaLogger.Debug("Ya escribi en memoria")

	resp := 200
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error codificando respuesta de WRITE realizado: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("Confirmacion WRITE realizado enviada a CPU"),
		)
	}
}

func HandlerLeerMemoria(w http.ResponseWriter, r *http.Request){
	global.MemoriaLogger.Debug("Entre a HandlerLeerMemoria")
	var lectura structs.Lectura
	if err := json.NewDecoder(r.Body).Decode(&lectura); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando la solicitud de lectura: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar la solicitud de lectura", http.StatusBadRequest)
		return
	}
	if lectura.DirFisica < 0 || lectura.DirFisica+lectura.Tamanio> len(global.MemoriaUsuario){
		http.Error(w, "Error: rango fuera de memoria", http.StatusBadRequest)
		return
	}
	global.MemoriaLogger.Debug("Se pudo leer en memoria")
	resultadoLectura := global.MemoriaUsuario[lectura.DirFisica: lectura.DirFisica+lectura.Tamanio]
	json.NewEncoder(w).Encode(resultadoLectura)
	global.MemoriaLogger.Debug("Lectura realizada enviada a CPU")
}