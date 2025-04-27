package structs

type ProcesoMemoria struct {
	PID          int
	Tamanio      int
	TablaDePaginas *TablaDePaginas
	EnSwap       bool
	Metricas     MetricasMemoria
	Path	     string
	Instrucciones string
}

type TablaDePaginas struct {
    NivelDeTabla int 	
    PaginaMarco []PaginaMarco		//		
    SiguienteNivel   *TablaDePaginas  // puntero a la tabla del siguiente nivel
}

type PaginaMarco struct {
	Pagina int
	Marco int
}

var MemoriaUsuario []byte

type MetricasMemoria struct {
	AccesosTabla    int
	InstSolicitadas int
	BajadasSwap     int
	SubidasMem      int
	LecturasMem     int
	EscriturasMem   int
}

type EntradaSwap struct {
	PID    int
	Pagina int
	Offset int64 // posicion en archivo swap
}
