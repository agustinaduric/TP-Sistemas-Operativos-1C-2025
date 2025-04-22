package structs

type ProcesoMemoria struct {
	PID          int
	Tamanio      int
	//TablaPaginas *TablaDePaginas
	EnSwap       bool
	Metricas     MetricasMemoria
}
/*
type EntradaTablaDePaginas struct {
	Presente       bool            //Cambiar nombre por alguno mejor, no me gusta
	Frame          int             // numero de marco si es ultimo nivel
	TablaSiguiente *TablaDePaginas // punterovich a siguiente nivel si no es ultimo
}

type TablaDePaginas struct {
	Entradas []*EntradaTablaDePaginas
}

Se nota que no estudie la teoria, revisar mas tarde

*/
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
