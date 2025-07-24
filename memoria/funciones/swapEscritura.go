package fmemoria

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// GuardarProcesoEnSwap escribe en swap un bloque para el proceso con formato:
// [texto header] + [datos] todo bajo swapMutex con logs de lock/unlock
func GuardarProcesoEnSwap(pid int) error {
	global.MemoriaMutex.Lock()
	marcos := RecolectarMarcos(pid)
	global.MemoriaMutex.Unlock()

	pageCount := len(marcos)

	// Escribir header
	if err := EScribirStringIntEnSwap("PID: ", pid); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	if err := EScribirStringIntEnSwap("Cantidad Paginas: ", pageCount); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}

	// Escribir cada página
	for idx, marco := range marcos {
		err := EscribirMarcoEnSwap(marco)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"GuardarProcesoEnSwap: error escribiendo página %d (marco %d) para PID %d: %v\n",
				idx, marco, pid, err)
		}
	}

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
	global.MemoriaLogger.Debug("EscribirByteEnSwap: intentando tomar swapMutex")
	swapMutex.Lock()
	global.MemoriaLogger.Debug("EscribirByteEnSwap: tomó swapMutex")
	defer func() {
		swapMutex.Unlock()
		global.MemoriaLogger.Debug("EscribirByteEnSwap: liberó swapMutex")
	}()

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
