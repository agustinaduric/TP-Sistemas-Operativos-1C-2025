package fmemoria

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

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
