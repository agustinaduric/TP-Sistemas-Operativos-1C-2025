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

// RecuperarProcesoDeSwap busca en swap el bloque para pid, restaura páginas y elimina el bloque
func RecuperarProcesoDeSwap(pid int) error {
	proc, _ := BuscarProceso(pid)
	saltos := proc.Metricas.BajadasSwap - 1

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()
	// 1) Leer bloque completo
	bloque, err := leerBloqueSwap(pid, saltos)
	if err != nil {
		return fmt.Errorf("RecuperarProcesoDeSwap: %w", err)
	}

	// 2) Parsear header + datos
	pageCount, pageBytes, err := parsearBloque(bloque)
	if err != nil {
		return fmt.Errorf("RecuperarProcesoDeSwap: %w", err)
	}

	// 3) Recolectar marcos locales
	marcos := RecolectarMarcos(pid)
	if len(marcos) < pageCount {
		return fmt.Errorf("RecuperarProcesoDeSwap: PID=%d esperaba %d marcos, encontró %d", pid, pageCount, len(marcos))
	}

	// 4) Restaurar datos en memoria principal bajo MemoriaMutex
	pageSize := int(global.MemoriaConfig.PageSize)
	for i := 0; i < pageCount; i++ {
		marco := marcos[i]
		inicio := marco * pageSize
		copy(global.MemoriaUsuario[inicio:inicio+pageSize], pageBytes[i*pageSize:(i+1)*pageSize])
	}

	return nil
}

func leerBloqueSwap(pid int, skip int) ([]byte, error) {
	f, err := os.Open(global.MemoriaConfig.SwapPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	marker := fmt.Sprintf("PID: %d", pid)

	// saltar skip apariciones
	for saltos := 0; saltos <= skip; {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("leerBloqueSwap: %w", err)
		}
		if strings.HasPrefix(line, marker) {
			saltos++
		}
	}

	// ahora reader está justo después de la cabecera de la (skip+1)-ésima aparición
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("PID: %d\n", pid))
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if strings.HasPrefix(line, "PID: ") {
			break
		}
		buf.WriteString(line)
		if err == io.EOF {
			break
		}
	}
	return buf.Bytes(), nil
}

// parsearBloque extrae la cantidad de marcos y el slice de bytes de datos según el nuevo formato:
//
//	PID: <pid>\n
//	Marcos: <N>\n
//	Marco1: <bytes> \n
//	...
func parsearBloque(bloque []byte) (int, []byte, error) {
	r := bufio.NewReader(bytes.NewReader(bloque))

	// 1) Saltar la línea "PID: <pid>\n"
	if _, err := r.ReadString('\n'); err != nil {
		return 0, nil, fmt.Errorf("parsearBloque: no pude leer cabecera PID: %w", err)
	}

	// 2) Leer y parsear "Cantidad Paginas:<N>\n"
	headerLine, err := r.ReadString('\n')
	if err != nil {
		return 0, nil, fmt.Errorf("parsearBloque: no pude leer cabecera Cantidad Paginas: %w", err)
	}
	headerLine = strings.TrimSpace(headerLine)
	const pref = "Cantidad Paginas:"
	if !strings.HasPrefix(headerLine, pref) {
		return 0, nil, fmt.Errorf("parsearBloque: cabecera inválida '%s'", headerLine)
	}
	numStr := strings.TrimSpace(strings.TrimPrefix(headerLine, pref))
	pageCount, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, nil, fmt.Errorf("parsearBloque: error al convertir pageCount: %w", err)
	}

	// 3) Saltar la línea vacía tras cabeceras (si existe)
	if next, _ := r.Peek(1); len(next) > 0 && next[0] == '\n' {
		r.ReadByte() // descarta '\n'
	}

	// 4) Leer cada bloque de marco:
	pageSize := int(global.MemoriaConfig.PageSize)
	total := pageCount * pageSize
	data := make([]byte, total)

	for i := 0; i < pageCount; i++ {
		// 4.1) Saltar línea "MarcoX:\n"
		if _, err := r.ReadString('\n'); err != nil {
			return 0, nil, fmt.Errorf("parsearBloque: no pude leer etiqueta Marco en página %d: %w", i, err)
		}
		// 4.2) Leer exactamente pageSize bytes
		off := i * pageSize
		if _, err := io.ReadFull(r, data[off:off+pageSize]); err != nil {
			return 0, nil, fmt.Errorf("parsearBloque: error leyendo datos de página %d: %w", i, err)
		}
		// 4.3) Saltar la línea de separación tras datos (un '\n')
		if sep, _ := r.Peek(1); len(sep) > 0 && sep[0] == '\n' {
			r.ReadByte()
		}
	}

	return pageCount, data, nil
}
