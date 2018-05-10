package runtime

import "testing"

func TestKubernetesShouldJoinNetworks(t *testing.T) {
	k := kubernetes{}
	s := k.getRuntime()
	if s != "kubernetes" {
		t.Fatal("runtime does not match kubernetes")
	}
}

func TestKubernetesGetRuntime(t *testing.T) {
	k := kubernetes{}

	jn, postCreateHook, postDestroyHook := k.shouldJoinNetworks()
	if jn || postCreateHook != nil || postDestroyHook != nil {
		t.Fatal("should join networks, or sand box hooks were not nil.")
	}
}
