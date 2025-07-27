package global

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var ConfigCargadito config.CPUConfig
var Nombre string
var WgCPU sync.WaitGroup

var Instruccion string              //lo que recibimos de memoria
var Instruccion_ejecutando []string // procesado y dividido en operacion y parametros

var Proceso_Ejecutando structs.PIDyPC_Enviar_CPU

// var InstruccionRecibida = make(chan int)
var Proceso_reconectado = make(chan int)
var Ciclofinalizado = make(chan int)
var SyscallEnviada = make(chan int)

var Hubo_syscall bool
var Hubo_interrupcion bool

var CpuLogger *logger.LoggerStruct

var Datos_Memoria structs.Datos_memoria
var Hayinterrupcion bool = false

var Page_size int        // se recibe en el handshake con memoria
var Number_of_levels int // idem el de arriba
var Entries_per_page int // idem

//-----------------------------------------------------------TLB-----------------------------------------------------------------------------------------------

type ResultadoBusqueda int

const (
	SEARCH_ERROR ResultadoBusqueda = iota
	SEARCH_OK
)

type RespuestaTLB int

const (
	HIT RespuestaTLB = iota
	MISS
)

type EntradaDeTLB struct {
	PID       int
	NroPagina int
	NroMarco  int
}

var TLB []EntradaDeTLB

type SolicitudDeMarco struct {
	PID     int
	Indices []int // un indice por cada nivel
}

var MarcoEncontrado int

//-----------------------------------------------------------Cache-----------------------------------------------------------------------------------------------

var CachePaginas []structs.EntradaCache

var PunteroClock int

var AlgoritmoCache string


//-----------------------------------------------------------FUNCIONES AUXULIARES-----------------------------------------------------------------------------------------------

func String_a_int(cadena string) int {
	var numero int
	numero, err := strconv.Atoi(cadena)
	if err != nil {
		CpuLogger.Error(fmt.Sprintf("Error de conversión de string a int"))
		os.Exit(1)
	}
	return numero
}

func Int_a_String(numero int) string {
	var cadena string
	cadena = strconv.Itoa(numero)
	return cadena
}
