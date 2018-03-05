package runtime

func (k kubernetes) getRuntime() string {
	return "kubernetes"
}

func (k kubernetes) shouldJoinNetworks() (bool, PostSandboxCreate, PostSandboxDestroy) {
	return false, nil, nil
}
