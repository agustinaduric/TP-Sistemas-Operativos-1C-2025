package fmemoria

import (
	"os/exec"
)

// VisualizadorDeDump abre un xterm mostrando el contenido hexadecimal del archivo .dmp
// y no realiza ningún log, mensaje o interacción adicional.
func VisualizadorDeDump(ruta string) {
	exec.Command(
		"xterm",
		"-hold", // mantiene abierta la ventana hasta que se cierre manualmente
		"-e",    // comando a ejecutar
		"bash", "-c",
		"hexdump -C "+ruta+".dmp | less",
	).Start()
}

// VisualizadorDeSwap abre un xterm mostrando el contenido hexadecimal del swapfile
// y no realiza ningún log, mensaje o interacción adicional.
func VisualizadorDeSwap() {
	exec.Command(
		"xterm",
		"-hold",
		"-e",
		"bash", "-c",
		"hexdump -C /home/utnso/swapfile.bin | less",
	).Start()
}
