package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	kubeversiontypes "k8s.io/apimachinery/pkg/version"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb/registry"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/handler"
)

// MsgBufferSize - The buffer for the message channel.
const MsgBufferSize = 20

// App - All the application pieces that are installed.
type App struct {
	broker   *broker.AnsibleBroker
	args     Args
	config   Config
	dao      *dao.Dao
	log      *Log
	registry []registry.Registry
	engine   *broker.WorkEngine
}

//CreateApp - Creates the application
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

	//TODO: Let's take all these validations and delegate them to the client
	// pkg.
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

	// Initializing clients as soon as we have deps ready.
	err = initClients(app.log.Logger, app.config.Dao.GetEtcdConfig())
	if err != nil {
		app.log.Error(err.Error())
		os.Exit(1)
	}

	app.log.Debug("Connecting Dao")
	app.dao, err = dao.NewDao(app.config.Dao, app.log.Logger)

	k8scli, err := clients.Kubernetes(app.log.Logger)
	if err != nil {
		app.log.Error(err.Error())
		os.Exit(1)
	}

	restcli := k8scli.CoreV1().RESTClient()
	body, err := restcli.Get().AbsPath("/version").Do().Raw()
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
	for _, r := range app.config.Registry {
		reg, err := registry.NewRegistry(r, app.log.Logger)
		if err != nil {
			app.log.Errorf(
				"Failed to initialize %v Registry err - %v \n", r.Name, err)
			os.Exit(1)
		}
		app.registry = append(app.registry, reg)
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

// Recover - Recover the application
// TODO: Make this a go routine once we have a strong and well tested
// recovery sequence.
func (a *App) Recover() {
	msg, err := a.broker.Recover()

	if err != nil {
		a.log.Error(err.Error())
	}

	a.log.Notice(msg)
}

// Start - Will start the application to listen on the specified port.
func (a *App) Start() {
	// TODO: probably return an error or some sort of message such that we can
	// see if we need to go any further.

	if a.config.Broker.Recovery {
		a.log.Info("Initiating Recovery Process")
		a.Recover()
	}

	a.log.Notice("Ansible Service Broker Started")
	listeningAddress := "0.0.0.0:1338"
	a.log.Notice("Listening on http://%s", listeningAddress)
	err := http.ListenAndServe(":1338", handler.NewHandler(a.broker, a.log.Logger, a.config.Broker))
	if err != nil {
		a.log.Error("Failed to start HTTP server")
		a.log.Error(err.Error())
		os.Exit(1)
	}
}

func initClients(log *logging.Logger, ec clients.EtcdConfig) error {
	// Designed to panic early if we cannot construct required clients.
	// this likely means we're in an unrecoverable configuration or environment.
	// Best we can do is alert the operator as early as possible.
	//
	// Deliberately forcing the injection of deps here instead of running as a
	// method on the app. Forces developers at authorship time to think about
	// dependencies / make sure things are ready.
	log.Notice("Initializing clients...")
	log.Debug("Trying to connect to etcd")

	etcdClient, err := clients.Etcd(ec, log)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	version, err := etcdClient.GetVersion(ctx)
	if err != nil {
		return err
	}

	log.Info("Etcd Version [Server: %s, Cluster: %s]", version.Server, version.Cluster)

	log.Debug("Connecting to Cluster")
	_, err = clients.Kubernetes(log)
	if err != nil {
		return err
	}

	return nil
}
