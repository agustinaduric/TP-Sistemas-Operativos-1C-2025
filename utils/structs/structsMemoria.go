package structs

type ProcesoMemoria struct {
	PID           int             `json:"pid"`
	Tamanio       int             `json:"tamanio"`
	EnSwap        bool            `json:"en_swap"`
	Metricas      MetricasMemoria `json:"metricas"`
	Path          string          `json:"path"`
	Instrucciones []Instruccion   `json:"instrucciones"`
}

type MetricasMemoria struct {
	AccesosTabla    int `json:"accesos_tabla"`
	InstSolicitadas int `json:"inst_solicitadas"`
	BajadasSwap     int `json:"bajadas_swap"`
	SubidasMem      int `json:"subidas_mem"`
	LecturasMem     int `json:"lecturas_mem"`
	EscriturasMem   int `json:"escrituras_mem"`
}

type ProcesoTP struct {
	PID int
	TPS []Tp
}

type Tp struct {
	Paginas []int
}

type Devolucion_DumpMemory struct {
	PID       int
	Respuesta string
}

type EspacioLibreRespuesta struct {
	BytesLibres int
}

type Datos_memoria struct {
	Tamaño_pagina    int
	Cant_entradas    int
	Numeros_de_nivel int
}
