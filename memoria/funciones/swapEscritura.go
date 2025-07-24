package fmemoria

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func GuardarProcesoEnSwap(pid int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] SwapPath = %s", global.MemoriaConfig.SwapPath))
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Inicio PID=%d", pid))

	global.MemoriaMutex.Lock()
	marcos := RecolectarMarcos(pid)
	global.MemoriaMutex.Unlock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Marcos recolectados: %v", marcos))

	pageCount := len(marcos)
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Cantidad de marcos: %d", pageCount))

	// 1. Escribir header
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Escribiendo header PID"))
	if err := EScribirStringIntEnSwap("PID: ", pid); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Header PID escrito: PID: %d", pid))

	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Escribiendo header Marcos"))
	if err := EScribirStringIntEnSwap("Cantidad Paginas:", pageCount); err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Header Marcos escrito: Marcos: %d", pageCount))
	global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] Escribiendo salto de línea tras cabeceras")
	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		f.Close()
		return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
	}
	f.Close()
	// 2. Asegurar línea vacía si no hay marcos
	if pageCount == 0 {
		global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] No hay marcos, escribiendo línea vacía")
		f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] Línea vacía escrita")
		return nil
	}

	// 3. Escribir cada marco
	for idx, marco := range marcos {
		global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Comenzando a escribir marco %d (índice %d)", marco, idx))

		// Línea tipo "Marco1:"
		global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Escribiendo etiqueta Marco%d:", idx+1))
		f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		label := fmt.Sprintf("Marco%d:\n", idx+1)
		if _, err := f.WriteString(label); err != nil {
			f.Close()
			return fmt.Errorf("GuardarProcesoEnSwap: %w", err)
		}
		f.Close()
		global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Etiqueta escrita: %s", label))

		// Contenido binario del marco
		global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Escribiendo contenido binario del marco %d", marco))
		if err := EscribirMarcoEnSwap(marco); err != nil {
			return fmt.Errorf("GuardarProcesoEnSwap: error escribiendo marco %d: %w", idx+1, err)
		}
		global.MemoriaLogger.Debug(fmt.Sprintf("[GuardarProcesoEnSwap] Contenido del marco %d escrito", marco))

		// Asegurar salto de línea tras cada marco
		global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] Escribiendo salto de línea tras marco")
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
		global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] Salto de línea escrito")
	}

	global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] Liberando marcos en memoria principal")
	LiberarMarcos(pid)
	global.MemoriaLogger.Debug("[GuardarProcesoEnSwap] Finalizado")
	return nil
}

// EscribirMarcoEnSwap escribe el contenido de un marco en swap, userMutex protege MemoriaUsuario
func EscribirMarcoEnSwap(marco int) error {
	pageSize := int(global.MemoriaConfig.PageSize)

	global.MemoriaLogger.Debug(fmt.Sprintf("[EscribirMarcoEnSwap] Inicio marco=%d, pageSize=%d", marco, pageSize))

	// rango
	global.MemoriaMutex.Lock()
	memLen := len(global.MemoriaUsuario)
	global.MemoriaMutex.Unlock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[EscribirMarcoEnSwap] MemoriaUsuario len=%d", memLen))

	if marco < 0 || marco*pageSize+pageSize > memLen {
		return fmt.Errorf("EscribirMarcoEnSwap: marco %d fuera de rango", marco)
	}

	// copiar datos
	buf := make([]byte, pageSize)
	global.MemoriaMutex.Lock()
	copy(buf, global.MemoriaUsuario[marco*pageSize:(marco+1)*pageSize])
	global.MemoriaMutex.Unlock()
	global.MemoriaLogger.Debug(fmt.Sprintf("[EscribirMarcoEnSwap] Copiado buffer para marco %d: %v", marco, buf))

	// volcar datos
	for i, b := range buf {
		global.MemoriaLogger.Debug(fmt.Sprintf("[EscribirMarcoEnSwap] Escribiendo byte %d of marco %d: 0x%02X ('%c')", i, marco, b, b))
		if err := EscribirByteEnSwap(b); err != nil {
			return fmt.Errorf("EscribirMarcoEnSwap: %w", err)
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("[EscribirMarcoEnSwap] Fin marco=%d", marco))
	return nil
}

// EScribirStringIntEnSwap no requiere lock de swapMutex porque se llama dentro
func EScribirStringIntEnSwap(prefijo string, valor int) error {
	línea := fmt.Sprintf("%s%d\n", prefijo, valor)
	global.MemoriaLogger.Debug(fmt.Sprintf("[EScribirStringIntEnSwap] Preparando linea: %q", línea))

	f, err := os.OpenFile(global.MemoriaConfig.SwapPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("AgregarStringIntEnSwap: no pudo abrir %s: %w", global.MemoriaConfig.SwapPath, err)
	}
	defer f.Close()

	n, err := f.WriteString(línea)
	if err != nil {
		return fmt.Errorf("AgregarStringIntEnSwap: escritura fallida en %s: %w", global.MemoriaConfig.SwapPath, err)
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("[EScribirStringIntEnSwap] Escribió %d bytes: %q", n, línea))
	return nil
}

// EscribirByteEnSwap toma swapMutex por byte (aunque idealmente no) con logs
func EscribirByteEnSwap(b byte) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("[EscribirByteEnSwap] Entrando con byte: 0x%02X ('%c')", b, b))

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
	global.MemoriaLogger.Debug("[EscribirByteEnSwap] Byte escrito correctamente")
	return nil
}
