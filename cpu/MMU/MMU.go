package mmu

import "github.com/sisoputnfrba/tp-golang/cpu/global"

func DL_a_DF(direccion_logica_string string) string {
	var direccion_logica int = global.String_a_int(direccion_logica_string)
	var nro_pagina int = direccion_logica / global.Page_size
	var entrada_nivel_X int = (nro_pagina/global.Entries_per_page ^ (global.Number_of_levels)) % global.Entries_per_page //TENGO QUE VER COMO SE HACE BIEN
	var desplazamiento int = direccion_logica % global.Page_size

	var nro_marco int = obtener_marco(entrada_nivel_X)

	var direccion_fisica int = (nro_marco * global.Page_size) + desplazamiento
	var direccion_fisica_string string = global.Int_a_String(direccion_fisica)
	return direccion_fisica_string
}

func obtener_marco(nro_marco int) int {
	panic("unimplemented")
}
