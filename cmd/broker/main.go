package main

import (
	"github.com/fusor/ansible-service-broker/pkg/app"
)

func main() {
	app := app.CreateApp()
	app.Start()
	////////////////////////////////////////////////////////////
	// TODO:
	// try/finally to make sure we clean things up cleanly?
	//if stopsignal {
	//app.stop() // Stuff like close open files
	//}
	////////////////////////////////////////////////////////////
}
