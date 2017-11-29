package runtime

type fakeProvider struct{}

func (f fakeProvider) CreateSandbox(podName string, namespace string, targets []string, apbRole string) (string, error) {
	//TODO: Write tests for CreateSandbox using the fake kubernetes client
	return "", nil
}

func (f fakeProvider) DestroySandbox(podName string, namespace string, targets []string, configNamespace string, keepNamespace bool, keepNamespaceOnError bool) {
	//TODO: Write tests for DestroySandbox using the fake kubernetes client
	return
}
