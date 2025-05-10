package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Conectarse_con_Kernel(identificador string) {
	var Proceso structs.CPU_a_kernel = structs.CPU_a_kernel{
		IP:            global.ConfigCargadito.IpCPu,
		Puerto:        global.ConfigCargadito.PortCpu,
		Identificador: identificador,
		Disponible:    true,
	}
	body, err := json.Marshal(Proceso)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/ConectarCPU", global.ConfigCargadito.IpKernel, global.ConfigCargadito.PortKernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando CPU:%s a puerto:%d", os.Args[1], global.ConfigCargadito.PortKernel)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
}

func Enviar_syscall() {
	panic("unimplemented")
}
