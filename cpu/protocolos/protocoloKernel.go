package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/cpu/cache"
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
	url := fmt.Sprintf("http://%s:%d/conectarcpu", global.ConfigCargadito.IpKernel, global.ConfigCargadito.PortKernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando CPU:%s a puerto:%d", os.Args[1], global.ConfigCargadito.PortKernel)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
}

func Enviar_syscall(DevolucionSyscall structs.DevolucionCpu) {
	body, err := json.Marshal(DevolucionSyscall)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/devolucion", global.ConfigCargadito.IpKernel, global.ConfigCargadito.PortKernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando CPU:%s a puerto:%d", os.Args[1], global.ConfigCargadito.PortKernel)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	if global.Hayinterrupcion{ global.SyscallEnviada<-0}
	global.TLB = nil
	cache.LimpiarCacheDelProceso(global.Proceso_Ejecutando.PID)
}

func Ocurrio_Interrupcion(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var interrupcion string
	err := decoder.Decode(&interrupcion)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}
	if interrupcion == "interrupcion" {
		global.CpuLogger.Info("## Llega interrupción al puerto Interrupt")
		global.Hayinterrupcion = true
	}
}

func Reconectar_Proceso(w http.ResponseWriter, r *http.Request) {
	global.CpuLogger.Debug("## Llego una solicitud de reconexion")
	decoder := json.NewDecoder(r.Body)
	var Reconectar string
	err := decoder.Decode(&Reconectar)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}
	if Reconectar == "Reconectar" {
		global.CpuLogger.Debug("se envia señal para la reconexion")
		global.Proceso_reconectado <- 0

	}
}
