package runtime

func (k fakeKubernetes) GetRuntime() string {
	return "kubernetes"
}
