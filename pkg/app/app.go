package app

import (
	"fmt"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/fusor/ansible-service-broker/pkg/broker"
	"github.com/fusor/ansible-service-broker/pkg/dao"
	//"github.com/fusor/ansible-service-broker/pkg/handler"
	//"net/http"
	"os"
)

type App struct {
	broker   *broker.AnsibleBroker
	args     Args
	config   Config
	dao      *dao.Dao
	log      *Log
	registry ansibleapp.Registry
}

func CreateApp() App {
	var err error

	fmt.Println("============================================================")
	fmt.Println("==           Starting Ansible Service Broker...           ==")
	fmt.Println("============================================================")

	app := App{}

	// Writing directly to stderr because log has not been bootstrapped
	if app.args, err = CreateArgs(); err != nil {
		os.Stderr.WriteString("ERROR: Failed to validate input\n")
	}

	// Writing directly to stderr because log has not been bootstrapped
	if app.args, err = CreateArgs(); err != nil {
		os.Stderr.WriteString("ERROR: Failed to validate input\n")
		os.Stderr.WriteString(err.Error())
		ArgsUsage()
		os.Exit(127)
	}

	if app.config, err = CreateConfig(app.args.ConfigFile); err != nil {
		os.Stderr.WriteString("ERROR: Failed to read config file\n")
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	if app.log, err = NewLog(app.config.Log); err != nil {
		os.Stderr.WriteString("ERROR: Failed to initialize logger\n")
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	app.log.Debug("Connecting Dao")
	if app.dao, err = dao.NewDao(app.config.Dao, app.log.Logger); err != nil {
		app.log.Error("Failed to initialize Dao\n")
		app.log.Error(err.Error())
		os.Exit(1)
	}

	app.log.Debug("Connecting Registry")
	if app.registry, err = ansibleapp.NewRegistry(
		app.config.Registry, app.log.Logger,
	); err != nil {
		app.log.Error("Failed to initialize Dao\n")
		app.log.Error(err.Error())
		os.Exit(1)
	}

	app.log.Debug("Creating AnsibleBroker")
	if app.broker, err = broker.NewAnsibleBroker(
		app.dao, app.log.Logger, app.registry,
	); err != nil {
		app.log.Error("Failed to create AnsibleBroker\n")
		app.log.Error(err.Error())
		os.Exit(1)
	}

	return app
}

func (a *App) Start() {
	a.log.Notice("Ansible Service Broker Started")

	////////////////////////////////////////////////////////////
	// TODO: Expecting to start the http server here with the initialized broker.
	////////////////////////////////////////////////////////////
	//a.log.Notice("Listening on http://localhost:1338")
	//http.ListenAndServe(":1338", handler.NewHandler(a.broker))
	////////////////////////////////////////////////////////////
}
