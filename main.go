package main

import (
	"os"

	gotir "github.com/grantseltzer/dwarf-to-gotir/pkg"
)

func main() {
	gotir.ParseFromPath(os.Args[1])
}
