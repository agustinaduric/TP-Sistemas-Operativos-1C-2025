package fmemoria

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// RecuperarProcesoDeSwap restaura del swap el proceso pidBuscado:
//  1. extrae su bloque de swap,
//  2. lee cuántas páginas tiene,
//  3. lee esas páginas (bytes crudos),
//  4. las escribe de vuelta en MemoriaUsuario usando MapMemoriaDeUsuario,
//  5. y elimina su bloque del archivo swap.bin.
func RecuperarProcesoDeSwap(pidBuscado int) error {
	global.MemoriaLogger.Debug(fmt.Sprintf("RecuperarProcesoDeSwap: inicio PID=%d", pidBuscado))

	// 1) Leer todo el swap
	contenido, err := os.ReadFile(global.MemoriaConfig.SwapPath)
	if err != nil {
		return fmt.Errorf("no puedo leer swap: %w", err)
	}

	// 2) Partir en bloques y quedarnos con el que importa
	bloques, resto, err := dividirBloques(contenido, pidBuscado)
	if err != nil {
		return err
	}
	if len(bloques) == 0 {
		return fmt.Errorf("PID=%d no encontrado en swap", pidBuscado)
	}
	bloque := bloques[0]

	// 3) Parsear cantidad de páginas y datos de esas páginas
	paginaCount, datosPaginas, err := parsearBloque(bloque)
	if err != nil {
		return err
	}
	global.MemoriaLogger.Debug(fmt.Sprintf("  PID=%d tiene %d páginas en swap", pidBuscado, paginaCount))

	// 4) Recolectar los marcos asignados antes al PID
	marcos := RecolectarMarcos(pidBuscado)
	if len(marcos) < paginaCount {
		return fmt.Errorf("no hay suficientes marcos para PID=%d: necesito %d, tengo %d",
			pidBuscado, paginaCount, len(marcos))
	}

	// 5) Escribir cada página de vuelta en MemoriaUsuario
	pageSize := int(global.MemoriaConfig.PageSize)
	for i := 0; i < paginaCount; i++ {
		inicio := i * pageSize
		fin := inicio + pageSize
		chunk := datosPaginas[inicio:fin]
		marco := marcos[i]
		offset := marco * pageSize

		global.MemoriaMutex.Lock()
		copy(global.MemoriaUsuario[offset:offset+pageSize], chunk)
		global.MemoriaMutex.Unlock()

		global.MemoriaLogger.Debug(fmt.Sprintf(
			"  PID=%d: restaurada página %d en marco %d (bytes %d–%d)",
			pidBuscado, i, marco, inicio, fin,
		))
	}

	// 6) Sobrescribir swap.bin sin el bloque recuperado
	if err := os.WriteFile(global.MemoriaConfig.SwapPath, resto, 0664); err != nil {
		return fmt.Errorf("no puedo actualizar swap: %w", err)
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("RecuperarProcesoDeSwap: fin PID=%d", pidBuscado))
	return nil
}

// dividirBloques separa data en bloques por "PID: " y devuelve:
//   - bloques que comienzan con el PID buscado,
//   - resto del archivo sin esos bloques.
func dividirBloques(data []byte, pid int) ([][]byte, []byte, error) {
	partes := bytes.Split(data, []byte("PID: "))
	var bloques [][]byte
	var bufResto bytes.Buffer

	for i, parte := range partes {
		if i == 0 {
			bufResto.Write(parte)
			continue
		}
		encabezado := fmt.Sprintf("%d\n", pid)
		if bytes.HasPrefix(parte, []byte(encabezado)) {
			bloque, rem, err := extraerBloque(parte)
			if err != nil {
				return nil, nil, err
			}
			bloques = append(bloques, bloque)
			bufResto.Write(rem)
		} else {
			bufResto.Write([]byte("PID: "))
			bufResto.Write(parte)
		}
	}
	return bloques, bufResto.Bytes(), nil
}

// extraerBloque lee desde parte:
//   - línea PID,
//   - línea "Cantidad Paginas: N",
//   - N * pageSize bytes,
//
// devuelve el bloque completo y el resto de bytes posteriores.
func extraerBloque(parte []byte) ([]byte, []byte, error) {
	lector := bufio.NewReader(bytes.NewReader(parte))
	// descartar línea PID
	_, err := lector.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}
	// leer Cantidad Paginas
	linea, err := lector.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}
	campos := strings.Fields(linea)
	if len(campos) != 2 {
		return nil, nil, fmt.Errorf("formato inválido: %q", linea)
	}
	N, err := strconv.Atoi(campos[1])
	if err != nil {
		return nil, nil, err
	}
	pageSize := int(global.MemoriaConfig.PageSize)
	total := N * pageSize

	// leer datos de páginas
	datos := make([]byte, total)
	if _, err := io.ReadFull(lector, datos); err != nil {
		return nil, nil, err
	}
	// calcular longitud del header
	consumed := len(parte) - lector.Buffered()
	bloque := parte[:consumed]
	resto := parte[consumed:]
	return bloque, resto, nil
}

// parsearBloque recibe un bloque completo y retorna N y los bytes de páginas
func parsearBloque(bloque []byte) (int, []byte, error) {
	lector := bufio.NewReader(bytes.NewReader(bloque))
	// descartar línea PID
	if _, err := lector.ReadString('\n'); err != nil {
		return 0, nil, err
	}
	// leer Cant Paginas
	linea, err := lector.ReadString('\n')
	if err != nil {
		return 0, nil, err
	}
	campos := strings.Fields(linea)
	N, err := strconv.Atoi(campos[1])
	if err != nil {
		return 0, nil, err
	}
	// resto son los bytes
	datos, err := io.ReadAll(lector)
	return N, datos, err
}
