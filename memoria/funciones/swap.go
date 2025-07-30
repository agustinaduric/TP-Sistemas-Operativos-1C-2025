package fmemoria

import (
	"fmt"
	"os"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

var swapMutex sync.Mutex

// PedidoDeDesSuspension: no toma directamente el mutex, delega en DesuspenderProceso
func PedidoDeDesSuspension(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("PedidoDeDesSuspension: inicio PID=%d", pid))

	proc, ok := BuscarProceso(pid)
	if !ok {
		global.MemoriaLogger.Error(fmt.Sprintf("  PID=%d no existe en memoria principal", pid))
		return fmt.Errorf("PID=%d no encontrado", pid)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("  proceso hallado Tamanio=%d", proc.Tamanio))

	if !hayEspacio(proc.Tamanio) {
		global.MemoriaLogger.Debug(fmt.Sprintf("  espacio insuficiente para PID=%d (necesita %d bytes)", pid, proc.Tamanio))
		return fmt.Errorf("espacio insuficiente para PID=%d", pid)
	}
	global.MemoriaLogger.Debug("  espacio disponible, procediendo a DesuspenderProceso")

	return DesuspenderProceso(pid)
}

// SuspenderProceso: toma swapMutex con logs de lock/unlock
func SuspenderProceso(pid int) error {
	global.MemoriaLogger.Debug("[SuspenderProceso] intentando tomar swapMutex")
	swapMutex.Lock()
	global.MemoriaLogger.Debug("[SuspenderProceso] tomó swapMutex")
	defer func() {
		swapMutex.Unlock()
		global.MemoriaLogger.Debug("[SuspenderProceso] liberó swapMutex")
	}()

	IncrementarBajadasSwap(pid)
	global.MemoriaLogger.Debug(fmt.Sprintf("SuspenderProceso: inicio PID=%d", pid))

	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].EnSwap = true
			global.MemoriaLogger.Debug(fmt.Sprintf("  marcado EnSwap, BajadasSwap=%d", global.Procesos[i].Metricas.BajadasSwap))

			if err := GuardarProcesoEnSwap(pid); err != nil {
				global.MemoriaLogger.Error(fmt.Sprintf("  error GuardarProcesoEnSwap: %s", err))
				return err
			}
			global.MemoriaLogger.Debug("SuspenderProceso: proceso guardado en swap.bin")
			//VisualizadorDeSwap()
			return nil
		}
	}

	global.MemoriaLogger.Error(fmt.Sprintf("SuspenderProceso: PID=%d no encontrado", pid))
	return fmt.Errorf("PID=%d no encontrado", pid)
}

// DesuspenderProceso: toma swapMutex con logs de lock/unlock
func DesuspenderProceso(pid int) error {
	global.MemoriaLogger.Debug("[DesuspenderProceso] intentando tomar swapMutex")
	swapMutex.Lock()
	global.MemoriaLogger.Debug("[DesuspenderProceso] tomó swapMutex")
	defer func() {
		swapMutex.Unlock()
		global.MemoriaLogger.Debug("[DesuspenderProceso] liberó swapMutex")
	}()

	IncrementarSubidasMem(pid)
	OcuparMarcos(pid)
	InicializarProcesoTP(pid)
	global.MemoriaLogger.Debug(fmt.Sprintf("DesuspenderProceso: inicio PID=%d", pid))

	err := RecuperarProcesoDeSwap(pid)
	if err != nil {
		global.MemoriaLogger.Error(fmt.Sprintf("  RecuperarProcesoDeSwap falló: %s", err))
		return err
	}
	global.MemoriaLogger.Debug("  proceso recuperado de swap.bin")

	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].EnSwap = false
			global.MemoriaLogger.Debug(fmt.Sprintf("  actualizado EnSwap=false, SubidasMem=%d",
				global.Procesos[i].Metricas.SubidasMem))
			return nil
		}
	}

	global.MemoriaLogger.Debug("  proceso agregado a memoria principal tras des-suspensión")
	return nil
}

func LimpiarSwap() {
	path := global.MemoriaConfig.SwapPath
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {

			global.MemoriaLogger.Debug("LimpiarSwap: no había swap previo, nada que borrar")
		} else {
			global.MemoriaLogger.Debug(fmt.Sprintf("LimpiarSwap: error borrando '%s': %v", path, err))
		}
	} else {
		global.MemoriaLogger.Debug(fmt.Sprintf("Swap Previo ELiminado"))
	}
}
