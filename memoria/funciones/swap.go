package fmemoria

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

//SI este archivo anda es un milagr

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
	//LiberarPaginasProcesoTP(pid)

	global.MemoriaLogger.Debug(fmt.Sprintf("SuspenderProceso: inicio PID=%d", pid))

	for i := range global.Procesos {
		if global.Procesos[i].PID == pid {
			global.Procesos[i].EnSwap = true
			global.Procesos[i].Metricas.BajadasSwap++
			global.MemoriaLogger.Debug(fmt.Sprintf("  marcado EnSwap, BajadasSwap=%d", global.Procesos[i].Metricas.BajadasSwap))

			if err := GuardarProcesoEnSwap(pid); err != nil {
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
	OcuparMarcos(pid) //esto es en la mockeada, osea el mapmemoriadeusuario, aunque sea la mockeada es super canonica
	AsignarMarcosAProcesoTPPorPID(pid)
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
			global.Procesos[i].Metricas.SubidasMem++
			global.MemoriaLogger.Debug(fmt.Sprintf("  actualizado EnSwap=false, SubidasMem=%d", global.Procesos[i].Metricas.SubidasMem))
			return nil
		}
	}
	global.MemoriaLogger.Debug("  proceso agregado a memoria principal tras des-suspensión")
	return nil
}

// recolectarMarcos devuelve la lista de índices de marcos ocupados por pid
func RecolectarMarcos(pid int) []int {
	var marcos []int
	for idx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, idx)
		}
	}
	return marcos
}
