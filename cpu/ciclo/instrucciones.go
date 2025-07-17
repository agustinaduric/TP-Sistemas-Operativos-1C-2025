package ciclo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	mmu "github.com/sisoputnfrba/tp-golang/cpu/MMU"
	"github.com/sisoputnfrba/tp-golang/cpu/cache"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func WRITE(dirLogica int, datos string) {
	global.CpuLogger.Debug(fmt.Sprintf("Entro a WRITE, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirLogica))
	pagina := dirLogica / global.Page_size // ver responsabilidad

	//cache
	if global.EntradasMaxCache > 0 {
		cache.EscribirEnCache(global.Proceso_Ejecutando.PID, pagina, []byte(datos))
		global.CpuLogger.Debug(fmt.Sprintf("PID: %d hizo write en cache, Pagina: %d - Dato: %s", global.Proceso_Ejecutando.PID, pagina, datos))
		return
	}

	//memoria
	dirFisica := mmu.DL_a_DF(dirLogica)

	soliEscritura := structs.Escritura{
		PID:       global.Proceso_Ejecutando.PID,
		DirFisica: dirFisica,
		Datos:     []byte(datos),
	}
	body, err := json.Marshal(soliEscritura)
	if err != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al serializar: %s", err.Error()))
		return
	}
	urlMemoria := fmt.Sprintf("http://%s:%d/escribir", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	respEnvio, errEnviar := http.Post(urlMemoria, "application/json", bytes.NewBuffer(body))
	if errEnviar != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al enviar solicitud escritura a memoria: %s", errEnviar.Error()))
		return
	}
	global.CpuLogger.Debug(fmt.Sprintf("Envie WRITE a memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirFisica))
	defer respEnvio.Body.Close()
	// me responde memoria:
	global.CpuLogger.Debug(fmt.Sprintf("Me respondio el WRITE memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirFisica))
	if respEnvio.StatusCode != 200 {
		global.CpuLogger.Error(fmt.Sprintf("Memoria devolvio error en WRITE: %d", respEnvio.StatusCode))
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: %d, - Accion: ESCRIBIR en MEMORIA PRINCIPAL, Direccion fisica: %d, Valor Escrito: %s", global.Proceso_Ejecutando.PID, dirFisica, datos))
}

func READ(dirLogica int, tamanio int) {
	global.CpuLogger.Debug(fmt.Sprintf("Entro a READ, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirLogica))
	pagina := dirLogica / global.Page_size // ver responsabilidad

	//cache
	if global.EntradasMaxCache > 0 {
		hayEnCache, dato := cache.BuscarEncache(global.Proceso_Ejecutando.PID, pagina)
		if hayEnCache { // hit
			global.CpuLogger.Info(fmt.Sprintf("PID: %d - Acción: LEER desde CACHE - Dirección Física: %d - Valor: %s", global.Proceso_Ejecutando.PID, dirLogica, string([]byte{dato})))
			return
		}
	}

	dirFisica := mmu.DL_a_DF(dirLogica)

	//memoria
	soliLectura := structs.Lectura{
		PID:       global.Proceso_Ejecutando.PID,
		DirFisica: dirFisica,
		Tamanio:   tamanio,
	}
	body, err := json.Marshal(soliLectura)
	if err != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al serializar: %s", err.Error()))
		return
	}
	urlMemoria := fmt.Sprintf("http://%s:%d/leer", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	respEnvio, errEnviar := http.Post(urlMemoria, "application/json", bytes.NewBuffer(body))
	if errEnviar != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al enviar solicitud de lectura a memoria: %s", errEnviar.Error()))
		return
	}
	global.CpuLogger.Debug(fmt.Sprintf("Envie READ a memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirFisica))
	defer respEnvio.Body.Close()

	// me traje los datos:
	global.CpuLogger.Debug(fmt.Sprintf("Llego READ de memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirFisica))
	var datosLeidos []byte
	errLectura := json.NewDecoder(respEnvio.Body).Decode(&datosLeidos)
	if errLectura != nil {
		global.CpuLogger.Error(fmt.Sprintf("Error al decodificar la lectura: %s", errLectura.Error()))
	}
	if global.EntradasMaxCache > 0 {
		cache.EscribirEnCache(global.Proceso_Ejecutando.PID, pagina, datosLeidos)
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: %d, - Accion: LEER desde MEMORIA PRINCIPAL, Direccion fisica: %d, Valor Leido: %s", global.Proceso_Ejecutando.PID, dirFisica, string(datosLeidos)))
}
