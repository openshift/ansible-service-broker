package runtime

func (o fakeOpenshift) GetRuntime() string {
	return "openshift"
}
