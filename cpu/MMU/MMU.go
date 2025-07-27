package mmu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
)

func DL_a_DF(direccion_logica int) int {
	var nro_pagina int = direccion_logica / global.Page_size
	var desplazamiento int = direccion_logica % global.Page_size

	var nro_marco int = ObtenerMarco(nro_pagina)

	global.CpuLogger.Info(fmt.Sprintf("## PID: %d - OBTENER MARCO - Página: %d - Marco: %d", global.Proceso_Ejecutando.PID, nro_pagina, nro_marco))

	var direccion_fisica int = (nro_marco * global.Page_size) + desplazamiento
	return direccion_fisica
}

func ObtenerMarco(nro_pagina int) int {
	if global.ConfigCargadito.TlbEntries  != 0 {
		if ConsultarMarcoEnTLB(nro_pagina) == global.HIT {
			return global.MarcoEncontrado
		}
	}

	entradasPorNivel := make([]int, global.Number_of_levels)
	for x := 1; x <= global.Number_of_levels; x++ {
		potencia := int(math.Pow(float64(global.Entries_per_page), float64(global.Number_of_levels-x)))
		entrada := (nro_pagina / potencia) % global.Entries_per_page
		entradasPorNivel[x-1] = entrada
	}
	nro_marco := SolicitarMarco(entradasPorNivel)
	global.CpuLogger.Debug(fmt.Sprintf("marco obtenido. Pagina: %d , Marco: %d", nro_pagina, nro_marco))
	if global.ConfigCargadito.TlbEntries  != 0{
	AgregarATLB(nro_pagina, nro_marco)
	}
	return nro_marco
}

func SolicitarMarco(indice []int) int {

	global.CpuLogger.Debug("ENTRE A SOLICITAR MARCO")

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
