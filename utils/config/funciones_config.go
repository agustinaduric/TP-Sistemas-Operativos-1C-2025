package config

import (
	"fmt"
)

func IntToStringConPuntos(valor int) string {
	return fmt.Sprintf(":%d", valor)
}