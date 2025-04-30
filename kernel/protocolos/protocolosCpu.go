package protocolos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Enviar_datos_a_cpu(pcb_a_cargar structs.PCB) int {
	var PIDyPC structs.PIDyPC_Enviar_CPU = structs.PIDyPC_Enviar_CPU{
		PID: pcb_a_cargar.PID,
		PC:  pcb_a_cargar.PC,
	}
	global.MutexCpuDisponible.Lock()
	var Cpu_disponible structs.CPU = Buscar_CPU_libre()
	global.MutexCpuDisponible.Unlock()
	if Cpu_disponible.Id == 0 {
		return 0
	}
	body, err := json.Marshal(PIDyPC)
	if err != nil {
		log.Printf("error codificando el proceso: %s", err.Error())
	}
	url := fmt.Sprintf("http://%s:%d/datoCPU", Cpu_disponible.Config.IpCPu, Cpu_disponible.Config.PortCpu)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando proceso de PID:%d puerto:%d", pcb_a_cargar.PID, Cpu_disponible.Config.PortCpu)
	}
	log.Printf("respuesta del servidor: %s", resp.Status)
	return resp.StatusCode
}

func Buscar_CPU_libre() structs.CPU {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Disponible {
			return structs.CPUs_Conectados[i]
		}

	}
	log.Printf("No hay CPU's libres >:(")
	return structs.CPU{}
}