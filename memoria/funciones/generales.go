package fmemoria

import (
	"encoding/json"
	"fmt"
	"io"
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
	mux.HandleFunc("/pagina", HandlerSolicitudPagina)

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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
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

	global.MemoriaLogger.Debug(
		fmt.Sprintf("## Tamaño de página: %d QUE SE LE PASAN A CPU", datos.Tamaño_pagina),
	)
	global.MemoriaLogger.Debug(
		fmt.Sprintf("## Cantidad de entradas por página: %d QUE SE LE PASAN A CPU", datos.Cant_entradas),
	)
	global.MemoriaLogger.Debug(
		fmt.Sprintf("## Niveles de paginación: %d QUE SE LE PASAN A CPU", datos.Numeros_de_nivel),
	)

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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
	global.MemoriaLogger.Debug("HandlerCargarProceso: entrada")
	var proc structs.Proceso_a_enviar
	if err := json.NewDecoder(r.Body).Decode(&proc); err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error decodificando proceso: %s", err.Error()),
		)
		http.Error(w, "Error al decodificar el proceso recibido", http.StatusBadRequest)
		return
	}
	rutaCompleta := "../pruebas/" + proc.PATH
	instrucciones, err := CargarInstrucciones(rutaCompleta) //____---
	if err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error cargando instrucciones para PID=%d: %s", proc.PID, err.Error()),
		)
		http.Error(w, "Error al cargar las instrucciones", http.StatusInternalServerError)
		return
	}
	InicializarProceso(proc.PID, proc.Tamanio, instrucciones, rutaCompleta)

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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
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
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
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
	time.Sleep(time.Duration(global.MemoriaConfig.SwapDelay) * time.Millisecond)
	global.MemoriaLogger.Debug("HandlerDessuspenderProceso: entrada")

	// Leer todo el cuerpo de la petición
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo cuerpo", http.StatusBadRequest)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerDessuspenderProceso: lectura de body fallida: %s", err))
		return
	}
	defer r.Body.Close()

	// Intentar decodificar como objeto JSON {"pid":X}
	var reqObj struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(data, &reqObj); err != nil {
		// Si falla, intentar decodificar como número puro
		var pidSolo int
		if err2 := json.Unmarshal(data, &pidSolo); err2 != nil {
			http.Error(w, "Error al decodificar solicitud de des-suspensión", http.StatusBadRequest)
			global.MemoriaLogger.Error(fmt.Sprintf(
				"HandlerDessuspenderProceso: decodificación fallida (objeto: %v, número: %v)", err, err2))
			return
		}
		reqObj.PID = pidSolo
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("HandlerDessuspenderProceso: pid=%d recibido", reqObj.PID))

	if err := PedidoDeDesSuspension(reqObj.PID); err != nil {
		http.Error(w, fmt.Sprintf("No se pudo des-suspender PID=%d: %s", reqObj.PID, err), http.StatusConflict)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerDessuspenderProceso: PedidoDeDesSuspension falló: %s", err))
		return
	}

	// retraso de acceso tabla luego de des-suspender
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay*CalcularAccesosTablas(len(RecolectarMarcos(reqObj.PID)))) * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("OK"); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerDessuspenderProceso: error codificando respuesta OK: %s", err))
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("HandlerDessuspenderProceso: PID=%d des-suspendido exitosamente", reqObj.PID))
	}
}

// HandlerSuspenderProceso adapta la solicitud para aceptar un número crudo o un objeto JSON para suspensión.
func HandlerSuspenderProceso(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.SwapDelay) * time.Millisecond)
	global.MemoriaLogger.Debug("HandlerSuspenderProceso: entrada")

	// Leer todo el cuerpo de la petición
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo cuerpo", http.StatusBadRequest)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerSuspenderProceso: lectura de body fallida: %s", err))
		return
	}
	defer r.Body.Close()

	// Intentar decodificar como objeto JSON {"pid":X}
	var reqObj struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(data, &reqObj); err != nil {
		// Si falla, intentar decodificar como número puro
		var pidSolo int
		if err2 := json.Unmarshal(data, &pidSolo); err2 != nil {
			http.Error(w, "Error al decodificar solicitud de suspensión", http.StatusBadRequest)
			global.MemoriaLogger.Error(fmt.Sprintf(
				"HandlerSuspenderProceso: decodificación fallida (objeto: %v, número: %v)", err, err2))
			return
		}
		reqObj.PID = pidSolo
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("HandlerSuspenderProceso: pid=%d recibido", reqObj.PID))

	if err := SuspenderProceso(reqObj.PID); err != nil {
		http.Error(w, fmt.Sprintf("No se pudo suspender PID=%d: %s", reqObj.PID, err.Error()), http.StatusConflict)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerSuspenderProceso: SuspenderProceso falló para PID=%d: %s", reqObj.PID, err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("OK"); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerSuspenderProceso: error codificando respuesta OK: %s", err.Error()))
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("HandlerSuspenderProceso: PID=%d suspendido exitosamente", reqObj.PID))
	}
}

func HandlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
	global.MemoriaLogger.Debug("HandlerFinalizarProceso: entrada")

	// Leer todo el cuerpo de la petición
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo cuerpo", http.StatusBadRequest)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerFinalizarProceso: lectura de body fallida:%s ", err))
		return
	}
	defer r.Body.Close()

	// Intentar decodificar como objeto JSON {"pid":X}
	var reqObj struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(data, &reqObj); err != nil {
		// Si falla, intentar decodificar como número puro
		var pidSolo int
		if err2 := json.Unmarshal(data, &pidSolo); err2 != nil {
			http.Error(w, "Error al decodificar solicitud de finalización", http.StatusBadRequest)
			global.MemoriaLogger.Error(fmt.Sprintf(
				"HandlerFinalizarProceso: decodificación fallida (objeto: %v, número: %v)", err, err2))
			return
		}
		reqObj.PID = pidSolo
	}

	// Llamar a FinalizarProceso con el PID obtenido
	if err := FinalizarProceso(reqObj.PID); err != nil {
		http.Error(w, fmt.Sprintf("No se pudo finalizar PID=%d: %s", reqObj.PID, err), http.StatusConflict)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerFinalizarProceso: FinalizarProceso falló: %s", err))
		return
	}

	// Responder OK
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("OK"); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerFinalizarProceso: respuesta JSON falló: %s", err))
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("HandlerFinalizarProceso: PID=%d finalizado", reqObj.PID))
	}
}

func HandlerMemoryDump(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
	global.MemoriaLogger.Debug("HandlerMemoryDump: entrada")

	// Leer todo el cuerpo de la petición
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo cuerpo", http.StatusBadRequest)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerMemoryDump: lectura de body fallida: %s", err))
		return
	}
	defer r.Body.Close()

	// Intentar decodificar como objeto JSON {"pid":X}
	var reqObj struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(data, &reqObj); err != nil {
		// Si falla, intentar decodificar como número puro
		var pidSolo int
		if err2 := json.Unmarshal(data, &pidSolo); err2 != nil {
			http.Error(w, "Error al decodificar solicitud de dump", http.StatusBadRequest)
			global.MemoriaLogger.Error(fmt.Sprintf(
				"HandlerMemoryDump: decodificación fallida (objeto: %v, número: %v)", err, err2))
			return
		}
		reqObj.PID = pidSolo
	}

	// Llamar a DumpMemory con el PID obtenido
	if err := DumpMemory(reqObj.PID); err != nil {
		http.Error(w, fmt.Sprintf("No se pudo generar dump para PID=%d: %s", reqObj.PID, err), http.StatusInternalServerError)
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerMemoryDump: DumpMemory falló: %s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := structs.Devolucion_DumpMemory{
		PID:       reqObj.PID,
		Respuesta: "OK",
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("HandlerMemoryDump: respuesta JSON falló: %s", err))
		http.Error(w, "Error enviando respuesta", http.StatusInternalServerError)
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("HandlerMemoryDump: dump generado para PID=%d", reqObj.PID))
	}
}

func HandlerSolicitudMarco(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay) * time.Millisecond)
	var req struct {
		PID     int
		Indices []int
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar solicitud de marco", http.StatusBadRequest)
		return
	}
	//retraso de acceso tabla
	time.Sleep(time.Duration(global.MemoriaConfig.MemoryDelay*CalcularAccesosTablas(len(RecolectarMarcos(req.PID)))) * time.Millisecond)

	marco := Marco(req.PID, req.Indices) //miri enviarme indices
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(marco); err != nil {
		http.Error(w, "Error al codificar respuesta de marco", http.StatusInternalServerError)
	}
}

