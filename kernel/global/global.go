package global

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var KernelLogger *logger.LoggerStruct
var ConfigCargadito config.KernelConfig
var WgKernel sync.WaitGroup

//-----------------------------------------------SEMOFOROS DE PLANIFICADORES-------------------------------------------------------

var ProcesoCargado = make(chan int)
var ProcesoListo = make(chan int)
var ProcesoEnSuspReady = make(chan int)
var ProcesoParaFinalizar = make(chan int)

//-----------------------------------------------ADMINISTRACION DE CPU-------------------------------------------------------

var MutexCpuDisponible sync.Mutex
var MutexCpuNoDisponibles sync.Mutex

var (
	SemaforosCPU      = make(map[string]chan struct{})
	MutexSemaforosCPU sync.Mutex
)

var Contador int = 0
var MutexContador sync.Mutex

//-----------------------------------------------MUTEX COLAS-------------------------------------------------------

var MutexNEW sync.Mutex
var MutexREADY sync.Mutex
var MutexEXEC sync.Mutex
var MutexBLOCKED sync.Mutex
var MutexSUSP_BLOCKED sync.Mutex
var MutexSUSP_READY sync.Mutex
var MutexEXIT sync.Mutex

var MutexBLOCKED_IO sync.Mutex

//-----------------------------------------------FUNCIONES AUXILIARES-------------------------------------------------------

func Pop_estado(Cola *structs.ColaProcesos) structs.PCB {

	if len(*Cola) == 0 {
		return structs.PCB{}
	}

	pcb := (*Cola)[0]
	*Cola = (*Cola)[1:] // directamente recortás el slice

	return pcb
}

func Push_estado(Cola *structs.ColaProcesos, pcb structs.PCB) {
	*Cola = append(*Cola, pcb) // Usamos el puntero a Cola para modificar el slice
}

func Extraer_estado(Cola *structs.ColaProcesos, PID int) structs.PCB {
	var extraido structs.PCB
	for i := 0; i < len(*Cola); i++ {
		if (*Cola)[i].PID == PID {
			extraido = (*Cola)[i]
			*Cola = append((*Cola)[:i], (*Cola)[i+1:]...)

			return extraido
		}
	}
	return structs.PCB{}
}

