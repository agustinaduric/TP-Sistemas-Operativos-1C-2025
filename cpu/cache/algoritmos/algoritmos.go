package algoritmoCache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mmu "github.com/sisoputnfrba/tp-golang/cpu/MMU"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Clock(entrada structs.EntradaCache) {
	global.CpuLogger.Debug("Entro a clock")
	for i := 0; i < len(global.CachePaginas); i++ {
		entradaActual := &global.CachePaginas[global.PunteroClock]
		if !entradaActual.BitUso {
			if entradaActual.BitModificado {
				global.CpuLogger.Debug(fmt.Sprintf("Entrada modificada, se la mando a memoria antes de reemplazar"))
				EnviarEscribirAMemoria(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, entradaActual)
				global.CpuLogger.Debug(fmt.Sprintf("Se envio la entrada a memoria antes de reemplazar"))
			}
			global.CpuLogger.Debug(fmt.Sprintf("Reemplazo CLOCK en posicion: %d", global.PunteroClock))
			global.CachePaginas[global.PunteroClock] = entrada
			global.CpuLogger.Debug(fmt.Sprintf("ya reemplace la entrada"))
			avanzarPuntero()
			return
		}
		entradaActual.BitUso = false
		global.CpuLogger.Debug("Bit de uso pasa a 0 y continuo")
		avanzarPuntero()
	}
}

func ClockM(entrada structs.EntradaCache) {
	global.CpuLogger.Debug("Entro a clock-m")
	// primera vuelta (U = 0 y M = 0)
	for i := 0; i < len(global.CachePaginas); i++ {
		entradaActual := &global.CachePaginas[global.PunteroClock]
		if !entradaActual.BitUso && !entradaActual.BitModificado {
			global.CachePaginas[global.PunteroClock] = entrada
			global.CpuLogger.Debug(fmt.Sprintf("Reemplazo, U= 0 y M=0"))
			avanzarPuntero()
			return
		}
		entradaActual.BitUso = false
		avanzarPuntero()
	}
	//segunda vuelta
	for i := 0; i < len(global.CachePaginas); i++ {
		entradaActual := &global.CachePaginas[global.PunteroClock]
		if !entradaActual.BitUso && entradaActual.BitModificado {
			global.CpuLogger.Debug(fmt.Sprintf("Entrada modificada U=0 M=1, se la mando a memoria antes de reemplazar"))
			EnviarEscribirAMemoria(global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory, entradaActual)
			global.CpuLogger.Debug(fmt.Sprintf("Se envio la entrada a memoria antes de reemplazar"))
			global.CachePaginas[global.PunteroClock] = entrada
			global.CpuLogger.Debug(fmt.Sprintf("ya reemplace la entrada"))
			avanzarPuntero()
			return
		}
		entradaActual.BitUso = false
		global.CpuLogger.Debug("Bit de uso pasa a 0 y continuo")
		avanzarPuntero()
	}
	global.CpuLogger.Error(fmt.Sprintf("NO hay victima para clock-m"))
}

func avanzarPuntero() {
	global.PunteroClock++
	if global.PunteroClock >= global.EntradasMaxCache {
		global.PunteroClock = 0
		global.CpuLogger.Debug("El puntero de clock volvio a cero")
	}
	global.CpuLogger.Debug("El puntero de clock avanzo")
}

func EnviarEscribirAMemoria(ip string, puerto int, entrada *structs.EntradaCache) {
	desplazamiento := 0
	marco := mmu.ObtenerMarco(entrada.Pagina)
	dirFisica := marco*global.Page_size + desplazamiento

	escritura := structs.Escritura{
		PID:       entrada.PID,
		DirFisica: dirFisica,
		Datos:     entrada.Contenido,
	}
	body, err := json.Marshal(escritura)
	if err != nil {
		log.Printf("error codificando UPDATE: %s", err.Error())
		return
	}
	url := fmt.Sprintf("http://%s:%d/escribir", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando UPDATE de memoria: %s", err.Error())
		return
	}
	global.CpuLogger.Debug(fmt.Sprintf("Envie UPDATE a memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirFisica))
	defer resp.Body.Close()
	// me responde memoria:
	global.CpuLogger.Debug(fmt.Sprintf("Me respondio el UPDATE memoria, PID: %d, Direccion: %d", global.Proceso_Ejecutando.PID, dirFisica))
	if resp.StatusCode != 200 {
		log.Printf("Memoria respondio con error en Memory Update: %d", resp.StatusCode)
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: %d - Memory Update - Pagina: %d - Marco: %d", entrada.PID, entrada.Pagina, marco))
}
