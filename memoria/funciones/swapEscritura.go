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

var swapMutex sync.Mutex

func GuardarProcesoEnSwap(nuevo structs.ProcesoMemoria) error {
	// 1) Validar existencia de tablas para el PID
	var procTP *structs.ProcesoTP
	for i := range global.ProcesosTP {
		if global.ProcesosTP[i].PID == nuevo.PID {
			procTP = &global.ProcesosTP[i]
			break
		}
	}
	if procTP == nil {
		return fmt.Errorf("GuardarProcesoEnSwap: PID %d no encontrado", nuevo.PID)
	}

	// 2) Escribir el PID al principio del bloque
	if err := EscribirPIDEnSwap(nuevo.PID); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}

	// 3) Recorrer la tabla multinivel e invocar EscribirMarcoEnSwap para cada hoja
	var dump func(tabla []structs.Tp)
	dump = func(tabla []structs.Tp) {
		for _, entrada := range tabla {
			if !entrada.EsUltimoNivel {
				dump(entrada.TablaSiguienteNivel)
			} else if entrada.NumeroMarco >= 0 {
				// Volcar el contenido del marco
				if err := EscribirMarcoEnSwap(entrada.NumeroMarco); err != nil {
					// No detenemos la escritura por error de un marco, pero lo registramos
					fmt.Fprintf(os.Stderr, "error GuardarProcesoEnSwap marco %d: %v\n", entrada.NumeroMarco, err)
				}
			}
		}
	}
	dump(procTP.TablaNivel1)

	return nil
}

func EscribirMarcoEnSwap(marco int) error {
	frameSize := int(global.MemoriaConfig.PageSize)
	if marco < 0 || marco >= len(global.MapMemoriaDeUsuario) {
		return fmt.Errorf("marco %d fuera de rango", marco)
	}

	swapMutex.Lock()
	defer swapMutex.Unlock()
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir swap: %w", err)
	}
	defer f.Close()

	offset := marco * frameSize
	datos := unsafe.Slice(&global.MemoriaUsuario[offset], frameSize)
	if _, err := f.Write(datos); err != nil {
		return fmt.Errorf("error escribiendo marco %d: %w", marco, err)
	}

	return nil
}

// EscribirPIDEnSwap escribe un entero (PID) en el archivo swap.bin.
func EscribirPIDEnSwap(pid int) error {
	swapMutex.Lock()
	defer swapMutex.Unlock()
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir swap: %w", err)
	}
	defer f.Close()

	if err := binary.Write(f, binary.LittleEndian, int32(pid)); err != nil {
		return fmt.Errorf("error escribiendo PID en swap: %w", err)
	}

	return nil
}
