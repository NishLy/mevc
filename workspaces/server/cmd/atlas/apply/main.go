package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/NishLy/go-fiber-boilerplate/config"
)

func main() {
	configApp, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	cmd := exec.Command(
		"atlas", "migrate", "apply",
		"--url", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			configApp.DBUSER,
			configApp.DBPASS,
			configApp.DBHOST,
			configApp.DBPORT,
			configApp.DBNAME,
		),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic("atlas apply failed: " + err.Error())
	}
}
