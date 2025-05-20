package global

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var ConfigCargadito config.CPUConfig

var Instruccion string              //lo que recibimos de memoria
var Instruccion_ejecutando []string // procesado y dividido en operacion y parametros

var Proceso_Ejecutando structs.PIDyPC_Enviar_CPU

var InstruccionRecibida = make(chan int)

var Hubo_syscall bool
var Hayinterrupcion bool = false

var Page_size int        // se recibe en el handshake con memoria
var Number_of_levels int // idem el de arriba
var Entries_per_page int // idem

func String_a_int(cadena string) int {
	var numero int
	numero, err := strconv.Atoi(cadena)
	if err != nil {
		fmt.Println("Error de conversión:", err)
		os.Exit(1)
	}
	return numero
}

func Int_a_String(numero int) string {
	var cadena string
	cadena = strconv.Itoa(numero)
	return cadena
}
