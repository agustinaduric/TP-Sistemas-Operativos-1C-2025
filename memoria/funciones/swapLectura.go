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
	global.MemoriaMutex.Lock()
	for i := 0; i < pageCount; i++ {
		marco := marcos[i]
		inicio := marco * pageSize
		copy(global.MemoriaUsuario[inicio:inicio+pageSize], pageBytes[i*pageSize:(i+1)*pageSize])
	}
	global.MemoriaMutex.Unlock()

	// 5) Eliminar bloque de swap
	if err := eliminarBloqueSwap(pid); err != nil {
		return fmt.Errorf("RecuperarProcesoDeSwap: %w", err)
	}

	return nil
}

// leerBloqueSwap abre swap.bin y devuelve el bloque de texto+bytes para PID
func leerBloqueSwap(pid int) ([]byte, error) {
	path := global.MemoriaConfig.SwapPath
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("leerBloqueSwap: no pudo abrir %s: %w", path, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	scanner := bufio.NewReader(f)
	marker := fmt.Sprintf("PID: %d", pid)
	found := false

	for {
		line, err := scanner.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("leerBloqueSwap: lectura fallida: %w", err)
		}
		if strings.HasPrefix(line, marker) {
			found = true
			buf.WriteString(line)
			break
		}
		if err == io.EOF {
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("leerBloqueSwap: PID=%d no encontrado", pid)
	}

	for {
		chunk := make([]byte, 4096)
		n, err := scanner.Read(chunk)
		if n > 0 {
			if idx := bytes.Index(chunk[:n], []byte("\nPID: ")); idx >= 0 {
				buf.Write(chunk[:idx])
				break
			}
			buf.Write(chunk[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("leerBloqueSwap: error leyendo bloque: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// parsearBloque extrae la cantidad de páginas y los bytes de datos de un bloque
func parsearBloque(bloque []byte) (int, []byte, error) {
	parts := bytes.SplitN(bloque, []byte("\n"), 3)
	if len(parts) < 2 {
		return 0, nil, fmt.Errorf("parsearBloque: formato inválido")
	}
	cant, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(string(parts[1]), "Cantidad Paginas:")))
	if err != nil {
		return 0, nil, fmt.Errorf("parsearBloque: count parse error: %w", err)
	}
	if cant == 0 {
		return 0, nil, nil
	}
	if len(parts) < 3 {
		return 0, nil, fmt.Errorf("parsearBloque: esperaba datos de páginas, no los encontré")
	}
	return cant, parts[2], nil
}

// eliminarBloqueSwap remueve el bloque completo de pid de swap.bin
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
			skipping = true
		}
		if !skipping {
			out.WriteString(line)
		}
		if skipping && strings.HasPrefix(line, "PID: ") {
			skipping = false
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
