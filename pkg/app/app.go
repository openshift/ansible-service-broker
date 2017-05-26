package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	kubeversiontypes "k8s.io/apimachinery/pkg/version"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/handler"
)

const MsgBufferSize = 20

var Version = "v0.1.0"

type App struct {
	broker   *broker.AnsibleBroker
	args     Args
	config   Config
	dao      *dao.Dao
	log      *Log
	registry apb.Registry
	engine   *broker.WorkEngine
}

func CreateApp() App {
	var err error
	app := App{}

	// Writing directly to stderr because log has not been bootstrapped
	if app.args, err = CreateArgs(); err != nil {
		os.Stderr.WriteString("ERROR: Failed to validate input\n")
		os.Stderr.WriteString(err.Error() + "\n")
		ArgsUsage()
		os.Exit(127)
	}

	if app.args.Version {
		fmt.Println(Version)
		os.Exit(0)
	}

	fmt.Println("============================================================")
	fmt.Println("==           Starting Ansible Service Broker...           ==")
	fmt.Println("============================================================")

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
	serv, clust, err := app.dao.GetEtcdVersion(app.config.Dao)
	if err != nil {
		app.log.Error("Failed to connect to Etcd\n")
		app.log.Error(err.Error())
		os.Exit(1)
	}
	app.log.Info("Etcd Version [Server: %s, Cluster: %s]", serv, clust)

	app.log.Debug("Creating Cluster Client")
	client, err := apb.NewClient(app.log.Logger)
	if err != nil {
		app.log.Error(err.Error())
		os.Exit(1)
	}
	app.log.Info("Cluster Client Created")

	app.log.Debug("Connecting to Cluster")
	body, err := client.RESTClient.Get().AbsPath("/version").Do().Raw()
	if err != nil {
		app.log.Error(err.Error())
		os.Exit(1)
	}
	switch {
	case err == nil:
		var kubeServerInfo kubeversiontypes.Info
		err = json.Unmarshal(body, &kubeServerInfo)
		if err != nil && len(body) > 0 {
			app.log.Error(err.Error())
			os.Exit(1)
		}
		app.log.Info("Kubernetes version: %v", kubeServerInfo)
	case kapierrors.IsNotFound(err) || kapierrors.IsUnauthorized(err) || kapierrors.IsForbidden(err):
	default:
		app.log.Error(err.Error())
		os.Exit(1)
	}

	app.log.Debug("Connecting Registry")
	if app.registry, err = apb.NewRegistry(
		app.config.Registry, app.log.Logger,
	); err != nil {
		app.log.Error("Failed to initialize Registry\n")
		app.log.Error(err.Error())
		os.Exit(1)
	}

	app.log.Debug("Initializing WorkEngine")
	app.engine = broker.NewWorkEngine(MsgBufferSize)
	app.log.Debug("Initializing Provision WorkSubscriber")
	app.engine.AttachSubscriber(broker.NewProvisionWorkSubscriber(app.dao, app.log.Logger))

	app.log.Debug("Creating AnsibleBroker")
	if app.broker, err = broker.NewAnsibleBroker(
		app.dao, app.log.Logger, app.config.Openshift, app.registry, *app.engine, app.config.Broker,
	); err != nil {
		app.log.Error("Failed to create AnsibleBroker\n")
		app.log.Error(err.Error())
		os.Exit(1)
	}

	return app
}

// TODO: should this be a go routine?
// if so, what happens if the http listener is started and getting requests
// before we have recovered our state?
func (a *App) Recover() {
	msg, err := a.broker.Recover()

	if err != nil {
		a.log.Error(err.Error())
	}

	a.log.Notice(msg)
}

func (a *App) Start() {
	a.log.Info("Initiating Recovery Process")
	// TODO: probably return an error or some sort of message such that we can
	// see if we need to go any further. Do I even need a method on the App?
	// Should this just be called from main before we call Start? I feel like
	// main shouldn't give a crap if we need to recover or not.
	a.Recover()

	a.log.Notice("Ansible Service Broker Started")
	listeningAddress := "0.0.0.0:1338"
	a.log.Notice("Listening on http://%s", listeningAddress)
	err := http.ListenAndServe(":1338", handler.NewHandler(a.broker, a.log.Logger, a.config.Broker.DevBroker))
	if err != nil {
		a.log.Error("Failed to start HTTP server")
		a.log.Error(err.Error())
		os.Exit(1)
	}
}
