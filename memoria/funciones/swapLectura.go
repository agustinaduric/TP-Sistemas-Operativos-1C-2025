package fmemoria

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// swapMutex sincroniza acceso concurrente al archivo swap durante lectura/escritura de bloques.
var swapMutex sync.Mutex

func RecuperarProcesoDeSwap(pidBuscado int) (structs.ProcesoMemoria, error) {
	// 1) Buscar índice del proceso en global.Procesos para actualizar EnSwap
	global.MemoriaMutex.Lock()
	var procIdx int = -1
	for i, pm := range global.Procesos {
		if pm.PID == pidBuscado {
			procIdx = i
			break
		}
	}
	global.MemoriaMutex.Unlock()
	if procIdx < 0 {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: PID=%d no existe en memoria principal", pidBuscado)
	}

	// 2) Snapshot de marcos reservados para el PID
	global.MemoriaMutex.Lock()
	var marcos []int
	for frameIdx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pidBuscado {
			marcos = append(marcos, frameIdx)
		}
	}
	global.MemoriaMutex.Unlock()
	if len(marcos) == 0 {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: PID=%d no tiene marcos reservados en memoria principal", pidBuscado)
	}
	pageCountExpected := len(marcos)
	pageSize := int(global.MemoriaConfig.PageSize)

	// 3) Abrir swapfile y temp bajo swapMutex
	swapMutex.Lock()
	defer swapMutex.Unlock()

	swapPath := global.MemoriaConfig.SwapPath
	f, err := os.Open(swapPath)
	if err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: no se puede abrir swapfile: %w", err)
	}
	defer f.Close()

	// Crear temp file en la misma carpeta
	dir, file := filepath.Split(swapPath)
	tmpName := filepath.Join(dir, file+".tmp")
	tf, err := os.Create(tmpName)
	if err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: no se puede crear temp file: %w", err)
	}
	// En caso de error posterior, eliminar temp
	defer func() {
		tf.Close()
		if err != nil {
			os.Remove(tmpName)
		}
	}()

	// 4) Leer secuencialmente bloques del swap
	headerBuf := make([]byte, 8) // 4 bytes PID + 4 bytes PageCount
	var offset int64 = 0
	// idxMarco para copiar páginas en orden de marcos
	idxMarco := 0

	for {
		// Leer header
		n, errRead := f.ReadAt(headerBuf, offset)
		if errRead == io.EOF && n == 0 {
			// fin de archivo
			break
		}
		if errRead != nil && errRead != io.EOF {
			return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: error leyendo header en offset %d: %w", offset, errRead)
		}
		if n < 8 {
			return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: header incompleto en offset %d", offset)
		}
		// Parsear PID y PageCount
		blockPID := int(int32(binary.LittleEndian.Uint32(headerBuf[0:4])))
		pageCount := int(binary.LittleEndian.Uint32(headerBuf[4:8]))
		// Validar pageCount no negativo
		if pageCount < 0 {
			return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: pageCount negativo %d en offset %d", pageCount, offset)
		}
		// Calcular tamaño de datos
		blockSize := int64(pageCount) * int64(pageSize)
		offset += 8

		if blockPID == pidBuscado {
			// Validar cantidad esperada
			if pageCount != pageCountExpected {
				return structs.ProcesoMemoria{}, fmt.Errorf(
					"RecuperarProcesoEnSwap: número de páginas en swap (%d) distinto de marcos reservados (%d) para PID %d",
					pageCount, pageCountExpected, pidBuscado)
			}
			// Copiar cada página a memoria principal
			for j := 0; j < pageCount; j++ {
				// Leer página j
				pageBuf := make([]byte, pageSize)
				n2, err2 := f.ReadAt(pageBuf, offset+int64(j*pageSize))
				if err2 != nil && err2 != io.EOF {
					return structs.ProcesoMemoria{}, fmt.Errorf(
						"RecuperarProcesoEnSwap: error leyendo página %d de PID %d en offset %d: %w",
						j, pidBuscado, offset+int64(j*pageSize), err2)
				}
				if n2 < pageSize {
					return structs.ProcesoMemoria{}, fmt.Errorf(
						"RecuperarProcesoEnSwap: lectura incompleta de página %d de PID %d", j, pidBuscado)
				}
				// Obtener marco reservado
				if idxMarco >= len(marcos) {
					return structs.ProcesoMemoria{}, fmt.Errorf(
						"RecuperarProcesoEnSwap: más páginas en swap (%d) que marcos reservados (%d) para PID %d",
						pageCount, len(marcos), pidBuscado)
				}
				frame := marcos[idxMarco]
				// Copiar a memoria bajo lock
				global.MemoriaMutex.Lock()
				start := frame * pageSize
				copy(global.MemoriaUsuario[start:start+pageSize], pageBuf)
				// Verificar consistencia de MapMemoriaDeUsuario
				if global.MapMemoriaDeUsuario[frame] != pidBuscado {
					global.MemoriaLogger.Error(fmt.Sprintf(
						"RecuperarProcesoEnSwap: inconsistencia en marco reservado: frame=%d no marcado para PID=%d",
						frame, pidBuscado))
					global.MapMemoriaDeUsuario[frame] = pidBuscado
				}
				global.MemoriaMutex.Unlock()
				idxMarco++
			}
			// No escribimos este bloque en temp: se elimina del swap.
		} else {
			// PID distinto: copiar header + datos al temp
			// Escribir header
			if _, err := tf.Write(headerBuf); err != nil {
				return structs.ProcesoMemoria{}, fmt.Errorf(
					"RecuperarProcesoEnSwap: error escribiendo header PID %d en temp: %w", blockPID, err)
			}
			// Copiar datos en streaming de pageSize en pageSize
			bytesLeft := blockSize
			buf := make([]byte, pageSize)
			curOffset := offset
			for bytesLeft > 0 {
				toRead := pageSize
				if int64(toRead) > bytesLeft {
					toRead = int(bytesLeft)
				}
				n2, err2 := f.ReadAt(buf[:toRead], curOffset)
				if err2 != nil && err2 != io.EOF {
					return structs.ProcesoMemoria{}, fmt.Errorf(
						"RecuperarProcesoEnSwap: error leyendo datos PID %d en offset %d: %w",
						blockPID, curOffset, err2)
				}
				if n2 < toRead {
					return structs.ProcesoMemoria{}, fmt.Errorf(
						"RecuperarProcesoEnSwap: lectura incompleta datos PID %d en offset %d",
						blockPID, curOffset)
				}
				if _, err := tf.Write(buf[:toRead]); err != nil {
					return structs.ProcesoMemoria{}, fmt.Errorf(
						"RecuperarProcesoEnSwap: error escribiendo datos PID %d en temp: %w", blockPID, err)
				}
				curOffset += int64(toRead)
				bytesLeft -= int64(toRead)
			}
		}
		// Avanzar offset al siguiente bloque
		offset += blockSize
	}

	// 5) Cerrar y reemplazar swapfile
	if err := tf.Sync(); err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: error en Sync temp file: %w", err)
	}
	if err := tf.Close(); err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: error cerrando temp file: %w", err)
	}
	f.Close()

	// Renombrar temp sobre swapfile original
	if err := os.Rename(tmpName, swapPath); err != nil {
		return structs.ProcesoMemoria{}, fmt.Errorf("RecuperarProcesoEnSwap: no se pudo renombrar temp a swapfile: %w", err)
	}

	// 6) Marcar EnSwap = false en global.Procesos
	global.MemoriaMutex.Lock()
	global.Procesos[procIdx].EnSwap = false
	procCopy := global.Procesos[procIdx]
	global.MemoriaMutex.Unlock()

	return procCopy, nil
}
