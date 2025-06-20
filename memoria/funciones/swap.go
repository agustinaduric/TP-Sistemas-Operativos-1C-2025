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
	//TODO
	swapMutex.Unlock()
	return nil
}

func RecuperarProcesoDeSwap(pidBuscado int) (structs.ProcesoMemoria, error) {
	swapMutex.Lock()
	//TODO
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

	if err := DesuspenderProceso(pid); err != nil { //ACA LLAMA A DESUSPENDE 
		global.MemoriaLogger.Error(fmt.Sprintf("  DesuspenderProceso falló PID=%d: %s", pid, err))
		return err
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("PedidoDeDesSuspension: PID=%d des-suspendido con éxito", pid))
	return nil
}

func SuspenderProceso(pid int) error {

	IncrementarBajadasSwap(pid)
	LiberarMarcos(pid) //esto es en la mockeada, osea el mapmemoriadeusuario, aunque sea la mockeada es super canonica
	LiberarPaginasProcesoTP(pid)

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

	IncrementarSubidasMem(pid)
	asignarMarcosAProcesoTPPorPID(pid)
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

func LiberarPaginasProcesoTP(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"LiberarPaginasProcesoTP: inicio PID=%d", pid,
	))

	encontrado := false
	for idx := range global.ProcesosTP {
		if global.ProcesosTP[idx].PID == pid {
			encontrado = true

			// Recorrer cada tabla y cada página para resetear a -1
			for t := range global.ProcesosTP[idx].TPS {
				for p := range global.ProcesosTP[idx].TPS[t].Paginas {
					global.ProcesosTP[idx].TPS[t].Paginas[p] = -1
					global.MemoriaLogger.Debug(fmt.Sprintf(
						"  PID=%d Tabla[%d].Paginas[%d] liberada (ahora = -1)",
						pid, t, p,
					))
				}
			}

			global.MemoriaLogger.Debug(fmt.Sprintf(
				"LiberarPaginasProcesoTP: fin PID=%d, todas las páginas liberadas",
				pid,
			))
			break
		}
	}

	if !encontrado {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"LiberarPaginasProcesoTP: PID=%d no encontrado en ProcesosTP",
			pid,
		))
	}
}

func asignarMarcosAProcesoTPPorPID(pid int) {
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"asignarMarcosAProcesoTPPorPID: inicio PID=%d", pid,
	))

	// 1. Buscar el ProcesoTP
	var procTP *structs.ProcesoTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == pid {
			procTP = &global.ProcesosTP[i]
			break
		}
	}
	if procTP == nil {
		global.MemoriaLogger.Error(fmt.Sprintf(
			"asignarMarcosAProcesoTPPorPID: PID=%d no encontrado en ProcesosTP",
			pid,
		))
		return
	}

	// 2. Configurar índices de tabla y página
	tablaIdx, paginaIdx := 0, 0
	numTablas := len(procTP.TPS)
	numPaginas := 0
	if numTablas > 0 {
		numPaginas = len(procTP.TPS[0].Paginas)
	}

	global.MarcosMutex.Lock()
	defer global.MarcosMutex.Unlock()

	// 3. Recorrer MapMemoriaDeUsuario y asignar
	for marco, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante != pid {
			continue
		}

		procTP.TPS[tablaIdx].Paginas[paginaIdx] = marco
		global.MemoriaLogger.Debug(fmt.Sprintf(
			"  PID=%d asigna marco %d a Tabla[%d].Paginas[%d]",
			pid, marco, tablaIdx, paginaIdx,
		))

		// Avanzar índices
		paginaIdx++
		if paginaIdx >= numPaginas {
			paginaIdx = 0
			tablaIdx++
			if tablaIdx >= numTablas {
				break
			}
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"asignarMarcosAProcesoTPPorPID: fin PID=%d", pid,
	))
}
