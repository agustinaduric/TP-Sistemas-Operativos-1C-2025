package fmemoria

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

// RecuperarProcesoDeSwap busca en swap el bloque para pid, restaura páginas y elimina el bloque
func RecuperarProcesoDeSwap(pid int) error {

	global.MemoriaMutex.Lock()
	defer global.MemoriaMutex.Unlock()
	// 1) Leer bloque completo
	bloque, err := leerBloqueSwap(pid)
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

	// 5) Eliminar bloque de swap
	if err := eliminarBloqueSwap(pid); err != nil {
		return fmt.Errorf("RecuperarProcesoDeSwap: %w", err)
	}

	return nil
}

func leerBloqueSwap(pid int) ([]byte, error) {
	path := global.MemoriaConfig.SwapPath
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("leerBloqueSwap: no pudo abrir %s: %w", path, err)
	}
	defer f.Close()

	marker := fmt.Sprintf("PID: %d", pid)
	reader := bufio.NewReader(f)

	// 1) Avanzar hasta encontrar la línea "PID: <pid>"
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("leerBloqueSwap: PID=%d no encontrado", pid)
			}
			return nil, fmt.Errorf("leerBloqueSwap: lectura fallida: %w", err)
		}
		if strings.HasPrefix(line, marker) {
			break
		}
	}

	// 2) Desde aquí, leer TODO el contenido hasta el inicio del siguiente bloque
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("PID: %d\n", pid))

	// Leer línea por línea para mejor control
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Si llegamos al final del archivo, agregamos lo que tengamos
				if len(line) > 0 {
					buf.WriteString(line)
				}
				break
			}
			return nil, fmt.Errorf("leerBloqueSwap: error leyendo bloque: %w", err)
		}

		// Si encontramos el inicio de otro bloque PID, paramos ANTES de agregarlo
		if strings.HasPrefix(line, "PID: ") && !strings.HasPrefix(line, marker) || (err == io.EOF) {
			break
		}

		buf.WriteString(line)
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
	global.MemoriaLogger.Debug(fmt.Sprintf("[parsearBloque] RAW (%d bytes): %q", len(bloque), bloque))
	// 1) Separamos hasta 3 trozos por "\n"
	parts := bytes.SplitN(bloque, []byte("\n"), 3)
	global.MemoriaLogger.Debug(fmt.Sprintf("[parsearBloque] partes encontradas: %d", len(parts)))
	for i, p := range parts {
		global.MemoriaLogger.Debug(fmt.Sprintf("[parsearBloque] parts[%d]=%q", i, p))
	}
	// Verificación mínima
	if len(parts) < 2 {
		return 0, nil, fmt.Errorf("parsearBloque: formato inválido, %d líneas", len(parts))
	}

	// 2) Extraemos el count de la cabecera "Cantidad Paginas: N"
	header := strings.TrimSpace(string(parts[1]))
	const pref = "Cantidad Paginas:"
	if !strings.HasPrefix(header, pref) {
		return 0, nil, fmt.Errorf("parsearBloque: cabecera inválida '%s'", header)
	}
	numStr := strings.TrimSpace(strings.TrimPrefix(header, pref))
	cant, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, nil, fmt.Errorf("parsearBloque: count parse error: %w", err)
	}
	if cant == 0 {
		return 0, nil, nil
	}

	// 3) Determinamos los datos binarios:
	//    a) Si SplitN devolvió 3 partes, el tercer elemento es todo el data[].
	//    b) Si sólo devolvió 2, buscamos manualmente donde acaba la 2ª línea.
	var data []byte
	if len(parts) >= 3 {
		data = parts[2]
	} else {
		// Ubicar el segundo '\n' en todo el bloque
		first := bytes.Index(bloque, []byte("\n"))
		if first < 0 {
			return 0, nil, fmt.Errorf("parsearBloque: no hay primer '\\n'")
		}
		second := bytes.Index(bloque[first+1:], []byte("\n"))
		if second < 0 {
			return 0, nil, fmt.Errorf("parsearBloque: no hay segundo '\\n'")
		}
		data = bloque[first+1+second+1:]
	}

	return cant, data, nil
}

// eliminarBloqueSwap remueve TODO el bloque del proceso pid, basándose en el prefijo "PID: <pid>"
func eliminarBloqueSwap(pid int) error {
	orig := global.MemoriaConfig.SwapPath
	tmp := filepath.Join(filepath.Dir(orig), "swap.tmp")

	in, err := os.Open(orig)
	if err != nil {
		return fmt.Errorf("eliminarBloqueSwap: %w", err)
	}
	defer in.Close()

	out, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("eliminarBloqueSwap: %w", err)
	}
	defer out.Close()

	scanner := bufio.NewReader(in)
	marker := fmt.Sprintf("PID: %d", pid)
	skipping := false

	for {
		line, err := scanner.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("eliminarBloqueSwap: %w", err)
		}

		if strings.HasPrefix(line, marker) {
			// empezamos a saltar el bloque completo
			skipping = true
			// saltamos esta línea también
		} else if skipping && strings.HasPrefix(line, "PID: ") {
			// encontramos el siguiente bloque, dejamos de saltar
			skipping = false
			// este nuevo bloque lo escribimos
			out.WriteString(line)
		} else if !skipping {
			out.WriteString(line)
		}

		if err == io.EOF {
			break
		}
	}

	in.Close()
	out.Close()
	if err := os.Rename(tmp, orig); err != nil {
		return fmt.Errorf("eliminarBloqueSwap: %w", err)
	}
	return nil
}
