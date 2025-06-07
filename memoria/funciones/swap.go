package fmemoria

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//SI este archivo anda es un milagro

const rutaSwap = "swap.bin"

var swapMutex sync.Mutex

func GuardarProcesoEnSwap(nuevo structs.ProcesoMemoria) error {
	swapMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("GuardarProcesoEnSwap: entrada PID=%d", nuevo.PID))

	// 1. Leer lista actual
	var lista []structs.ProcesoMemoria
	if _, err := os.Stat(rutaSwap); err == nil {
		global.MemoriaLogger.Debug("  swap.bin existe, leyéndolo")
		datos, err := ioutil.ReadFile(rutaSwap)
		if err != nil {
			global.MemoriaLogger.Error(fmt.Sprintf("  error leyendo %s: %s", rutaSwap, err))
			return fmt.Errorf("error leyendo %s: %w", rutaSwap, err)
		}
		if len(datos) > 0 {
			if err := json.Unmarshal(datos, &lista); err != nil {
				global.MemoriaLogger.Error(fmt.Sprintf("  error unmarshalling JSON: %s", err))
				return fmt.Errorf("error unmarshalling JSON: %w", err)
			}
		}
	}

	// 2. Agregar nuevo
	lista = append(lista, nuevo)
	global.MemoriaLogger.Debug(fmt.Sprintf("  proceso añadido a lista, total=%d", len(lista)))

	// 3. Serializar y escribir
	bytesSalida, err := json.MarshalIndent(lista, "", "  ")
	if err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  error marshaling JSON: %s", err))
		return fmt.Errorf("error haciendo Marshal de lista: %w", err)
	}
	if err := ioutil.WriteFile(rutaSwap, bytesSalida, 0644); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  error escribiendo %s: %s", rutaSwap, err))
		return fmt.Errorf("error escribiendo %s: %w", rutaSwap, err)
	}

	global.MemoriaLogger.Debug("GuardarProcesoEnSwap: éxito al escribir swap.bin")
	swapMutex.Unlock()
	return nil
}

func RecuperarProcesoDeSwap(pidBuscado int) (structs.ProcesoMemoria, error) {
	swapMutex.Lock()
	global.MemoriaLogger.Debug(fmt.Sprintf("RecuperarProcesoDeSwap: entrada PID=%d", pidBuscado))

	// Leer archivo
	datos, err := ioutil.ReadFile(rutaSwap) //veo esta oracion tachada, si ustedes tambien la ven asi pongan un numero aleatorio en este comentario :
	if err != nil {
		if os.IsNotExist(err) {
			global.MemoriaLogger.Error(fmt.Sprintf("  %s no existe", rutaSwap))
			return structs.ProcesoMemoria{}, fmt.Errorf("%s no existe", rutaSwap)
		}
		global.MemoriaLogger.Error(fmt.Sprintf("  error leyendo %s: %s", rutaSwap, err))
		return structs.ProcesoMemoria{}, fmt.Errorf("error leyendo %s: %w", rutaSwap, err)
	}
	if len(datos) == 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("  %s está vacío", rutaSwap))
		return structs.ProcesoMemoria{}, fmt.Errorf("%s está vacío", rutaSwap)
	}
	global.MemoriaLogger.Debug("  archivo swap.bin leído correctamente")

	// Unmarshal
	var lista []structs.ProcesoMemoria
	if err := json.Unmarshal(datos, &lista); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  error unmarshalling JSON: %s", err))
		return structs.ProcesoMemoria{}, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	// Buscar
	indice := -1
	for i, proc := range lista {
		if proc.PID == pidBuscado {
			indice = i
			break
		}
	}
	if indice < 0 {
		global.MemoriaLogger.Error(fmt.Sprintf("  PID=%d no encontrado en lista", pidBuscado))
		return structs.ProcesoMemoria{}, fmt.Errorf("PID=%d no encontrado en swap", pidBuscado)
	}
	procRecuperado := lista[indice]
	global.MemoriaLogger.Debug(fmt.Sprintf("  PID=%d encontrado en índice %d", pidBuscado, indice))

	// Eliminar de la lista
	lista = append(lista[:indice], lista[indice+1:]...)
	global.MemoriaLogger.Debug(fmt.Sprintf("  PID=%d eliminado de lista, quedan %d procesos", pidBuscado, len(lista)))

	// Reescribir swap.bin
	bytesSalida, err := json.MarshalIndent(lista, "", "  ")
	if err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  error marshaling JSON actualizado: %s", err))
		return structs.ProcesoMemoria{}, fmt.Errorf("error marshaling lista actualizada: %w", err)
	}
	if err := ioutil.WriteFile(rutaSwap, bytesSalida, 0644); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  error escribiendo %s: %s", rutaSwap, err))
		return structs.ProcesoMemoria{}, fmt.Errorf("error reescribiendo %s: %w", rutaSwap, err)
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("RecuperarProcesoDeSwap: PID=%d recuperado y swap.bin actualizado", pidBuscado))
	swapMutex.Unlock()

	return procRecuperado, nil
}