func DumpMemory(pid int) error {
	cfg := global.MemoriaConfig
	global.MemoriaLogger.Debug(fmt.Sprintf("DumpMemory: inicio PID=%d", pid))
	global.MemoriaLogger.Info(fmt.Sprintf("## PID: %d - Memory Dump solicitado", pid))

	// 1) Recolectar marcos asignados
	var marcosOcupados []int
	global.MemoriaLogger.Debug("[DumpMemory] intentando tomar MemoriaMutex para recolectar marcos")
	global.MemoriaMutex.Lock()
	global.MemoriaLogger.Debug("[DumpMemory] tomó MemoriaMutex para recolectar marcos")
	for idx, ocup := range global.MapMemoriaDeUsuario {
		if ocup == pid {
			marcosOcupados = append(marcosOcupados, idx)
		}
	}
	global.MemoriaMutex.Unlock()
	global.MemoriaLogger.Debug("[DumpMemory] liberó MemoriaMutex tras recolectar marcos")

	if len(marcosOcupados) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("DumpMemory: PID=%d sin marcos asignados", pid))
		return fmt.Errorf("PID=%d no tiene memoria asignada", pid)
	}

	sort.Ints(marcosOcupados)

	// 2) Asegurar que el directorio exista
	global.MemoriaLogger.Debug(fmt.Sprintf("[DumpMemory] asegurando existencia de directorio '%s'", cfg.DumpPath))
	if err := os.MkdirAll(cfg.DumpPath, 0755); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("DumpMemory: fallo creando directorio '%s': %s", cfg.DumpPath, err))
		return fmt.Errorf("no se pudo crear directorio dump: %w", err)
	}

	// 3) Crear el archivo de dump
	timestamp := time.Now().Format("20060102_150405.000") // con milisegundos
	nombre := fmt.Sprintf("%d-%s.dmp", pid, timestamp)
	ruta := filepath.Join(cfg.DumpPath, nombre)
	global.MemoriaLogger.Debug(fmt.Sprintf("[DumpMemory] creando archivo '%s'", ruta))

	archivo, err := os.Create(ruta)
	if err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("DumpMemory: fallo creando '%s': %s", ruta, err))
		return fmt.Errorf("no se pudo crear dump: %w", err)
	}
	defer archivo.Close()

	// 4) Escribir cada página protegiendo MemoriaUsuario
	pageSize := cfg.PageSize
	for _, marco := range marcosOcupados {
		// Tomar mutex para copiar la página
		global.MemoriaLogger.Debug("[DumpMemory] intentando tomar MemoriaMutex para copiar página")
		global.MemoriaMutex.Lock()
		global.MemoriaLogger.Debug("[DumpMemory] tomó MemoriaMutex para copiar página")

		inicio := marco * pageSize
		fin := inicio + pageSize
		if inicio < 0 || fin > len(global.MemoriaUsuario) {
			global.MemoriaLogger.Error(fmt.Sprintf(
				"DumpMemory: marco inválido %d → rango [%d,%d) fuera de memoria",
				marco, inicio, fin,
			))
			global.MemoriaMutex.Unlock()
			return fmt.Errorf("memoria fuera de rango al dump: marco %d", marco)
		}

		// Copiamos bajo lock
		datos := make([]byte, pageSize)
		copy(datos, global.MemoriaUsuario[inicio:fin])

		// Liberar mutex antes de escribir en disco
		global.MemoriaMutex.Unlock()
		global.MemoriaLogger.Debug("[DumpMemory] liberó MemoriaMutex antes de escribir página")

		// Escritura
		escritos, err := archivo.Write(datos)
		if err != nil {
			global.MemoriaLogger.Error(fmt.Sprintf(
				"DumpMemory: error escribiendo marco %d: %s", marco, err,
			))
			return fmt.Errorf("error al escribir dump: %w", err)
		}
		global.MemoriaLogger.Debug(fmt.Sprintf(
			"DumpMemory: PID=%d marco %d volcado (%d bytes)", pid, marco, escritos,
		))
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"DumpMemory: fin PID=%d, archivo '%s' creado", pid, ruta,
	))
	return nil
}
func HandlerSolicitudPagina(w http.ResponseWriter, r *http.Request) {
	global.MemoriaLogger.Debug(fmt.Sprintf("INGRESE al HandlerSolicitudPagina"))
	var direccionFisica int
	if err := json.NewDecoder(r.Body).Decode(&direccionFisica); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("Error decodificando DF: %s", err.Error()))
		http.Error(w, "Error al decodificar la solicitud de instruccion", http.StatusBadRequest)
		return
	}
	pagina := ObtenerPagina(direccionFisica)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pagina); err != nil {
		http.Error(w, "Error al codificar respuesta de pagina", http.StatusInternalServerError)
	}
}
