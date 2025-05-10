package ciclo

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/protocolos"
)

func Ciclo() {
	//se_devolvio_contexto = 0
	global.Hubo_syscall = false

	for {
		fetch() // busca la siguiente instrucción.
		decode_and_execute()
		if global.Hubo_syscall {
			break
		}
		//checkInterrupt()     // se fija si hay interrupciones.
		//if(se_devolvio_contexto){ break}
	}
}

func fetch() {
	protocolos.Solicitar_instruccion()
	global.InstruccionRecibida <- 0 // espero que la memoria me mande la instruccion
}

func decode_and_execute() {
	global.Instruccion_ejecutando = strings.Fields(global.Instruccion) // Dividimos las partes de la instruccion en un slice global.
	if global.Instruccion_ejecutando[0] == "WRITE" || global.Instruccion_ejecutando[0] == "READ" {
		//direccionfisica := EL.CRACK.DE.MUX(global.Instruccion_ejecutando[1])
	}

	switch global.Instruccion_ejecutando[0] {
	case "WRITE":
		var direccion int = String_a_int(global.Instruccion_ejecutando[1])
		WRITE(direccion, global.Instruccion_ejecutando[2])
	case "READ":
		var direccion int = String_a_int(global.Instruccion_ejecutando[1])
		var tamanio int = String_a_int(global.Instruccion_ejecutando[2])
		READ(direccion, tamanio)
	case "NOOP":
		// no hace nada
	case "GOTO":
		var numero int = String_a_int(global.Instruccion_ejecutando[1])
		global.Proceso_Ejecutando.PC = global.Proceso_Ejecutando.PC + numero
	case "IO":
		//
		global.Hubo_syscall = true
	case "INIT_PROC":
		//
		global.Hubo_syscall = true
	case "DUMP_MEMORY":
		//
		global.Hubo_syscall = true
	case "EXIT":
		//
		global.Hubo_syscall = true
	}

}

/*void checkInterrupt(void){
    if(hay_interrupcion){

        void *ptr_tipo_interrupcion = list_get_minimum(lista_interrupciones, elemento_minimo);
        int tipo_interrupcion = *(int *)ptr_tipo_interrupcion;

        char* interrupcion_string = _enum_interrupcion_string(tipo_interrupcion);

        log_info(cpu_log_debug, "Se detecto una interrupcion de tipo < %s >", interrupcion_string);

        devolver_contexto_ejecucion(tipo_interrupcion);
        pthread_mutex_lock(&mutex_lista_interrupciones);
            list_clean_and_destroy_elements(lista_interrupciones, free);
        pthread_mutex_unlock(&mutex_lista_interrupciones);
        pthread_mutex_lock(&mutex_hay_interrupcion);
                hay_interrupcion = 0;
        pthread_mutex_unlock(&mutex_hay_interrupcion);
        se_devolvio_contexto = 1;
    }else{
        log_info(cpu_log_debug, "No hay interrupciones");
    }
}


char* _enum_interrupcion_string(int tipo_interrupcion){

    if (tipo_interrupcion == FIN_DE_QUANTUM){ return "FIN DE QUANTUM"; }
    if (tipo_interrupcion == FIN_DE_PROCESO){ return "FIN DE PROCESO"; }
    if (tipo_interrupcion == RECURSO_INVALIDO){ return "RECURSO INVALIDO"; }
    if (tipo_interrupcion == PETICION_RECURSO){ return "PETICION RECURSO"; }

    return "DESCONOCIDO";
}

void *elemento_minimo(void *elem1, void *elem2){
    return (*(int *)elem1 < *(int *)elem2) ? elem1 : elem2;
}
*/

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
