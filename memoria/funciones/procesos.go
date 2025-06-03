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
	// 1. Debug inicial con PID
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Buscando Proceso: PID=%d", pid),
	)

	for i, proc := range global.Procesos {
		if proc.PID == pid {
			global.MemoriaLogger.Debug(
				fmt.Sprintf("Proceso encontrado: PID=%d, Índice=%d", pid, i),
			)
			return proc, true
		}
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Proceso con PID=%d no encontrado", pid),
	)
	return structs.ProcesoMemoria{}, false
}

func BuscarInstruccion(pid int, pc int) (structs.Instruccion, error) {
	global.MemoriaLogger.Debug(fmt.Sprintf("Buscando Instrucción: PID=%d, PC=%d", pid, pc))

	proceso, existe := BuscarProceso(pid)
	if !existe {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error al buscar instrucción en memoria: proceso con PID %d no existe", pid),
		)
		return structs.Instruccion{}, fmt.Errorf("BuscarInstruccion: el pid %d no existe", pid)
	}

	if pc < 0 || pc >= len(proceso.Instrucciones) {

		global.MemoriaLogger.Error(
			fmt.Sprintf(
				"Error al buscar instrucción en memoria: PC %d fuera de rango [0,%d) para PID %d",
				pc, len(proceso.Instrucciones), pid,
			),
		)
		return structs.Instruccion{}, fmt.Errorf(
			"BuscarInstruccion: pc %d fuera de rango [0,%d) para pid %d",
			pc, len(proceso.Instrucciones), pid,
		)
	}

	instr := proceso.Instrucciones[pc]
	global.MemoriaLogger.Info(
		fmt.Sprintf(
			"## PID: %d - Obtener instrucción: %d - Operación: %s - Argumentos: %v",
			pid, pc, instr.Operacion, instr.Argumentos,
		),
	)

	return instr, nil
}

func CargarInstrucciones(ruta string) ([]structs.Instruccion, error) {
	global.MemoriaLogger.Debug(
		fmt.Sprintf("Cargando instrucciones desde ruta: %s", ruta),
	)

	archivo, err := os.Open(ruta)
	if err != nil {
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error al abrir archivo de instrucciones '%s': %s", ruta, err.Error()),
		)
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
		global.MemoriaLogger.Error(
			fmt.Sprintf("Error leyendo líneas de '%s': %s", ruta, err.Error()),
		)
		return nil, err
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("Total de instrucciones cargadas desde '%s': %d", ruta, len(instrucciones)),
	)

	return instrucciones, nil
}

//ESTA FUNCION VA A TENER LLAMADOS A FUNCIONES QUE VAN A HACER ESTO
//  Verifica espacio en MemoriaUsuario
//  Reserva paginas 
//  Carga instrucciones del archivo en memoria interna A CHEKCEAR
// Agrega ProcesoMemoria a la lista interna de procesos
//func CrearProceso(pid int, size int, instruccionesPath string) error

// SuspenderProceso realiza la lógica de suspensión de un proceso:
// - Serializa páginas modificadas a swapfile.bin (actualiza SwapOffsets) ¿?
// - Libera los marcos en MemoriaUsuario
// - Marca EnSwap = true y actualiza métricas.BajadasASwap
//func SuspenderProceso(pid int) error

// ReanudarProceso realiza la lógica de cancelar suspensión:
// - Lee páginas desde swapfile.bin según SwapOffsets ¿?
// - Reasigna marcos en MemoriaUsuario y libera entradas de swap
// - Marca EnSwap = false y actualiza métricas.SubidasDeSwap
//func ReanudarProceso(pid int) error

// FinalizarProceso libera todos los recursos de un proceso:
// - Libera marcos en MemoriaUsuario y marca SwapOffsets como libres
// - Emite log con MetricasMemoria y elimina entrada en lista interna
//func FinalizarProceso(pid int) error

ETSA SI QUE A CHEKAER
// TraducirDireccion convierte una dirección lógica a física:
// - Dado pid y direcciónLógica, calcula page y offset
// - Recorre N niveles de TablaRaiz (incrementa métricas.AccesosTabla)
// - Retorna direcciónFisica = frame*PageSize + offset o error si falla
//func TraducirDireccion(pid int, direccionLogica int) (int, error)

// LeerMemoria realiza lectura de bloque en espacio de usuario:
// - Traduce direcciónLógica a física con TraducirDireccion
// - Copia 'size' bytes desde MemoriaUsuario[direccionFisica:]
// - Incrementa métricas.LecturasUsuario
// - Retorna el slice de bytes o error
// - Nota: Si estas leyendo esto es xq te da miedo la palabra byte, trankilo
func LeerMemoria(pid int, direccionLogica int, size int) ([]byte, error)

// EscribirMemoria realiza escritura de bloque en espacio de usuario:
// - Traduce direcciónLógica a física con TraducirDireccion
// - Copia 'data' a MemoriaUsuario[direccionFisica:]
// - Incrementa métricas.EscriturasUsuario
// - Retorna nil o error si fuera de límites
//func EscribirMemoria(pid int, direccionLogica int, data []byte) error

// LeerPagina devuelve el contenido de una página completa dado un frame:
// - Calcula offset = frame * PageSize
// - Retorna MemoriaUsuario[offset:offset+PageSize] o error si fuera de límites
// - Incrementa métricas.LecturasUsuario
//func LeerPagina(frame int) ([]byte, error)

// EscribirPagina escribe una página completa dado un frame:
// - Calcula offset = frame * PageSize
// - Escribe 'data' (de tamaño PageSize) en MemoriaUsuario[offset:]
// - Incrementa métricas.EscriturasUsuario
// - Retorna nil o error si data no coincide con PageSize
//func EscribirPagina(frame int, data []byte) error




