package fmemoria

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func BuscarProceso(pid int) (structs.ProcesoMemoria, bool) {
	for i := 0; i < len(global.Procesos); i++ {
		if global.Procesos[i].PID == pid {
			return global.Procesos[i], true
		}
	}
	return structs.ProcesoMemoria{}, false
}

func BuscarInstruccion(pid int, pc int) (structs.Instruccion, error) {
	proceso, existe := BuscarProceso(pid)
	if !existe {
		return structs.Instruccion{}, fmt.Errorf("BuscarInstruccion: el pid %d no existe", pid)
	}
	if pc < 0 || pc >= len(proceso.Instrucciones) {
		return structs.Instruccion{}, fmt.Errorf("BuscarInstruccion: pc %d fuera de rango [0,%d) para pid %d",
			pc, len(proceso.Instrucciones), pid)
	}
	return proceso.Instrucciones[pc], nil
}

func CargarInstrucciones(ruta string) ([]structs.Instruccion, error) {
	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer archivo.Close()

	var instrucciones []structs.Instruccion
	escáner := bufio.NewScanner(archivo)
	for escáner.Scan() {
		línea := strings.TrimSpace(escáner.Text())
		if línea == "" {
			continue
		}
		partes := strings.Fields(línea)
		instrucciones = append(instrucciones, structs.Instruccion{
			Operacion:  partes[0],
			Argumentos: partes[1:],
		})
	}
	if err := escáner.Err(); err != nil {
		return nil, err
	}
	return instrucciones, nil
}
