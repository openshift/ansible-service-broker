//
// Copyright (c) 2018 Red Hat, Inc.
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
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	apirbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/client-go/informers"
	authenticationclient "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	"k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/jessevdk/go-flags"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/auth"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/config"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/handler"
	"github.com/openshift/ansible-service-broker/pkg/registries"
	agnosticruntime "github.com/openshift/ansible-service-broker/pkg/runtime"
	logutil "github.com/openshift/ansible-service-broker/pkg/util/logging"
	"github.com/openshift/ansible-service-broker/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Scheme - the runtime scheme
	Scheme = runtime.NewScheme()
	// Codecs -k8s codecs for the scheme
	Codecs = serializer.NewCodecFactory(Scheme)
	// log - logging object
	log = logutil.NewLog()
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
	config   *config.Config
	dao      dao.Dao
	registry []registries.Registry
	engine   *broker.WorkEngine
}

func apiServer(config *config.Config,
	args Args,
	providers []auth.Provider) (*genericapiserver.GenericAPIServer, error) {

	log.Debug("calling NewSecureServingOptions")
	secureServing := genericoptions.NewSecureServingOptions()
	secureServing.ServerCert = genericoptions.GeneratableKeyCert{CertKey: genericoptions.CertKey{
		CertFile: config.GetString("broker.ssl_cert"),
		KeyFile:  config.GetString("broker.ssl_cert_key"),
	}}
	secureServing.BindPort = 1338
	secureServing.BindAddress = net.ParseIP("0.0.0.0")
	if err := secureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewConfig(Codecs)
	if err := secureServing.ApplyTo(serverConfig); err != nil {
		log.Debugf("error applying to %#v", err)
		return nil, err
	}

	k8s, err := clients.Kubernetes()
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		client, err := authenticationclient.NewForConfig(k8s.ClientConfig)
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
	s := informers.NewSharedInformerFactory(k8s.Client, 2*time.Hour)
	metav1.AddToGroupVersion(Scheme, metav1.Unversioned)
	return serverConfig.Complete(s).New("ansible-service-broker", genericapiserver.EmptyDelegate)
}

