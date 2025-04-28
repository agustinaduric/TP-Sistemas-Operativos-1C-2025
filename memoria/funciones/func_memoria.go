package fmemoria

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/comunicacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var procesos = make(map[int]*structs.ProcesoMemoria)

func IniciarConfiguracionMemoria(filePath string) config.MemoriaConfig {
	var config config.MemoriaConfig
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func LevantarServidorMemoria(configCargadito config.MemoriaConfig) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", comunicacion.RecibirMensaje)

	puerto := config.IntToStringConPuntos(configCargadito.PortMemory)

	log.Printf("Servidor de Memoria escuchando en %s", puerto)
	err := http.ListenAndServe(puerto, mux)
	if err != nil {
		log.Fatalf("Error al levantar el servidor: %v", err)
	}
}

func IniciarMemoriaUsuario(configCargadito config.MemoriaConfig) []byte {
	return make([]byte, configCargadito.MemorySize)
}

func CantidadMarcos(configCargadito config.MemoriaConfig) int {
	return configCargadito.MemorySize / configCargadito.PageSize
}

func IniciarBitMapMemoriaUsuario(configCargadito config.MemoriaConfig) []int { //QUiero creer que se inicializan todos en 0
	return make([]int, CantidadMarcos(configCargadito))
}

func MarcosDisponibles(configCargadito config.MemoriaConfig, bitMapMemoriaUsuario []int) int {
	var contador int = 0
	for i := 0; i < CantidadMarcos(configCargadito); i++ {
		if bitMapMemoriaUsuario[i] == 1 {
			contador++
		}
	}
	return contador
}

func BuscarInstruccion(pid int, pc int) (structs.Instruccion, error) {
	proceso, existe := procesos[pid]
	if !existe {
		return structs.Instruccion{}, fmt.Errorf("BuscarInstruccion: el pid %d no existe", pid)
	}
	if pc < 0 || pc >= len(proceso.Instrucciones) {
		return structs.Instruccion{}, fmt.Errorf("BuscarInstruccion: pc %d fuera de rango [0,%d) para pid %d",
			pc, len(proceso.Instrucciones), pid)
	}
	return proceso.Instrucciones[pc], nil
}

// trankilo compilador idiota, la vamos a usar
func hayEspacio(configCargadito config.MemoriaConfig, bitMapMemoriaUsuario []int, tamanioProceso int) bool {
	divisionTamPag := tamanioProceso / configCargadito.PageSize
	var cantidadDeMarcosRequeridos int
	if divisionTamPag == 0 {
		cantidadDeMarcosRequeridos = divisionTamPag
	} else {
		cantidadDeMarcosRequeridos = divisionTamPag + 1
	}

	if cantidadDeMarcosRequeridos <= MarcosDisponibles(configCargadito, bitMapMemoriaUsuario) {
		return true
	}
	return false
}

func InicializarTablas(proceso *structs.ProcesoMemoria, niveles, entradas int) {
	proceso.TablaDePaginas = crearTablaRecursiva(1, niveles, entradas)
}

func crearTablaRecursiva(nivelActual, nivelesTotal, entradas int) *structs.TablaDePaginas {
	tabla := &structs.TablaDePaginas{
		NivelDeTabla: nivelActual,
		PaginaMarco:  make([]structs.PaginaMarco, entradas),
	}
	for i := 0; i < entradas; i++ {
		tabla.PaginaMarco[i].Pagina = -1
		tabla.PaginaMarco[i].Marco = -1
	}
	if nivelActual < nivelesTotal {
		tabla.SiguienteNivel = crearTablaRecursiva(nivelActual+1, nivelesTotal, entradas)
	} else {
		tabla.SiguienteNivel = nil
	}
	return tabla
}

func CargarInstrucciones(ruta string) ([]structs.Instruccion, error) {
	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer archivo.Close()

	var instrucciones []structs.Instruccion
	escáner := bufio.NewScanner(archivo)
	for escáner.Scan() {
		línea := strings.TrimSpace(escáner.Text())
		if línea == "" {
			continue
		}
		partes := strings.Fields(línea)
		instrucciones = append(instrucciones, structs.Instruccion{
			Operacion:  partes[0],
			Argumentos: partes[1:],
		})
	}
	if err := escáner.Err(); err != nil {
		return nil, err
	}
	return instrucciones, nil
}
