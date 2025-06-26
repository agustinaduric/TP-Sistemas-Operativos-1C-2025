package fmemoria

import (
	"fmt"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var MarcosMutex sync.Mutex

func MarcosDisponibles() int {

	global.MemoriaLogger.Debug("MarcosDisponibles: contando marcos libres")

	var contador int
	for i := 0; i < CantidadMarcos(); i++ {
		if global.MapMemoriaDeUsuario[i] == -1 {
			contador++
		}
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosDisponibles: TotalDisponibles=%d", contador),
	)
	return contador
}

func MarcosNecesitados(tamanioProceso int) int {

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosNecesitados: TamanioProceso=%d", tamanioProceso),
	)

	var necesarios int
	if rem := tamanioProceso % global.MemoriaConfig.PageSize; rem == 0 {
		necesarios = tamanioProceso / global.MemoriaConfig.PageSize
	} else {
		necesarios = (tamanioProceso / global.MemoriaConfig.PageSize) + 1
	}

	global.MemoriaLogger.Debug(
		fmt.Sprintf("MarcosNecesitados: Necesarios=%d (para TamanioProceso=%d)", necesarios, tamanioProceso),
	)
	return necesarios
}

func CantidadMarcos() int {
	total := global.MemoriaConfig.MemorySize / global.MemoriaConfig.PageSize

	global.MemoriaLogger.Debug(
		fmt.Sprintf("CantidadMarcos: MemorySize=%d, PageSize=%d, TotalMarcos=%d",
			global.MemoriaConfig.MemorySize,
			global.MemoriaConfig.PageSize,
			total,
		),
	)
	return total
}

func LiberarMarcos(pid int) {
	MarcosMutex.Lock()
	defer MarcosMutex.Unlock()
	global.MemoriaLogger.Debug(fmt.Sprintf("LiberarMarcos: inicio PID=%d", pid))

	for marco, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			global.MapMemoriaDeUsuario[marco] = -1
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"LiberarMarcos: marco %d liberado (antes ocupado por PID=%d)",
				marco, pid,
			))
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf("LiberarMarcos: fin PID=%d", pid))

}

func OcuparMarcos(pid int) {
	MarcosMutex.Lock()
	defer MarcosMutex.Unlock()
	global.MemoriaLogger.Debug(fmt.Sprintf("OcuparMarcos: inicio PID=%d", pid))

	// 1. Obtener el proceso para conocer su tamaño
	proc, _ := BuscarProceso(pid) // asumimos que ya existe porque hay espacio
	necesarios := MarcosNecesitados(proc.Tamanio)
	global.MemoriaLogger.Debug(fmt.Sprintf(
		"OcuparMarcos: PID=%d requiere %d marcos (tamaño %d bytes)",
		pid, necesarios, proc.Tamanio,
	))

	// 2. Asignar exactamente 'necesarios' marcos libres
	asignados := 0

	for idx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == -1 {
			global.MapMemoriaDeUsuario[idx] = pid
			asignados++
			global.MemoriaLogger.Debug(fmt.Sprintf(
				"OcuparMarcos: PID=%d ocupa marco %d (%d/%d)",
				pid, idx, asignados, necesarios,
			))
			if asignados == necesarios {
				break
			}
		}
	}

	global.MemoriaLogger.Debug(fmt.Sprintf(
		"OcuparMarcos: fin PID=%d — se asignaron %d marcos",
		pid, asignados,
	))
}

// recolectarMarcos devuelve la lista de índices de marcos ocupados por pid
func RecolectarMarcos(pid int) []int {
	MarcosMutex.Lock()
	defer MarcosMutex.Unlock()
	var marcos []int
	for idx, ocupante := range global.MapMemoriaDeUsuario {
		if ocupante == pid {
			marcos = append(marcos, idx)
		}
	}
	return marcos
}

// Marco recorre la jerarquía según índices y devuelve el marco físico o -1.
func Marco(pid int, indices []int) int {
	MarcosMutex.Lock()
	defer MarcosMutex.Unlock()
	cfg := global.MemoriaConfig
	niveles := cfg.NumberOfLevels

	if len(indices) != niveles+1 {
		global.MemoriaLogger.Error(fmt.Sprintf("Marco: longitud de índices inválida %d, se espera %d", len(indices), niveles+1))
		return -1
	}
	procTP := getProcesoTP(pid)
	if procTP == nil {
		return -1
	}

	nivelActual := procTP.TablaNivel1
	var hoja structs.Tp
	for lvl := 1; lvl <= niveles; lvl++ {
		idx := indices[lvl-1]
		if idx < 0 || idx >= len(nivelActual) {
			global.MemoriaLogger.Error(fmt.Sprintf("Marco: índice fuera de rango en nivel %d: %d", lvl, idx))
			return -1
		}
		entrada := nivelActual[idx]
		if lvl < niveles {
			nivelActual = entrada.TablaSiguienteNivel
		} else {
			hoja = entrada
		}
	}
	puntero := indices[niveles]
	if puntero < 0 || puntero >= len(hoja.NumeroMarco) {
		global.MemoriaLogger.Error(fmt.Sprintf("Marco: puntero inválido en hoja: %d", puntero))
		return -1
	}
	marco := hoja.NumeroMarco[puntero]
	IncrementarAccesosTabla(pid)

	global.MemoriaLogger.Debug(fmt.Sprintf("Marco: PID=%d, indices=%v → marco=%d", pid, indices, marco))
	return marco
}
