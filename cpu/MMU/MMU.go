package mmu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
)

func DL_a_DF(direccion_logica_string string) string {
	var direccion_logica int = global.String_a_int(direccion_logica_string)
	var nro_pagina int = direccion_logica / global.Page_size
	var desplazamiento int = direccion_logica % global.Page_size

	var nro_marco int = ObtenerMarco(nro_pagina)

	global.CpuLogger.Info(fmt.Sprintf("PID: <%d> - OBTENER MARCO - Página: <%d> - Marco: <%d>", global.Proceso_Ejecutando.PID, nro_pagina, nro_marco))

	var direccion_fisica int = (nro_marco * global.Page_size) + desplazamiento
	var direccion_fisica_string string = global.Int_a_String(direccion_fisica)
	return direccion_fisica_string
}

func ObtenerMarco(nro_pagina int) int {

	if ConsultarMarcoEnTLB(nro_pagina) == global.HIT {
		return global.MarcoEncontrado
	}

	entradasPorNivel := make([]int, global.Number_of_levels)
	for x := 1; x <= global.Number_of_levels; x++ {
		potencia := int(math.Pow(float64(global.Entries_per_page), float64(global.Number_of_levels-x)))
		entrada := (nro_pagina / potencia) % global.Entries_per_page
		entradasPorNivel[x-1] = entrada
	}
	nro_marco := SolicitarMarco(entradasPorNivel)
	AgregarATLB(nro_pagina, nro_marco)
	return nro_marco
}

func SolicitarMarco(indice []int) int {
	var Solicitud global.SolicitudDeMarco = global.SolicitudDeMarco{
		PID:     global.Proceso_Ejecutando.PID,
		Indices: indice,
	}
	body, err := json.Marshal(Solicitud)
	if err != nil {
		global.CpuLogger.Error(fmt.Sprintf("error codificando la solicitud de marco: %s", err.Error()))
	}
	url := fmt.Sprintf("http://%s:%d/solicitud-marco", global.ConfigCargadito.IpMemory, global.ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.CpuLogger.Error(fmt.Sprintf("error enviando solicitud de marco del proceso con PID:%d puerto:%d", global.Proceso_Ejecutando.PID, global.ConfigCargadito.PortMemory))
	}
	defer resp.Body.Close()
	var nro_marco int
	json.NewDecoder(resp.Body).Decode(&nro_marco)
	return nro_marco
}
