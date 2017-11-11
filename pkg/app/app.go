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

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeversiontypes "k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/server/routes"
	authenticationclient "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	"k8s.io/kubernetes/pkg/apis/rbac"
	v1beta1rbac "k8s.io/kubernetes/pkg/apis/rbac/v1beta1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/auth"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/handler"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/openshift/ansible-service-broker/pkg/registries"
	"github.com/openshift/ansible-service-broker/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Scheme - the runtime scheme
	Scheme = runtime.NewScheme()
	// Codecs -k8s codecs for the scheme
	Codecs = serializer.NewCodecFactory(Scheme)
)

const (
	// MsgBufferSize - The buffer for the message channel.
	MsgBufferSize = 20
	// defaultClusterURLPreFix - prefix for the ansible service broker.
	defaultClusterURLPreFix = "/ansible-service-broker"
)

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

func apiServer(log *logging.Logger, config Config, args Args, providers []auth.Provider) (*genericapiserver.GenericAPIServer, error) {
	log.Debug("calling NewSecureServingOptions")
	secureServing := genericoptions.NewSecureServingOptions()
	secureServing.ServerCert = genericoptions.GeneratableKeyCert{CertKey: genericoptions.CertKey{
		CertFile: config.Broker.SSLCert,
		KeyFile:  config.Broker.SSLCertKey,
	}}
	secureServing.BindPort = 1338
	secureServing.BindAddress = net.ParseIP("0.0.0.0")
	if err := secureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewConfig(Codecs)
	if err := secureServing.ApplyTo(serverConfig); err != nil {
		log.Debug("error applying to %#v", err)
		return nil, err
	}

	if len(providers) == 0 {
		clientConfig, err := clients.KubernetesConfig(log)
		if err != nil {
			return nil, err
		}
		client, err := authenticationclient.NewForConfig(clientConfig)
		if err != nil {
			return nil, err
		}

		authn := genericoptions.NewDelegatingAuthenticationOptions()
		authenticationConfig := authenticatorfactory.DelegatingAuthenticatorConfig{
			Anonymous:               true,
			TokenAccessReviewClient: client.TokenReviews(),
			CacheTTL:                authn.CacheTTL,
		}
		authenticator, _, err := authenticationConfig.New()
		if err != nil {
			return nil, err
		}
		serverConfig.Authenticator = authenticator

		authz := genericoptions.NewDelegatingAuthorizationOptions()
		if err := authz.ApplyTo(serverConfig); err != nil {
			return nil, err
		}
	}

	log.Debug("Creating k8s apiserver")
	return serverConfig.SkipComplete().New("ansible-service-broker", genericapiserver.EmptyDelegate)
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
		fmt.Println(version.Version)
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
	err = app.engine.AttachSubscriber(
		broker.NewUpdateWorkSubscriber(app.dao, app.log.Logger),
		broker.UpdateTopic)
	if err != nil {
		app.log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	app.log.Debugf("Active work engine topics: %+v", app.engine.GetActiveTopics())

	apb.InitializeSecretsCache(app.config.Secrets, app.log.Logger)
	// Initialize Metrics.
	metrics.Init(app.log.Logger)
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
	//Retrieve the auth providers if basic auth is configured.
	providers := auth.GetProviders(a.config.Broker.Auth, a.log.Logger)

	genericserver, servererr := apiServer(a.log.Logger, a.config, a.args, providers)
	if servererr != nil {
		a.log.Errorf("problem creating apiserver. %v", servererr)
		panic(servererr)
	}

	rules := []rbac.PolicyRule{}
	if !a.config.Broker.AutoEscalate {
		rules, err = retrieveClusterRoleRules(a.config.Openshift.SandboxRole, a.log.Logger)
		if err != nil {
			a.log.Errorf("Unable to retrieve cluster roles rules from cluster\n"+
				" You must be using OpenShift 3.7 to use the User rules check.\n%v", err)
			os.Exit(1)
		}
	}

	var clusterURL string
	if a.config.Broker.ClusterURL != "" {
		if !strings.HasPrefix("/", a.config.Broker.ClusterURL) {
			clusterURL = "/" + a.config.Broker.ClusterURL
		} else {
			clusterURL = a.config.Broker.ClusterURL
		}
	} else {
		clusterURL = defaultClusterURLPreFix
	}

	daHandler := prometheus.InstrumentHandler(
		"ansible-service-broker",
		handler.NewHandler(a.broker, a.log.Logger, a.config.Broker, clusterURL, providers, rules),
	)

	if clusterURL == "/" {
		genericserver.Handler.NonGoRestfulMux.HandlePrefix("/", daHandler)
	} else {
		genericserver.Handler.NonGoRestfulMux.HandlePrefix(fmt.Sprintf("%v/", clusterURL), daHandler)
	}

	defaultMetrics := routes.DefaultMetrics{}
	defaultMetrics.Install(genericserver.Handler.NonGoRestfulMux)

	a.log.Notice("Listening on https://%s", genericserver.SecureServingInfo.BindAddress)

	a.log.Notice("Ansible Service Broker Starting")
	err = genericserver.PrepareRun().Run(wait.NeverStop)
	a.log.Errorf("unable to start ansible service broker - %v", err)

	//TODO: Add Flag so we can still use the old way of doing this.
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

func retrieveClusterRoleRules(clusterRole string, log *logging.Logger) ([]rbac.PolicyRule, error) {
	k8scli, err := clients.Kubernetes(log)
	if err != nil {
		return nil, err
	}

	// Retrieve Cluster Role that has been defined.
	k8sRole, err := k8scli.Rbac().ClusterRoles().Get(clusterRole, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	rbacClusterRole := rbac.ClusterRole{}
	if v1beta1rbac.Convert_v1beta1_ClusterRole_To_rbac_ClusterRole(k8sRole, &rbacClusterRole, nil); err != nil {
		return nil, err
	}
	return rbacClusterRole.Rules, nil
}
