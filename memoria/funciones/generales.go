package fmemoria

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
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
	mux.HandleFunc("/desuspension-proceso", HandlerDesSuspenderProceso)
	mux.HandleFunc("/suspension-proceso", HandlerSuspenderProceso)
	mux.HandleFunc("/finalizar-proceso", HandlerFinalizarProceso)
	mux.HandleFunc("/memory-dump", HandlerMemoryDump)
	mux.HandleFunc("/solicitud-marco", HandlerSolicitudMarco)

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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
	global.MemoriaLogger.Debug("HandlerCargarProceso: entrada")

	var proc structs.Proceso_a_enviar
	if err := json.NewDecoder(r.Body).Decode(&proc); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando proceso: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar el proceso recibido", http.StatusBadRequest)
		return
	}

	instrucciones, err := CargarInstrucciones(proc.PATH) //____---
	if err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error cargando instrucciones para PID=%d: %s", proc.PID, err.Error()),
		)
		http.Error(w, "Error al cargar las instrucciones", http.StatusInternalServerError)
		return
	}
	InicializarProceso(proc.PID, proc.Tamanio, instrucciones, proc.PATH)

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

func HandlerEscribirMemoria(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
	global.MemoriaLogger.Debug("Entre a HandlerEscribirMemoria")
	var escritura structs.Escritura
	if err := json.NewDecoder(r.Body).Decode(&escritura); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando la solicitud de escritura: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar la solicitud de escritura", http.StatusBadRequest)
		return
	}
	EscribirMemoriaUsuario(escritura.PID, escritura.DirFisica, escritura.Datos)
	global.MemoriaLogger.Debug("Confirmacion WRITE realizado enviada a CPU")
}

func HandlerLeerMemoria(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
	global.MemoriaLogger.Debug("Entre a HandlerLeerMemoria")
	var lectura structs.Lectura
	if err := json.NewDecoder(r.Body).Decode(&lectura); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando la solicitud de lectura: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar la solicitud de lectura", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(LeerMemoriaUsuario(lectura.PID, lectura.DirFisica, lectura.Tamanio))
	global.MemoriaLogger.Debug("Lectura realizada enviada a CPU")
}

func HandlerDesSuspenderProceso(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.SwapDelay))

	global.MemoriaLogger.Debug("HandlerDessuspenderProceso: entrada")

	var req struct {
		PID int
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("HandlerDessuspenderProceso: error decodificando PID: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar solicitud de des-suspensión", http.StatusBadRequest)
		return
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"HandlerDessuspenderProceso: pid=%d recibido", req.PID,
	))

	if err := PedidoDeDesSuspension(req.PID); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("HandlerDessuspenderProceso: PedidoDeDesSuspension falló: %s", err.Error()),
		)
		http.Error(w, fmt.Sprintf("No se pudo des-suspender PID=%d: %s", req.PID, err.Error()), http.StatusConflict)
		return
	}
	//retraso de acceso tabla
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay * CalcularAccesosTablas(len(RecolectarMarcos(req.PID)))))
	resp := "OK"
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("HandlerDessuspenderProceso: error codificando respuesta OK: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("HandlerDessuspenderProceso: PID=%d des-suspendido exitosamente", req.PID),
		)
	}
}

func HandlerSuspenderProceso(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.SwapDelay))
	global.MemoriaLogger.Debug("HandlerSuspenderProceso: entrada")

	var req struct {
		PID int
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("HandlerSuspenderProceso: error decodificando PID: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar solicitud de suspensión", http.StatusBadRequest)
		return
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"HandlerSuspenderProceso: pid=%d recibido", req.PID,
	))
	//retraso de acceso tabla
	//time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay * CalcularAccesosTablas(len(RecolectarMarcos(req.PID)))))
	if err := SuspenderProceso(req.PID); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("HandlerSuspenderProceso: SuspenderProceso falló para PID=%d: %s", req.PID, err.Error()),
		)
		http.Error(w, fmt.Sprintf("No se pudo suspender PID=%d: %s", req.PID, err.Error()), http.StatusConflict)
		return
	}
	resp := "OK"
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("HandlerSuspenderProceso: error codificando respuesta OK: %s", err.Error()),
		)
	} else {
		global.MemoriaLogger.Debug(
			fmt.Sprintf("HandlerSuspenderProceso: PID=%d suspendido exitosamente", req.PID),
		)
	}
}

func HandlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
	global.MemoriaLogger.Debug("HandlerFinalizarProceso: entrada")
	var req struct {
		PID int `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar solicitud de finalización", http.StatusBadRequest)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerFinalizarProceso: decodificación fallida: %s", err))
		return
	}
	if err := FinalizarProceso(req.PID); err != nil {
		http.Error(w, fmt.Sprintf("No se pudo finalizar PID=%d: %s", req.PID, err), http.StatusConflict)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerFinalizarProceso: FinalizarProceso falló: %s", err))
		return
	}
	resp := "OK"
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerFinalizarProceso: respuesta JSON falló: %s", err))
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("HandlerFinalizarProceso: PID=%d finalizado", req.PID))
	}
}

func HandlerMemoryDump(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
	global.MemoriaLogger.Debug("HandlerMemoryDump: entrada")
	var req struct {
		PID int `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar solicitud de dump", http.StatusBadRequest)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerMemoryDump: decodificación fallida: %s", err))
		return
	}
	//retraso de acceso tabla
	//time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay * CalcularAccesosTablas(len(RecolectarMarcos(req.PID)))))
	if err := DumpMemory(req.PID); err != nil {
		http.Error(w, fmt.Sprintf("No se pudo generar dump para PID=%d: %s", req.PID, err), http.StatusInternalServerError)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerMemoryDump: DumpMemory falló: %s", err))
		return
	}
	resp := "OK"
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerMemoryDump: respuesta JSON falló: %s", err))
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("HandlerMemoryDump: dump generado para PID=%d", req.PID))
	}
}

func HandlerSolicitudMarco(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay))
	var req struct {
		PID     int
		Indices []int
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar solicitud de marco", http.StatusBadRequest)
		return
	}
	//retraso de acceso tabla
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay * CalcularAccesosTablas(len(RecolectarMarcos(req.PID)))))

	marco := Marco(req.PID, req.Indices) //miri enviarme indices
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(marco); err != nil {
		http.Error(w, "Error al codificar respuesta de marco", http.StatusInternalServerError)
	}
}

func DumpMemory(pid int) error { //Cabe aclarar que
	cfg := global.MemoriaConfig
	global.MemoriaLogger.Debug(fmt.Sprintf("DumpMemory: inicio PID=%d", pid))

	//Agarramos todos los marcos que le pertenecen al proceso
	var marcosOcupados []int
	global.MemoriaMutex.Lock()
	for indice, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcosOcupados = append(marcosOcupados, indice)
		}
	}
	global.MemoriaMutex.Unlock()

	if len(marcosOcupados) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("DumpMemory: PID=%d sin marcos asignados", pid))
		return fmt.Errorf("PID=%d no tiene memoria asignada", pid)
	}

	//cuando le pregunte a chatgpt cuestiones sobre el memory dump dijo que
	//era bueno ordenarlos, asi que lo hacemos
	sort.Ints(marcosOcupados)

	//archivitp
	timestamp := time.Now().Unix()
	nombreArchivo := fmt.Sprintf("%d-%d.dmp", pid, timestamp)
	rutaCompleta := filepath.Join(cfg.DumpPath, nombreArchivo)

	archivo, err := os.Create(rutaCompleta)
	if err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("DUmpMemory: fallo creando '%s': %s", rutaCompleta, err))
		return fmt.Errorf("no se pudo crear dump: %w", err)
	}
	defer archivo.Close()

	//y de chill recoremos memoria usuario asi eszcribimos en el archivo lo q dice cada marco, insta
	tamañoPagina := cfg.PageSize
	for _, numeroMarco := range marcosOcupados {
		desplazamiento := numeroMarco * tamañoPagina
		pagina := global.MemoriaUsuario[desplazamiento : desplazamiento+tamañoPagina]

		escritos, err := archivo.Write(pagina)
		if err != nil {
			global.MemoriaLogger.Error(fmt.Sprintf(
				"DumpMemory: error escribiendo marco %d: %s",
				numeroMarco, err,
			))
			return fmt.Errorf("error al escribir dump: %w", err)
		}

		global.MemoriaLogger.Debug(fmt.Sprintf(
			"DumpMemory: PID=%d marco %d volcado (%d bytes)",
			pid, numeroMarco, escritos,
		))
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"DumpMemory: fin PID=%d, archivo '%s' creado",
		pid, rutaCompleta,
	))
	return nil
}
