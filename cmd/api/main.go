package main

import "reservista.kz/internal/app"

const configsDir = "configs"
const envDir = ".env"

func main() {
	app.Run(configsDir, envDir)
}