func IniciarMetrica(estadoViejo string, estadoNuevo string, proceso *structs.PCB) {
	if proceso.Estado == structs.SUSP_BLOCKED && estadoViejo == "BLOCKED" {
		if estadoNuevo == "READY" {
			IniciarMetrica("SUSP_BLOCKED", "SUSP_READY", proceso)
			return
		}
		if estadoNuevo == "EXIT" {
			IniciarMetrica("SUSP_BLOCKED", "EXIT", proceso)
			return
		}
	}
	switch estadoNuevo {
	case "NEW":
		proceso.MetricasEstado[structs.NEW] = proceso.MetricasEstado[structs.NEW] + 1
		proceso.Estado = structs.NEW
		proceso.TiempoInicioEstado = time.Now()
		MutexNEW.Lock()
		Push_estado(&structs.ColaNew, *proceso)
		MutexNEW.Unlock()
		KernelLogger.Info(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", proceso.PID))
		ProcesoCargado <- 0
	case "READY":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.READY
		proceso.MetricasEstado[structs.READY] = proceso.MetricasEstado[structs.READY] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexREADY.Lock()
		Push_estado(&structs.ColaReady, *proceso)
		MutexREADY.Unlock()
		KernelLogger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado READY", proceso.PID, estadoViejo))
		KernelLogger.Debug("Se intenta enviar aviso desde plani largo a plani corto")
		ProcesoListo <- 0
		KernelLogger.Debug("Se envia aviso desde plani largo a plani corto")
	case "EXEC":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.EXEC
		proceso.MetricasEstado[structs.EXEC] = proceso.MetricasEstado[structs.EXEC] + 1
		proceso.TiempoInicioEstado = time.Now()
		proceso.Auxiliar = float64(proceso.TiemposEstado[structs.EXEC])

		if !proceso.Desalojado {proceso.TiempoRestante= proceso.EstimadoRafaga}
		

		KernelLogger.Debug(fmt.Sprintf("## (%d) Auxiliar: %f ", proceso.PID, proceso.Auxiliar))

		MutexEXEC.Lock()
		Push_estado(&structs.ColaExecute, *proceso)
		MutexEXEC.Unlock()
		KernelLogger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado EXEC", proceso.PID, estadoViejo))
	case "BLOCKED":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.BLOCKED
		proceso.MetricasEstado[structs.BLOCKED] = proceso.MetricasEstado[structs.BLOCKED] + 1
		proceso.TiempoInicioEstado = time.Now()
		go IniciarContadorDeSuspension(proceso, proceso.MetricasEstado[structs.BLOCKED])
		MutexBLOCKED.Lock()
		Push_estado(&structs.ColaBlocked, *proceso)
		MutexBLOCKED.Unlock()
		KernelLogger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado BLOCKED", proceso.PID, estadoViejo))
	case "SUSP_BLOCKED":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.SUSP_BLOCKED
		proceso.MetricasEstado[structs.SUSP_BLOCKED] = proceso.MetricasEstado[structs.SUSP_BLOCKED] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexSUSP_BLOCKED.Lock()
		Push_estado(&structs.ColaSuspBlocked, *proceso)
		MutexSUSP_BLOCKED.Unlock()
		KernelLogger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado SUSP_BLOCKED", proceso.PID, estadoViejo))
	case "SUSP_READY":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.SUSP_READY
		proceso.MetricasEstado[structs.SUSP_READY] = proceso.MetricasEstado[structs.SUSP_READY] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexSUSP_READY.Lock()
		Push_estado(&structs.ColaSuspReady, *proceso)
		MutexSUSP_READY.Unlock()
		KernelLogger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado SUSP_READY", proceso.PID, estadoViejo))
		ProcesoEnSuspReady <- 0
	case "EXIT":
		DetenerMetrica(estadoViejo, proceso)
		proceso.Estado = structs.EXIT
		proceso.MetricasEstado[structs.EXIT] = proceso.MetricasEstado[structs.EXIT] + 1
		proceso.TiempoInicioEstado = time.Now()
		MutexEXIT.Lock()
		Push_estado(&structs.ColaExit, *proceso)
		MutexEXIT.Unlock()
		ProcesoParaFinalizar <- 0
		KernelLogger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado EXIT", proceso.PID, estadoViejo))
	case "FINALIZADO":
		DetenerMetrica(estadoViejo, proceso)
		KernelLogger.Info(fmt.Sprintf("## (%d) - Finaliza el proceso", proceso.PID))
		KernelLogger.Info(fmt.Sprintf("## (%d) - Métricas de estado: NEW (%d) (%d), READY (%d) (%d), EXEC (%d) (%d), BLOCKED (%d) (%d), SUSP. BLOCKED (%d) (%d), SUSP. READY (%d) (%d), EXIT (%d) (%d)",
			proceso.PID,
			proceso.MetricasEstado[structs.NEW], proceso.TiemposEstado[structs.NEW].Milliseconds(),
			proceso.MetricasEstado[structs.READY], proceso.TiemposEstado[structs.READY].Milliseconds(),
			proceso.MetricasEstado[structs.EXEC], proceso.TiemposEstado[structs.EXEC].Milliseconds(),
			proceso.MetricasEstado[structs.BLOCKED], proceso.TiemposEstado[structs.BLOCKED].Milliseconds(),
			proceso.MetricasEstado[structs.SUSP_BLOCKED], proceso.TiemposEstado[structs.SUSP_BLOCKED].Milliseconds(),
			proceso.MetricasEstado[structs.SUSP_READY], proceso.TiemposEstado[structs.SUSP_READY].Milliseconds(),
			proceso.MetricasEstado[structs.EXIT], proceso.TiemposEstado[structs.EXIT].Milliseconds()))
	}
}

func DetenerMetrica(estadoViejo string, proceso *structs.PCB) {
	switch estadoViejo {
	case "NEW":
		Extraer_estado(&structs.ColaNew, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.NEW] += duracion
	case "READY":
		Extraer_estado(&structs.ColaReady, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.READY] += duracion
	case "EXEC":
		Extraer_estado(&structs.ColaExecute, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.EXEC] += duracion

		if !proceso.Desalojado {
			if proceso.TiempoRestante != 0{proceso.UltimaRafagaReal = float64(proceso.TiemposEstado[structs.EXEC]) - proceso.Auxiliar + proceso.TiempoGlobal
			 }else {proceso.UltimaRafagaReal = float64(proceso.TiemposEstado[structs.EXEC]) - proceso.Auxiliar }
			proceso.EstimadoRafaga = (float64(ConfigCargadito.Alpha) * proceso.UltimaRafagaReal) + ((1 - float64(ConfigCargadito.Alpha)) * proceso.EstimadoRafagaAnt)
			
			proceso.TiempoRestante= 0
			proceso.TiempoGlobal =0
			KernelLogger.Debug(fmt.Sprintf("## (%d) UltimaRafagaReal: %f ", proceso.PID, proceso.UltimaRafagaReal))
			
			KernelLogger.Debug(fmt.Sprintf("## (%d) EstimadoRafaga: %f ", proceso.PID, proceso.EstimadoRafaga))
		} else {
			aux := time.Since(proceso.TiempoInicioEstado)
			proceso.TiempoRestante = proceso.TiempoRestante - (float64(aux))
			KernelLogger.Debug(fmt.Sprintf("## (%d) TiempoRestante: %f ", proceso.PID, proceso.TiempoRestante))
			proceso.TiempoGlobal = proceso.TiempoGlobal + (float64(aux))

			//proceso.Desalojado = false
		}

	case "BLOCKED":
		Extraer_estado(&structs.ColaBlocked, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.BLOCKED] += duracion
	case "SUSP_BLOCKED":
		Extraer_estado(&structs.ColaSuspBlocked, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.SUSP_BLOCKED] += duracion
	case "SUSP_READY":
		Extraer_estado(&structs.ColaSuspReady, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.SUSP_READY] += duracion
	case "EXIT":
		Extraer_estado(&structs.ColaExit, proceso.PID)
		duracion := time.Since(proceso.TiempoInicioEstado)
		proceso.TiemposEstado[structs.EXIT] += duracion
	}
}

//-----------------------------------------------FUNCIONES QUE NO SE DONDE PONER PORQUE TODO SE CICLA----------------------------------------------------------------

func IniciarContadorDeSuspension(proceso *structs.PCB, contador int) {
	KernelLogger.Debug("Entro a la funcion IniciarContadorDeSuspension")
	time.Sleep(time.Duration(ConfigCargadito.SuspensionTime) * time.Millisecond)
	//KernelLogger.Debug(fmt.Sprintf("Estado: %s", proceso.Estado))
	//proceso.Estado != structs.BLOCKED
	if (!Esta_en_block(proceso.PID)) || (proceso.MetricasEstado[structs.BLOCKED] != contador) {
		return
	} //si el estado ya no es BLOCKED no hago nada y finalizo el hilo contador

	KernelLogger.Debug(fmt.Sprintf("Termino el conteo de suspension del proceso de PID: %d", proceso.PID))
	//si el estado sigue siendo blocked lo mando a memoria para que lo swapee y cambio su estado a SUSP_BLOCKED
	IniciarMetrica("BLOCKED", "SUSP_BLOCKED", proceso)
	if respuesta := MandarProcesoASuspension(proceso.PID); respuesta == "OK" {
		KernelLogger.Debug("Memoria avalo la suspension")
		//al suspender un proceso hay mas lugar en la memoria y capaz puede entrar otro proceso
		if len(structs.ColaSuspReady) == 0 {
			ProcesoCargado <- 0
			KernelLogger.Debug("Se envia aviso desde el contador para suspender a plani largo, la cola de SUSP_READY estaba vacia")
			return
		}
		ProcesoEnSuspReady <- 0
		KernelLogger.Debug("Se envia aviso desde el contador para suspender a plani mediano")
	} else {
		KernelLogger.Debug("Memoria no avalo la suspension, este mensaje no deberia estar supongo")
	}
}

func MandarProcesoASuspension(PID int) string {
	KernelLogger.Debug("Entramos a la funcion MandarProcesoASuspension")
	var Proceso int = PID
	body, err := json.Marshal(Proceso)
	if err != nil {
		KernelLogger.Error(fmt.Sprintf("error codificando el proceso: %s", err.Error()))
	}
	url := fmt.Sprintf("http://%s:%d/suspension-proceso", ConfigCargadito.IpMemory, ConfigCargadito.PortMemory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		KernelLogger.Error(fmt.Sprintf("error enviando proceso de PID:%d puerto:%d", PID, ConfigCargadito.PortMemory))
	}
	defer resp.Body.Close()
	var respuesta string
	json.NewDecoder(resp.Body).Decode(&respuesta)
	return respuesta
}

func Esta_en_block(PID int) bool {
	for i := 0; i < len(structs.ColaBlocked); i++ {
		if (structs.ColaBlocked)[i].PID == PID {
			//KernelLogger.Debug(fmt.Sprintf("PID: %d esta en la cola Bloqueados", PID))
			return true
		}

	}
	//KernelLogger.Debug(fmt.Sprintf("PID: %d NO esta en la cola Bloqueados", PID))
	return false

}

func Habilitar_CPU_con_plani_corto(identificador string) {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Identificador == identificador {
			structs.CPUs_Conectados[i].Disponible = true
			KernelLogger.Debug(fmt.Sprintf("Se habilita la cpu: %s y se manda señal al plani corto", identificador))
			ProcesoListo <- 0
			return
		}

	}
	return
}

func Deshabilitar_CPU_con_plani_corto(identificador string) {
	longitud := len(structs.CPUs_Conectados)
	for i := 0; i < longitud; i++ {
		if structs.CPUs_Conectados[i].Identificador == identificador {
			structs.CPUs_Conectados[i].Disponible = false
			//KernelLogger.Debug(fmt.Sprintf("Se habilita la cpu: %s y se manda señal al plani corto", identificador))
			//ProcesoListo <- 0
			return
		}

	}
	return
}


