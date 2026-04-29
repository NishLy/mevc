package main

import (
	"os"
	"os/exec"

	"github.com/NishLy/go-fiber-boilerplate/config"
)

func main() {
	configApp, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	name := "_schema"

	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	cmd := exec.Command(
		"atlas", "migrate", "diff", name,
		"--env", "gorm",
		"--dev-url", configApp.DB_DEVELOPMENT_URL,
		"--dir", "file://migrations",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic("atlas diff failed: " + err.Error())
	}
}
