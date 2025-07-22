package mmu

import (
	"fmt"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
)

func Inicializar_TLB() {
	global.MAX_ENTRADAS = global.ConfigCargadito.TlbEntries
	global.ALGORITMO_TLB = global.ConfigCargadito.TlbReplacement
}

func AgregarATLB(nroPagina int, nroMarco int) {
	if len(global.TLB) == global.MAX_ENTRADAS {
		EliminarEntradaConAlgoritmo()
	}
	AgregarNuevaEntrada(nroPagina, nroMarco)
}

func EliminarEntradaConAlgoritmo() {
	global.CpuLogger.Debug(
		fmt.Sprintf("Entre en ELiminarEntradaCOnAlgoritmo"),
	)
	global.CpuLogger.Debug(
		fmt.Sprintf("Tamaño de la TLB antes de eliminar: %d", len(global.TLB)),
	)
	global.TLB = global.TLB[1:]
	//en FIFO agarro el primero
	//en LRU tambien agarro el primero porque a medida que se referencian las entradas se van poniendo al final de la cola, entonces el primer elemento de la cola siempre es el que se referencio hace mas tiempo
}

func AgregarNuevaEntrada(nroPagina int, nroMarco int) {

	var entradaNueva global.EntradaDeTLB
	entradaNueva.PID = global.Proceso_Ejecutando.PID
	entradaNueva.NroPagina = nroPagina
	entradaNueva.NroMarco = nroMarco

	global.TLB = append(global.TLB, entradaNueva)
}

func BuscarEntradaTLB(nroPagina int) global.ResultadoBusqueda {
	var cantidadEntradas int = len(global.TLB)
	for i := 0; i < cantidadEntradas; i++ {
		var entrada global.EntradaDeTLB = global.TLB[i]
		if entrada.PID == global.Proceso_Ejecutando.PID && entrada.NroPagina == nroPagina {
			global.MarcoEncontrado = entrada.NroMarco
			if AlgoritmoEsLRU() {
				global.TLB = append(global.TLB[:i], global.TLB[i+1:]...) // la saco de la lista
				global.TLB = append(global.TLB, entrada)                 // la agrego al final
			}
			return global.SEARCH_OK
		}
	}
	return global.SEARCH_ERROR
}

func ConsultarMarcoEnTLB(nroPagina int) global.RespuestaTLB {
	var respuesta global.ResultadoBusqueda = BuscarEntradaTLB(nroPagina)
	if respuesta == global.SEARCH_OK {
		global.CpuLogger.Info(fmt.Sprintf("PID: <%d> - TLB HIT - Pagina: <%d>", global.Proceso_Ejecutando.PID, nroPagina))
		return global.HIT
	}
	global.CpuLogger.Info(fmt.Sprintf("PID: <%d> - TLB MISS - Pagina: <%d>", global.Proceso_Ejecutando.PID, nroPagina))
	return global.MISS
}

func AlgoritmoEsLRU() bool {
	return strings.EqualFold(global.ALGORITMO_TLB, "LRU")
}
