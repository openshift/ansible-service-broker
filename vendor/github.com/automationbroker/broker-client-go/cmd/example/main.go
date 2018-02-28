package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	examplecomclientset "github.com/automationbroker/broker-client-go/client/clientset/versioned"
)

var (
	kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	exampleClient, err := examplecomclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %v", err)
	}

	list, err := exampleClient.Automationbroker().Bundles("ansible-service-broker").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all databases: %v", err)
	}

	fmt.Printf("Got Bundles: \n")
	for _, bundle := range list.Items {
		fmt.Printf("bundle: %#v\n", bundle)
	}
	fmt.Printf("\n")

	jobList, err := exampleClient.AutomationbrokerV1().JobStates("ansible-service-broker").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all jobstates: %v", err)
	}
	fmt.Printf("Got JobLists: \n")
	for _, jobstate := range jobList.Items {
		fmt.Printf("job state: %#v\n", jobstate)
	}
	fmt.Printf("\n")

	bindList, err := exampleClient.AutomationbrokerV1().ServiceBindings("ansible-service-broker").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all jobstates: %v", err)
	}
	fmt.Printf("Got service bindings: \n")
	for _, servicebinding := range bindList.Items {
		fmt.Printf("job state: %#v\n", servicebinding)
	}
	fmt.Printf("\n")

	instList, err := exampleClient.AutomationbrokerV1().ServiceInstances("ansible-service-broker").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all jobstates: %v", err)
	}
	fmt.Printf("Got service instances: \n")
	for _, serviceinstance := range instList.Items {
		fmt.Printf("job state: %#v\n", serviceinstance)
	}
	fmt.Printf("\n")
}