func PedidoDeDesSuspension(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("PedidoDeDesSuspension: inicio PID=%d", pid))

	proc, ok := BuscarProceso(pid)
	if !ok {
		global.MemoriaLogger.Error(fmt.Sprintf("  PID=%d no existe en memoria principal", pid))
		return fmt.Errorf("PID=%d no encontrado", pid)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("  proceso hallado Tamanio=%d", proc.Tamanio))

	if !hayEspacio(proc.Tamanio) {
		global.MemoriaLogger.Error(fmt.Sprintf("  espacio insuficiente para PID=%d (necesita %d bytes)", pid, proc.Tamanio))
		return fmt.Errorf("espacio insuficiente para PID=%d", pid)
	}
	global.MemoriaLogger.Debug("  espacio disponible, procediendo a DesuspenderProceso")

	if err := DesuspenderProceso(pid); err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  DesuspenderProceso falló PID=%d: %s", pid, err))
		return err
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("PedidoDeDesSuspension: PID=%d des-suspendido con éxito", pid))
	return nil
}

func SuspenderProceso(pid int) error {

	LiberarMarcos(pid) //esto es en la mockeada, osea el mapmemoriadeusuario, aunque sea la mockeada es super canonica
	//TODO
	//Liberarmarcos en la memoria  jerarquica

	global.MemoriaLogger.Debug(fmt.Sprintf("SuspenderProceso: inicio PID=%d", pid))

	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].EnSwap = true
			global.Procesos[i].Metricas.BajadasSwap++
			global.MemoriaLogger.Debug(fmt.Sprintf("  marcado EnSwap, BajadasSwap=%d", global.Procesos[i].Metricas.BajadasSwap))

			if err := GuardarProcesoEnSwap(global.Procesos[i]); err != nil {
				global.MemoriaLogger.Error(fmt.Sprintf("  error GuardarProcesoEnSwap: %s", err))
				return err
			}
			global.MemoriaLogger.Debug("SuspenderProceso: proceso guardado en swap.bin")
			return nil
		}
	}

	global.MemoriaLogger.Error(fmt.Sprintf("SuspenderProceso: PID=%d no encontrado", pid))
	return fmt.Errorf("PID=%d no encontrado", pid)
}

func DesuspenderProceso(pid int) error {

	//TODO
	//marcar marcos en  la jerarquica
	OcuparMarcos(pid) //esto es en la mockeada, osea el mapmemoriadeusuario, aunque sea la mockeada es super canonica

	global.MemoriaLogger.Debug(fmt.Sprintf("DesuspenderProceso: inicio PID=%d", pid))

	proc, err := RecuperarProcesoDeSwap(pid)
	if err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  RecuperarProcesoDeSwap falló: %s", err))
		return err
	}
	global.MemoriaLogger.Debug("  proceso recuperado de swap.bin")

	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].EnSwap = false
			global.Procesos[i].Metricas.SubidasMem++
			global.MemoriaLogger.Debug(fmt.Sprintf("  actualizado EnSwap=false, SubidasMem=%d", global.Procesos[i].Metricas.SubidasMem))
			return nil
		}
	}

	// Si no existía en memoria principal, lo agrego:
	proc.EnSwap = false
	proc.Metricas.SubidasMem++
	global.Procesos = append(global.Procesos, proc)
	global.MemoriaLogger.Debug("  proceso agregado a memoria principal tras des-suspensión")
	return nil
}
