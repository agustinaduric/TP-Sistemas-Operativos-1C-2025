package global

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var MemoriaMutex sync.Mutex

var MemoriaLogger *logger.LoggerStruct

var MemoriaConfig config.MemoriaConfig

var MapMemoriaDeUsuario []int

var MemoriaUsuario []byte

var Procesos []structs.ProcesoMemoria

var ProcesosTP []structs.ProcesoTP

var IPkernel string

var PuertoKernel int
