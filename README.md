# TP Sistemas Operativos 1C 2025


## Documentación

- **Enunciado**: [Enlace](https://docs.google.com/document/d/1zoFRoBn9QAfYSr0tITsL3PD6DtPzO2sq9AtvE8NGrkc/edit?tab=t.0)
- **Pruebas**: [Enlace](https://docs.google.com/document/d/13XPliZvUBtYjaRfuVUGHWbYX8LBs8s3TDdaDa9MFr_I/edit?tab=t.0)

## Requisitos

- Go ≥ 1.22
- Git
- Entorno Linux

## Instalación

```bash
git clone https://github.com/agustinaduric/TP-Sistemas-Operativos-1C-2025.git
cd TP-Sistemas-Operativos-1C-2025
```
## Ejecución de pruebas

### Planificación Corto Plazo

```bash
# Memoria
go run memoria.go ./config/prueba1.config.json
# Kernel
go run kernel.go PLANI_CORTO_PLAZO 0 ./config/prueba1.config.json
# CPU (instancia 1)
go run cpu.go "1" ./config/prueba1.config.json
# CPU (instancia 2)
go run cpu.go "2" ./config/prueba1b.config.json
# I/O
go run io.go "DISCO" ./config/io.config.json
```

### Planificación Mediano/Largo Plazo

```bash
# Memoria
go run memoria.go ./config/prueba2.config.json
# Kernel
go run kernel.go PLANI_LYM_PLAZO 0 ./config/prueba2.config.json
# CPU
go run cpu.go "1" ./config/prueba12.config.json
# I/O
go run io.go "DISCO" ./config/io.config.json
```

### Memoria SWAP

```bash
# Memoria
go run memoria.go ./config/prueba3.config.json
# Kernel
go run kernel.go MEMORIA_IO 90 ./config/prueba3.config.json
# CPU
go run cpu.go "1" ./config/prueba3.config.json
# I/O
go run io.go "DISCO" ./config/io.config.json
```

### Memoria Caché

```bash
# Memoria
go run memoria.go ./config/prueba45.config.json
# Kernel
go run kernel.go MEMORIA_BASE 256 ./config/prueba45.config.json
# CPU
go run cpu.go "1" ./config/prueba4.config.json
# I/O
go run io.go "DISCO" ./config/io.config.json
```

### TLB

```bash
# Memoria
go run memoria.go ./config/prueba45.config.json
# Kernel
go run kernel.go MEMORIA_BASE_TLB 256 ./config/prueba45.config.json
# CPU
go run cpu.go "1" ./config/prueba5.config.json
# I/O
go run io.go "DISCO" ./config/io.config.json
```

### EStabilidad General

```bash
# Memoria
go run memoria.go ./config/prueba6.config.json
# Kernel
go run kernel.go ESTABILIDAD_GENERAL 0 ./config/prueba6.config.json
# CPU (4 instancias)
go run cpu.go "1" ./config/prueba6a.config.json
go run cpu.go "2" ./config/prueba6b.config.json
go run cpu.go "3" ./config/prueba6c.config.json
go run cpu.go "4" ./config/prueba6d.config.json
# I/O
go run io.go "DISCO" ./config/io.config.json
go run io.go "DISCO" ./config/prueba6b.config.json
go run io.go "DISCO" ./config/prueba6c.config.json
go run io.go "DISCO" ./config/prueba6d.config.json
```

## Extras

### Visualización de la SWAP

```bash
hexdump -C /home/utnso/swapfile.bin | less
```

### Ver dumps por PID y timestamp

```bash
hexdump -C /home/utnso/dump_files/<PID>-<TIMESTAMP>.dmp | less
```