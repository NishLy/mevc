// loader/main.go
package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
)

type User struct {
	ID   uint
	Name string
}

func main() {
	models := domain.GetDomains()
	stmts, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
