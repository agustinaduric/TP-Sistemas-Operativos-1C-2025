package ciclo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func WRITE(direccion int, datos string) {
	global.CpuLogger.Debug(fmt.Sprintf("Entro a WRITE, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, direccion))
	soliEscritura := structs.Escritura{
		PID: global.Proceso_Ejecutando.PID,
		DirFisica: direccion,
		Datos: []byte(datos),
	}
	body, err := json.Marshal(soliEscritura)
	if err != nil{
		global.CpuLogger.Error(fmt.Sprintf("Error al serializar: %s", err.Error()))
		return
	}
	urlMemoria := fmt.Sprintf("http://%s:%d/escribir", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	respEnvio, errEnviar := http.Post(urlMemoria, "application/json", bytes.NewBuffer(body))
	if errEnviar != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al enviar solicitud escritura a memoria: %s", errEnviar.Error()))
		return
	}
	global.CpuLogger.Debug(fmt.Sprintf("Envie WRITE a memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, direccion))
	defer respEnvio.Body.Close()
	// me responde memoria:
	global.CpuLogger.Debug(fmt.Sprintf("Me respondio el WRITE memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, direccion))
	if respEnvio.StatusCode != 200 {
		global.CpuLogger.Error(fmt.Sprintf("Memoria devolvio error en WRITE: %d", respEnvio.StatusCode))
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: %d, - Accion: ESCRIBIR, Direccion fisica: %d, Valor Escrito: %s", global.Proceso_Ejecutando.PID, direccion, datos))
}

func READ(direccion int, tamanio int) {
	global.CpuLogger.Debug(fmt.Sprintf("Entro a READ, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, direccion))
	soliLectura := structs.Lectura{
		PID: global.Proceso_Ejecutando.PID,
		DirFisica: direccion,
		Tamanio: tamanio,
	}
	body, err := json.Marshal(soliLectura)
	if err != nil{
		global.CpuLogger.Error(fmt.Sprintf("Error al serializar: %s", err.Error()))
		return
	}
	urlMemoria := fmt.Sprintf("http://%s:%d/leer", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	respEnvio, errEnviar := http.Post(urlMemoria, "application/json", bytes.NewBuffer(body))
	if errEnviar != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al enviar solicitud de lectura a memoria: %s", errEnviar.Error()))
		return
	}
	global.CpuLogger.Debug(fmt.Sprintf("Envie READ a memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, direccion))
	defer respEnvio.Body.Close()
	// me traje los datos:
	global.CpuLogger.Debug(fmt.Sprintf("Llego READ de memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, direccion))
	var datosLeidos []byte
	errLectura := json.NewDecoder(respEnvio.Body).Decode(&datosLeidos)
	if errLectura != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al decodificar la lectura: %s", errLectura.Error()))
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: %d, - Accion: LEER, Direccion fisica: %d, Valor Leido: %s", global.Proceso_Ejecutando.PID, direccion, string(datosLeidos)))
}
