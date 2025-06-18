package structs

import (
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

const (
	NOOP         string = "NOOP"
	WRITE        string = "WRITE"
	READ         string = "READ"
	GOTO         string = "GOTO"
	IO           string = "IO"
	INIT_PROC    string = "INIT_PROC"
	DUMP_MEMORY  string = "DUMP_MEMORY"
	EXIT_PROC    string = "EXIT"
	REPLANIFICAR string = "REPLANIFICAR"
)

//Tengo dudas de donde meter al offset, asi q no lo meti

type CPU struct {
	Config     config.CPUConfig //Cada cpu tiene su propio config por lo q entiendo
	Id         int              //Para reconocer a los multiples cpus
	PIDActual  int              // PID del proceso en Exec
	PCActual   int              // PC actual dentro del proceso
	TLB        []EntradaTLB     //
	Cache      []EntradaCache   //
	Disponible bool
}

// hay uno así ya en la de structs.go.Este nunca lo llegamos a usar
/*
type Instrucion struct {
	Tipo   string
	Param1 string
	Param2 string
}
*/
type EntradaTLB struct {
	Pagina    int
	Frame     int
	Timestamp int64 // para LRU o FIFO
}

type EntradaCache struct {
	PID int
	Pagina     int
	Contenido  []byte
	BitUso bool
	BitModificado bool
}

type DevolucionCpu struct {
	PID           int
	PC            int
	Motivo        string
	SolicitudIO   Solicitud
	ArchivoInst   string
	Tamaño        int
	Identificador string
}

type CPU_a_kernel struct {
	IP            string
	Puerto        int
	Identificador string
	Disponible    bool
}

type CPU_a_memoria struct {
	IP            string
	Puerto        int
	Identificador string
}