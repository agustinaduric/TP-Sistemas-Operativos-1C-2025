package fmemoria

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// GuardarProcesoEnSwap escribe en el swap un bloque para el proceso con formato:
//
//	[4 bytes PID][4 bytes Cantidad de paginas][Cantidad de paginas * Tamaño paginas bytes de datos de páginas]
func GuardarProcesoEnSwap(pid int) error {

	global.MemoriaMutex.Lock()
	var marcos []int
	for frameIdx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, frameIdx)
		}
	}
	global.MemoriaMutex.Unlock()

	if len(marcos) == 0 {
		return fmt.Errorf("GuardarProcesoEnSwap: PID %d no tiene marcos reservados en memoria principal", pid)
	}

	pageCount := len(marcos)

	// ) Escribir header: PID y cantidad de páginas
	if err := EscribirPIDEnSwap(pid); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	if err := EscribirCantidadPaginasEnSwap(pageCount); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}

	// ) Escribir cada página en orden de marcos (orden ascendente de índice de marco)
	for idx, marco := range marcos {
		if err := EscribirMarcoEnSwap(marco); err != nil {
			// Registramos el error, pero continuamos con las demás páginas
			fmt.Fprintf(os.Stderr, "GuardarProcesoEnSwap: error escribiendo página %d (marco %d) para PID %d: %v\n",
				idx, marco, pid, err)
		}
	}

	return nil
}

// EscribirPIDEnSwap escribe 4 bytes LittleEndian con el PID al archivo de swap.
func EscribirPIDEnSwap(pid int) error {
	swapMutex.Lock()
	defer swapMutex.Unlock()

	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("EscribirPIDEnSwap: error abriendo swapfile: %w", err)
	}
	defer f.Close()

	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(pid))
	if _, err := f.Write(buf[:]); err != nil {
		return fmt.Errorf("EscribirPIDEnSwap: error escribiendo PID %d: %w", pid, err)
	}
	return nil
}

// EscribirCantidadPaginasEnSwap escribe 4 bytes LittleEndian con pageCount al archivo de swap.
func EscribirCantidadPaginasEnSwap(pageCount int) error {
	swapMutex.Lock()
	defer swapMutex.Unlock()

	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("EscribirCantidadPaginasEnSwap: error abriendo swapfile: %w", err)
	}
	defer f.Close()

	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(pageCount))
	if _, err := f.Write(buf[:]); err != nil {
		return fmt.Errorf("EscribirCantidadPaginasEnSwap: error escribiendo pageCount=%d: %w", pageCount, err)
	}
	return nil
}

// EscribirMarcoEnSwap escribe el contenido de un marco específico en el swap.
func EscribirMarcoEnSwap(marco int) error {
	pageSize := int(global.MemoriaConfig.PageSize)

	// Validar rango del marco
	global.MemoriaMutex.Lock()
	memLen := len(global.MemoriaUsuario)
	global.MemoriaMutex.Unlock()
	if marco < 0 || (marco*pageSize+pageSize) > memLen {
		return fmt.Errorf("EscribirMarcoEnSwap: marco %d fuera de rango", marco)
	}

	// Leer datos de MemoriaUsuario bajo lock
	pageBuf := make([]byte, pageSize)
	global.MemoriaMutex.Lock()
	copy(pageBuf, global.MemoriaUsuario[marco*pageSize:(marco+1)*pageSize])
	global.MemoriaMutex.Unlock()

	// Escribir al swapfile bajo lock
	swapMutex.Lock()
	defer swapMutex.Unlock()

	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("EscribirMarcoEnSwap: error abriendo swapfile: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(pageBuf); err != nil {
		return fmt.Errorf("EscribirMarcoEnSwap: error escribiendo marco %d: %w", marco, err)
	}
	return nil
}
