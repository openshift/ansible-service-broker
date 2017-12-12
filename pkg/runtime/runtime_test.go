package runtime

import (
	"fmt"
	"testing"
)

type fakeProvider struct {
	fakeCoe
	cluster     string
	description string
}

type fakeCoe interface {
	getRuntime() string
}

type fakeOpenshift struct{}
type fakeKubernetes struct{}

func TestRuntime(t *testing.T) {
	testCases := []fakeProvider{
		{
			fakeCoe:     openshift{},
			cluster:     "openshift",
			description: "Testing GetRuntime() for openshift",
		},
		{
			fakeCoe:     kubernetes{},
			cluster:     "kubernetes",
			description: "Testing GetRuntime() for kubernetes",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.description), func(t *testing.T) {
			runtime := tc.fakeCoe.getRuntime()
			if tc.cluster != runtime {
				t.Fatalf("Expected %s and got %s", tc.cluster, runtime)
			}
		})
	}
}

func (f fakeProvider) ValidateRuntime() error {
	//TODO: Write tests for ValidateRuntime using the fake kubernetes client
	return nil
}

func (f fakeProvider) CreateSandbox(podName string, namespace string, targets []string, apbRole string) (string, error) {
	//TODO: Write tests for CreateSandbox using the fake kubernetes client
	return "", nil
}

func (f fakeProvider) DestroySandbox(podName string, namespace string, targets []string, configNamespace string, keepNamespace bool, keepNamespaceOnError bool) {
	//TODO: Write tests for DestroySandbox using the fake kubernetes client
	return
}

func (f fakeProvider) GetRuntime() string {
	return f.GetRuntime()
}
