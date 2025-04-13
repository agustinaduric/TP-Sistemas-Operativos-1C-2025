package memoria

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type ConfigMemoria struct {
	PuertoMemoria    int    `json:"port_memory"`
	MemorySize       int    `json:"memory_size"`
	PageSize         int    `json:"page_size"`
	EntriesPerPage   int    `json:"entries_per_page"`
	NumberOfLevels   int    `json:"number_of_levels"`
	MemoryDelay      int    `json:"memory_delay"`
	SwapPath         string `json:"swapfile_path"`
	SwapDelay        int    `json:"swap_delay"`
	NivelLog         string `json:"log_level"`
	DumpPath         string `json:"dump_path"`
}

var config ConfigMemoria


var (
	memoriaUsuario []byte       // En vez de (void*) esto es mas god
	memoriaMutex   sync.Mutex   // para proteger la memoria en concurrencia
	swapFile       *os.File     // archivo de SWAP abierto globalmente
)
