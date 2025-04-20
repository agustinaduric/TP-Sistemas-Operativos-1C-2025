package structs

import (
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

type TipoDeInstruccion string

const (
	NOOP             TipoDeInstruccion = "NOOP"
	WRITE            TipoDeInstruccion = "WRITE"
	READ             TipoDeInstruccion = "READ"
	GOTO             TipoDeInstruccion = "GOTO"
	IO_INST          TipoDeInstruccion = "IO"
	INIT_PROC_INST   TipoDeInstruccion = "INIT_PROC"
	DUMP_MEMORY_INST TipoDeInstruccion = "DUMP_MEMORY"
	EXIT_INST        TipoDeInstruccion = "EXIT"
)

//Tengo dudas de donde meter al offset, asi q no lo meti

type CPU struct {
	Config    config.CPUConfig //Cada cpu tiene su propio config por lo q entiendo
	Id        int              //Para reconocer a los multiples cpus
	PIDActual int              // PID del proceso en Exec
	PCActual  int              // PC actual dentro del proceso
	TLB       []EntradaTLB     // 
	Cache     []EntradaCache   // 
}

type Instrucion struct {
	Tipo   TipoDeInstruccion
	Param1 string
	Param2 string
}

type EntradaTLB struct {
	Pagina    int
	Frame     int
	Timestamp int64 // para LRU o FIFO
}

type EntradaCache struct {
	Pagina     int
	Contenido  []byte
	Modificado bool // si se escribio y hay que poner en Memoria
	Referencia bool // para CLOCK
}
