package fmemoria

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func GuardarProcesoEnSwap(pid int) error {
	global.MemoriaMutex.Lock()
	marcos := RecolectarMarcos(pid)
	global.MemoriaMutex.Unlock()

	pageCount := len(marcos)

	// 1. Escribir header
	if err := EScribirStringIntEnSwap("PID: ", pid); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	if err := EScribirStringIntEnSwap("Marcos: ", pageCount); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}

	// 2. Asegurar línea vacía si no hay marcos
	if pageCount == 0 {
		f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		return nil
	}

	// 3. Escribir cada marco
	for idx, marco := range marcos {
		// Línea tipo "Marco1:"
		f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		if _, err := f.WriteString(fmt.Sprintf("Marco%d:\n", idx+1)); err != nil {
			f.Close()
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		f.Close()

		// Contenido binario del marco
		if err := EscribirMarcoEnSwap(marco); err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: error escribiendo marco %d: %w", idx+1, err)
		}

		// Asegurar salto de línea tras cada marco
		f, err = os.OpenFile(global.MemoriaConfig.SwapPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		if _, err := f.WriteString("\n"); err != nil {
			f.Close()
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		f.Close()
	}

	LiberarMarcos(pid)
	return nil
}

// EscribirMarcoEnSwap escribe el contenido de un marco en swap, userMutex protege MemoriaUsuario
func EscribirMarcoEnSwap(marco int) error {
	pageSize := int(global.MemoriaConfig.PageSize)

	// rango
	global.MemoriaMutex.Lock()
	memLen := len(global.MemoriaUsuario)
	global.MemoriaMutex.Unlock()
	if marco < 0 || marco*pageSize+pageSize > memLen {
		return fmt.Errorf("EscribirMarcoEnSwap: marco %d fuera de rango", marco)
	}

	// copiar datos
	buf := make([]byte, pageSize)
	global.MemoriaMutex.Lock()
	copy(buf, global.MemoriaUsuario[marco*pageSize:(marco+1)*pageSize])
	global.MemoriaMutex.Unlock()

	// volcar datos
	for _, b := range buf {
		if err := EscribirByteEnSwap(b); err != nil {
			return fmt.Errorf("EscribirMarcoEnSwap: %w", err)
		}
	}
	return nil
}

// EScribirStringIntEnSwap no requiere lock de swapMutex porque se llama dentro
func EScribirStringIntEnSwap(prefijo string, valor int) error {
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("AgregarStringIntEnSwap: no pudo abrir %s: %w", global.MemoriaConfig.SwapPath, err)
	}
	defer f.Close()

	línea := fmt.Sprintf("%s%d\n", prefijo, valor)
	if _, err := f.WriteString(línea); err != nil {
		return fmt.Errorf("AgregarStringIntEnSwap: escritura fallida en %s: %w", global.MemoriaConfig.SwapPath, err)
	}

	return nil
}

// EscribirByteEnSwap toma swapMutex por byte (aunque idealmente no) con logs
func EscribirByteEnSwap(b byte) error {
	global.MemoriaLogger.Debug("Entre en [EscribirByteEnSwap]")

	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return fmt.Errorf("EscribirByteEnSwap: no pudo abrir %s: %w", global.MemoriaConfig.SwapPath, err)
	}
	defer f.Close()

	n, err := f.Write([]byte{b})
	if err != nil {
		return fmt.Errorf("EscribirByteEnSwap: error al escribir byte: %w", err)
	}
	if n != 1 {
		return fmt.Errorf("EscribirByteEnSwap: bytes escritos inesperados: %d", n)
	}
	return nil
}
