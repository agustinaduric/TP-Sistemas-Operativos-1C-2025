package fmemoria

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//SI este archivo anda es un milagro

var swapMutex sync.Mutex

func GuardarProcesoEnSwap(nuevo structs.ProcesoMemoria) error {
	swapMutex.Lock()
	// 1) Buscar la tabla multinivel para el PID
	var procTP *structs.ProcesoTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == nuevo.PID {
			procTP = &global.ProcesosTP[i]
			break
		}
	}
	if procTP == nil {
		return fmt.Errorf("GuardarProcesoEnSwap: PID %d no encontrado en ProcesosTP", nuevo.PID)
	}

	// 2) Abrir (o crear) el archivo swap
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir swap: %w", err)
	}
	defer f.Close()

	// 3) Escribir el PID (int32 little endian)
	if err := binary.Write(f, binary.LittleEndian, int32(nuevo.PID)); err != nil {
		return fmt.Errorf("error escribiendo PID en swap: %w", err)
	}

	// 4) Función recursiva para recorrer niveles hasta hoja y volcar páginas
	frameSize := int(global.MemoriaConfig.PageSize)
	var dumpNivel func(tabla []structs.Tp) error
	dumpNivel = func(tabla []structs.Tp) error {
		for _, entrada := range tabla {
			if !entrada.EsUltimoNivel {
				// descendemos
				if err := dumpNivel(entrada.TablaSiguienteNivel); err != nil {
					return err
				}
			} else {
				// nivel hoja: cada ptr apunta al inicio de una página
				for _, ptr := range entrada.ByteMemUsuario {
					if ptr == nil {
						continue
					}
					// convertir *byte a slice de longitud frameSize
					datos := unsafe.Slice(ptr, frameSize)
					// escribir contenido de página
					if _, err := f.Write(datos); err != nil {
						return fmt.Errorf("error escribiendo página en swap: %w", err)
					}
				}
			}
		}
		return nil
	}

	// 5) Volcar todas las páginas
	if err := dumpNivel(procTP.TablaNivel1); err != nil {
		return err
	}
	swapMutex.Unlock()
	return nil
}

func RecuperarProcesoDeSwap(pidBuscado int) (structs.ProcesoMemoria, error) {
	swapMutex.Lock()

	path := global.MemoriaConfig.SwapPath
	data, err := os.ReadFile(path)
	if err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("error leyendo swap: %w", err)
	}

	var proc structs.ProcesoMemoria
	var out []byte
	i := 0
	for i < len(data) {
		if i+8 > len(data) {
			break // datos corruptos
		}
		pid := int(int32(binary.LittleEndian.Uint32(data[i : i+4])))
		tam := int(int32(binary.LittleEndian.Uint32(data[i+4 : i+8])))
		numPages := (tam + int(global.MemoriaConfig.PageSize) - 1) / int(global.MemoriaConfig.PageSize)
		start := i + 8
		end := start + numPages*int(global.MemoriaConfig.PageSize)
		if end > len(data) {
			return structs.ProcesoMemoria{}, fmt.Errorf("datos de PID %d incompletos", pid)
		}

		block := data[i:end]
		if pid == pidBuscado {
			// reconstruir ProcesoMemoria
			proc.PID = pid
			proc.Tamanio = tam
			proc.EnSwap = false
			// podrías llenar instrucciones, métricas, etc., según convenga
			procData := make([]byte, len(block)-8)
			copy(procData, block[8:])
			proc.Path = ""
			proc.Instrucciones = nil
			proc.Metricas = structs.MetricasMemoria{}
			// asignar MemoriaUsuario:
			procMem := procData

		} else {
			// conservar en salida
			out = append(out, data[i:end]...)
		}
		i = end
	}

	// reescribir swap con los procesos que no coinciden
	if err := os.WriteFile(path, out, 0666); err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("error reescribiendo swap: %w", err)
	}

	if proc.PID == 0 {
		return proc, fmt.Errorf("PID %d no encontrado en swap", pidBuscado)
	}
	swapMutex.Unlock()
	return proc, nil
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
	//LiberarPaginasProcesoTP(pid)

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
	//asignarMarcosAProcesoTPPorPID(pid)
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
