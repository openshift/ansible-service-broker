//
// Copyright (c) 2017 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubeversiontypes "k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	//    genericoptions "k8s.io/apiserver/pkg/server/options"
	//   authenticationclient "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/handler"
	"github.com/openshift/ansible-service-broker/pkg/registries"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
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
	registry []registries.Registry
	engine   *broker.WorkEngine
}

func createClientConfigFromFile(configPath string) (*restclient.Config, error) {
	clientConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ApiServer(log *logging.Logger) (*App, error) {
	serverConfig := genericapiserver.NewConfig(Codecs)
	if err := o.SecureServing.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	// vvvv STOLEN FROM clients.go vvvv
	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Warning("Failed to create a InternalClientSet: %v.", err)

		log.Debug("Checking for a local Cluster Config")
		clientConfig, err = createClientConfigFromFile(homedir.HomeDir() + "/.kube/config")
		if err != nil {
			log.Error("Failed to create LocalClientSet")
			return nil, err
		}
	}
	// ^^^^ STOLEN FROM clients.go ^^^^

	client, err := authenticationclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	authenticationConfig := authenticatorfactory.DelegatingAuthenticatorConfig{
		Anonymous:               true,
		TokenAccessReviewClient: client.TokenReviews(),
		CacheTTL:                o.Authentication.CacheTTL,
	}
	authenticator, _, err := authenticationConfig.New()
	if err != nil {
		return nil, err
	}
	serverConfig.Authenticator = authenticator

	if err := o.Authorization.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	/*
	   config := &server.TemplateServiceBrokerConfig{
	       GenericConfig: serverConfig,

	       TemplateNamespaces: o.TSBConfig.TemplateNamespaces,
	       // TODO add the code to set up the client and informers that you need here
	   }
	   return config, nil
	*/

	// TSB had the following, TemplateServiceBrokerConfig.GenericConfig -- serverConfig
	// genericServer, err := c.TemplateServiceBrokerConfig.GenericConfig.SkipComplete().New("template-service-broker", delegationTarget)
	fmt.Println("apiserver creating?")
	genericServer, err := serverConfig.SkipComplete().New("ansible-service-broker", nil)

	fmt.Println("apiserver created")

	return nil, nil
}

// CreateApp - Creates the application
func CreateApp() App {
	var err error
	app := App{}

	// Writing directly to stderr because log has not been bootstrapped
	if app.args, err = CreateArgs(); err != nil {
		os.Exit(1)
	}

	if app.args.Version {
		fmt.Println(Version)
		os.Exit(0)
	}

	fmt.Println("============================================================")
	fmt.Println("==           Starting Ansible Service Broker...           ==")
	fmt.Println("============================================================")

	// TODO: Let's take all these validations and delegate them to the client
	// pkg.
	if app.config, err = CreateConfig(app.args.ConfigFile); err != nil {
		os.Stderr.WriteString("ERROR: Failed to read config file\n")
		os.Stderr.WriteString(err.Error() + "\n")
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
		reg, err := registries.NewRegistry(r, app.log.Logger)
		if err != nil {
			app.log.Errorf(
				"Failed to initialize %v Registry err - %v \n", r.Name, err)
			os.Exit(1)
		}
		app.registry = append(app.registry, reg)
	}

	app.log.Debug("Initializing WorkEngine")
	app.engine = broker.NewWorkEngine(MsgBufferSize)
	err = app.engine.AttachSubscriber(
		broker.NewProvisionWorkSubscriber(app.dao, app.log.Logger),
		broker.ProvisionTopic)
	if err != nil {
		app.log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	err = app.engine.AttachSubscriber(
		broker.NewDeprovisionWorkSubscriber(app.dao, app.log.Logger),
		broker.DeprovisionTopic)
	if err != nil {
		app.log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	app.log.Debugf("Active work engine topics: %+v", app.engine.GetActiveTopics())

	apb.InitializeSecretsCache(app.config.Secrets, app.log.Logger)
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

	if a.config.Broker.BootstrapOnStartup {
		a.log.Info("Broker configured to bootstrap on startup")
		a.log.Info("Attempting bootstrap...")
		if _, err := a.broker.Bootstrap(); err != nil {
			a.log.Error("Failed to bootstrap on startup!")
			a.log.Error(err.Error())
			os.Exit(1)
		}
		a.log.Notice("Broker successfully bootstrapped on startup")
	}

	interval, err := time.ParseDuration(a.config.Broker.RefreshInterval)
	a.log.Debug("RefreshInterval: %v", interval.String())
	if err != nil {
		a.log.Error(err.Error())
		a.log.Error("Not using a refresh interval")
	} else {
		ticker := time.NewTicker(interval)
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()
		go func() {
			for {
				select {
				case v := <-ticker.C:
					a.log.Info("Broker configured to refresh specs every %v seconds", interval)
					a.log.Info("Attempting bootstrap at %v", v.UTC())
					if _, err := a.broker.Bootstrap(); err != nil {
						a.log.Error("Failed to bootstrap")
						a.log.Error(err.Error())
					}
					a.log.Notice("Broker successfully bootstrapped")
				case <-ctx.Done():
					ticker.Stop()
					return
				}
			}
		}()
	}

	a.log.Notice("Ansible Service Broker Started")
	listeningAddress := "0.0.0.0:1338"
	if a.args.Insecure {
		a.log.Notice("Listening on http://%s", listeningAddress)
		err = http.ListenAndServe(":1338",
			handler.NewHandler(a.broker, a.log.Logger, a.config.Broker))
	} else {
		a.log.Notice("Listening on https://%s", listeningAddress)
		err = http.ListenAndServeTLS(":1338",
			a.config.Broker.SSLCert,
			a.config.Broker.SSLCertKey,
			handler.NewHandler(a.broker, a.log.Logger, a.config.Broker))
	}
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