// CreateApp - Creates the application
func CreateApp() App {
	var err error
	app := App{}

	// Writing directly to stderr because log has not been bootstrapped
	if app.args, err = CreateArgs(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
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
	if app.config, err = config.CreateConfig(app.args.ConfigFile); err != nil {
		os.Stderr.WriteString("ERROR: Failed to read config file\n")
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	c := logutil.LogConfig{
		LogFile: app.config.GetString("log.logfile"),
		Stdout:  app.config.GetBool("log.stdout"),
		Level:   app.config.GetString("log.level"),
		Color:   app.config.GetBool("log.color"),
	}
	if err = logutil.InitializeLog(c); err != nil {
		os.Stderr.WriteString("ERROR: Failed to initialize logger\n")
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	// Initializing clients as soon as we have deps ready.
	err = initClients(app.config)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Initialize Runtime
	log.Debug("Connecting to Cluster")
	agnosticruntime.NewRuntime(nil)
	agnosticruntime.Provider.ValidateRuntime()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	log.Debug("Connecting Dao")
	app.dao, err = dao.NewDao()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	log.Debug("Connecting Registry")
	for _, config := range app.config.GetSubConfigArray("registry") {
		reg, err := registries.NewRegistry(config, app.config.GetString("openshift.namespace"))
		if err != nil {
			log.Errorf(
				"Failed to initialize %v Registry err - %v \n", config.GetString("name"), err)
			os.Exit(1)
		}
		app.registry = append(app.registry, reg)
	}

	validateRegistryNames(app.registry)

	log.Debug("Initializing WorkEngine")
	app.engine = broker.NewWorkEngine(MsgBufferSize)
	err = app.engine.AttachSubscriber(
		broker.NewProvisionWorkSubscriber(app.dao),
		broker.ProvisionTopic)
	if err != nil {
		log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	err = app.engine.AttachSubscriber(
		broker.NewDeprovisionWorkSubscriber(app.dao),
		broker.DeprovisionTopic)
	if err != nil {
		log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	err = app.engine.AttachSubscriber(
		broker.NewUpdateWorkSubscriber(app.dao),
		broker.UpdateTopic)
	if err != nil {
		log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	err = app.engine.AttachSubscriber(
		broker.NewBindingWorkSubscriber(app.dao),
		broker.BindingTopic)
	if err != nil {
		log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	err = app.engine.AttachSubscriber(
		broker.NewUnbindingWorkSubscriber(app.dao),
		broker.UnbindingTopic)
	if err != nil {
		log.Errorf("Failed to attach subscriber to WorkEngine: %s", err.Error())
		os.Exit(1)
	}
	log.Debugf("Active work engine topics: %+v", app.engine.GetActiveTopics())

	apb.InitializeSecretsCache(app.config.GetSubConfigArray("secrets"))

	log.Debug("Creating AnsibleBroker")
	// Intiialize the cluster config.
	apb.InitializeClusterConfig(app.config.GetSubConfig("openshift"))
	if app.broker, err = broker.NewAnsibleBroker(
		app.dao, app.registry, *app.engine, app.config.GetSubConfig("broker"), app.config.GetString("openshift.namespace"),
	); err != nil {
		log.Error("Failed to create AnsibleBroker\n")
		log.Error(err.Error())
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
		log.Error(err.Error())
	}

	log.Notice(msg)
}

// Start - Will start the application to listen on the specified port.
func (a *App) Start() {
	// TODO: probably return an error or some sort of message such that we can
	// see if we need to go any further.

	if a.config.GetBool("broker.recovery") {
		log.Info("Initiating Recovery Process")
		a.Recover()
	}

	if a.config.GetBool("broker.bootstrap_on_startup") {
		log.Info("Broker configured to bootstrap on startup")
		log.Info("Attempting bootstrap...")
		if _, err := a.broker.Bootstrap(); err != nil {
			log.Error("Failed to bootstrap on startup!")
			log.Error(err.Error())
			os.Exit(1)
		}
		log.Notice("Broker successfully bootstrapped on startup")
	}

	interval, err := time.ParseDuration(a.config.GetString("broker.refresh_interval"))
	log.Debug("RefreshInterval: %v", interval.String())
	if err != nil {
		log.Error(err.Error())
		log.Error("Not using a refresh interval")
	} else {
		ticker := time.NewTicker(interval)
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()
		go func() {
			for {
				select {
				case v := <-ticker.C:
					log.Info("Broker configured to refresh specs every %v seconds", interval)
					log.Info("Attempting bootstrap at %v", v.UTC())
					if _, err := a.broker.Bootstrap(); err != nil {
						log.Error("Failed to bootstrap")
						log.Error(err.Error())
					}
					log.Notice("Broker successfully bootstrapped")
				case <-ctx.Done():
					ticker.Stop()
					return
				}
			}
		}()
	}
	//Retrieve the auth providers if basic auth is configured.
	providers := auth.GetProviders(a.config)

	genericserver, servererr := apiServer(a.config, a.args, providers)
	if servererr != nil {
		log.Errorf("problem creating apiserver. %v", servererr)
		panic(servererr)
	}

	rules := []rbac.PolicyRule{}
	if !a.config.GetBool("broker.auto_escalate") {
		rules, err = retrieveClusterRoleRules(a.config.GetString("openshift.sandbox_role"))
		if err != nil {
			log.Errorf("Unable to retrieve cluster roles rules from cluster\n"+
				" You must be using OpenShift 3.7 to use the User rules check.\n%v", err)
			os.Exit(1)
		}
	}

	var clusterURL string
	if a.config.GetString("broker.cluster_url") != "" {
		if !strings.HasPrefix("/", a.config.GetString("broker.cluster_url")) {
			clusterURL = "/" + a.config.GetString("broker.cluster_url")
		} else {
			clusterURL = a.config.GetString("broker.cluster_url")
		}
	} else {
		clusterURL = defaultClusterURLPreFix
	}

	daHandler := prometheus.InstrumentHandler(
		"ansible-service-broker",
		handler.NewHandler(a.broker, a.config, clusterURL, providers, rules),
	)

	if clusterURL == "/" {
		genericserver.Handler.NonGoRestfulMux.HandlePrefix("/", daHandler)
	} else {
		genericserver.Handler.NonGoRestfulMux.HandlePrefix(fmt.Sprintf("%v/", clusterURL), daHandler)
	}

	defaultMetrics := routes.DefaultMetrics{}
	defaultMetrics.Install(genericserver.Handler.NonGoRestfulMux)

	log.Noticef("Listening on https://%s", genericserver.SecureServingInfo.Listener.Addr().String())

	log.Notice("Ansible Service Broker Starting")
	err = genericserver.PrepareRun().Run(wait.NeverStop)
	log.Errorf("unable to start ansible service broker - %v", err)

	//TODO: Add Flag so we can still use the old way of doing this.
}

func initClients(c *config.Config) error {
	// Designed to panic early if we cannot construct required clients.
	// this likely means we're in an unrecoverable configuration or environment.
	// Best we can do is alert the operator as early as possible.
	//
	// Deliberately forcing the injection of deps here instead of running as a
	// method on the app. Forces developers at authorship time to think about
	// dependencies / make sure things are ready.
	log.Notice("Initializing clients...")
	log.Debug("Trying to connect to etcd")

	// Intialize the etcd configuration
	clients.InitEtcdConfig(c)
	etcdClient, err := clients.Etcd()
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	version, err := etcdClient.GetVersion(ctx)
	if err != nil {
		return err
	}

	log.Infof("Etcd Version [Server: %s, Cluster: %s]", version.Server, version.Cluster)

	_, err = clients.Kubernetes()
	if err != nil {
		return err
	}

	return nil
}

func retrieveClusterRoleRules(clusterRole string) ([]rbac.PolicyRule, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return nil, err
	}

	// Retrieve Cluster Role that has been defined.
	k8sRole, err := k8scli.Client.RbacV1beta1().ClusterRoles().Get(clusterRole, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return convertAPIRbacToK8SRbac(k8sRole).Rules, nil
}

// convertAPIRbacToK8SRbac - because we are using the kubernetes validation,
// and they have not started using the authoritative api package for their own
// types, we need to do some conversion here now that we are on client-go 5.0.X
func convertAPIRbacToK8SRbac(apiRole *apirbac.ClusterRole) *rbac.ClusterRole {
	rules := []rbac.PolicyRule{}
	for _, pr := range apiRole.Rules {
		rules = append(rules, rbac.PolicyRule{
			Verbs:           pr.Verbs,
			APIGroups:       pr.APIGroups,
			Resources:       pr.Resources,
			ResourceNames:   pr.ResourceNames,
			NonResourceURLs: pr.NonResourceURLs,
		})
	}
	return &rbac.ClusterRole{
		TypeMeta:   apiRole.TypeMeta,
		ObjectMeta: apiRole.ObjectMeta,
		Rules:      rules,
	}
}

func validateRegistryNames(registrys []registries.Registry) {
	names := map[string]bool{}
	for _, registry := range registrys {
		if _, ok := names[registry.RegistryName()]; ok {
			panic(fmt.Sprintf("Name of registry: %v must be unique", registry.RegistryName()))
		}
		names[registry.RegistryName()] = true
	}
}
