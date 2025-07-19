package fmemoria

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// GuardarProcesoEnSwap escribe en el swap un bloque para el proceso con formato:
//
//		[4 bytes PID]
//	 [4 bytes Cantidad de paginas]
//	 [Cantidad de paginas * Tamaño paginas bytes de datos de páginas]
func GuardarProcesoEnSwap(pid int) error {
	swapMutex.Lock()
	defer swapMutex.Unlock()
	global.MemoriaMutex.Lock()
	marcos := RecolectarMarcos(pid)
	global.MemoriaMutex.Unlock()

	//if len(marcos) == 0 {
	//	return fmt.Errorf("GuardarProcesoEnSwap: PID %d no tiene marcos reservados en memoria principal", pid)
	//}

	pageCount := len(marcos)

	// ) Escribir header: PID y cantidad de páginas
	if err := EScribirStringIntEnSwap("PID: ", pid); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	if err := EScribirStringIntEnSwap("Cantidad Paginas: ", pageCount); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	if pageCount > 0 {
		for idx, marco := range marcos {
			if err := EscribirMarcoEnSwap(marco); err != nil {
				// Registramos el error, pero continuamos con las demás páginas
				fmt.Fprintf(os.Stderr, "GuardarProcesoEnSwap: error escribiendo página %d (marco %d) para PID %d: %v\n",
					idx, marco, pid, err)
			}
		}
	}
	// ) Escribir cada página en orden de marcos (orden ascendente de índice de marco)

	return nil
}

// EscribirMarcoEnSwap escribe el contenido de un marco específico en el swap.
func EscribirMarcoEnSwap(marco int) error { //llama varias veces a escribir byte
	pageSize := int(global.MemoriaConfig.PageSize)

	// validar rango
	global.MemoriaMutex.Lock()
	memLen := len(global.MemoriaUsuario)
	global.MemoriaMutex.Unlock()
	if marco < 0 || marco*pageSize+pageSize > memLen {
		return fmt.Errorf("EscribirMarcoEnSwap: marco %d fuera de rango", marco)
	}

	// copiar el contenido del marco bajo lock
	buf := make([]byte, pageSize)
	global.MemoriaMutex.Lock()
	copy(buf, global.MemoriaUsuario[marco*pageSize:(marco+1)*pageSize])
	global.MemoriaMutex.Unlock()

	// volcar cada byte al swap
	for _, b := range buf {
		if err := EscribirByteEnSwap(b); err != nil {
			return fmt.Errorf("EscribirMarcoEnSwap: %w", err)
		}
	}
	return nil
}

func EScribirStringIntEnSwap(prefijo string, valor int) error {
	// Abrir o crear el archivo en modo append
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("AgregarStringIntEnSwap: no pudo abrir %s: %w", global.MemoriaConfig.SwapPath, err)
	}
	defer f.Close()

	// Formatear la línea: primero el prefijo, luego el número y un salto de línea
	línea := fmt.Sprintf("%s%d\n", prefijo, valor)

	// Escribirla al final del archivo
	if _, err := f.WriteString(línea); err != nil {
		return fmt.Errorf("AgregarStringIntEnSwap: escritura fallida en %s: %w", global.MemoriaConfig.SwapPath, err)
	}

	return nil
}

func EscribirByteEnSwap(b byte) error {
	swapMutex.Lock()
	swapMutex.Unlock()

	// Abrimos o creamos el archivo en modo append
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return fmt.Errorf("EscribirByteEnSwap: no pudo abrir %s: %w",
			global.MemoriaConfig.SwapPath, err)
	}
	defer f.Close()

	// Escribimos el byte
	n, err := f.Write([]byte{b})
	if err != nil {
		return fmt.Errorf("EscribirByteEnSwap: error al escribir byte: %w", err)
	}
	if n != 1 {
		return fmt.Errorf("EscribirByteEnSwap: bytes escritos inesperados: %d", n)
	}

	return nil
}
