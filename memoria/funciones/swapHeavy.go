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

func RecuperarProcesoDeSwap(pidBuscado int) (structs.ProcesoMemoria, error) {
	swapMutex.Lock()
	defer swapMutex.Unlock()

	// 1) Obtener lista de marcos reservados al PID
	var marcos []int
	for idx, p := range global.MapMemoriaDeUsuario {
		if p == pidBuscado {
			marcos = append(marcos, idx)
		}
	}
	if len(marcos) == 0 {
		return structs.ProcesoMemoria{}, fmt.Errorf("no hay marcos para PID %d", pidBuscado)
	}
	frameSize := int(global.MemoriaConfig.PageSize)

	// 2) Leer todo swap
	data, err := os.ReadFile(global.MemoriaConfig.SwapPath)
	if err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("error leyendo swap: %w", err)
	}

	var filtered []byte
	i := 0
	for i < len(data) {
		// leer PID
		if i+4 > len(data) {
			break
		}
		pid := int(int32(binary.LittleEndian.Uint32(data[i : i+4])))
		i += 4
		// contar páginas segun snapshot (map) para este pid
		count := 0
		for _, q := range global.MapMemoriaDeUsuario {
			if q == pid {
				count++
			}
		}
		blockSize := count * frameSize
		if i+blockSize > len(data) {
			return structs.ProcesoMemoria{}, fmt.Errorf("bloque corrupto PID %d", pid)
		}
		block := data[i : i+blockSize]
		// restaurar si es el buscado
		if pid == pidBuscado {
			for j, m := range marcos {
				start := j * frameSize
				copy(global.MemoriaUsuario[m*frameSize:(m+1)*frameSize], block[start:start+frameSize])
			}
		} else {
			// conservar en filtered
			var buf [4]byte
			binary.LittleEndian.PutUint32(buf[:], uint32(pid))
			filtered = append(filtered, buf[:]...)
			filtered = append(filtered, block...)
		}
		i += blockSize
	}

	// 3) Reescribir swap sin PID recuperado
	if err := os.WriteFile(global.MemoriaConfig.SwapPath, filtered, 0666); err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("error reescribiendo swap: %w", err)
	}

	// no debería llegar aquí pues existe en global.Procesos
	return structs.ProcesoMemoria{}, fmt.Errorf("PID %d no encontrado en memoria principal tras recuperar", pidBuscado)
}

// EscribirMarcoEnSwap escribe los bytes de un marco específico en el archivo swap.bin.
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
