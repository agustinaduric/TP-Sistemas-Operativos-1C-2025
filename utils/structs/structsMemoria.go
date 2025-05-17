package structs

type ProcesoMemoria struct {
	PID           int
	Tamanio       int
	EnSwap        bool
	Metricas      MetricasMemoria
	Path          string
	Instrucciones []Instruccion
}

type MetricasMemoria struct {
	AccesosTabla    int
	InstSolicitadas int
	BajadasSwap     int
	SubidasMem      int
	LecturasMem     int
	EscriturasMem   int
}


type Devolucion_DumpMemory struct {
	PID         int
	Respuesta   string
}

type EspacioLibreRespuesta struct{
	BytesLibres int 
}