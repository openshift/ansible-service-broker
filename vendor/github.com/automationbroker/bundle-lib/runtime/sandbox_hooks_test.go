package runtime

import "testing"

func createHook(p, n string, targets []string, role string) error {
	return nil
}

func destroyHook(p, n string, targets []string) error {
	return nil
}

func TestAddPreCreateSandbox(t *testing.T) {
	p := provider{}
	p.addPreCreateSandbox(createHook)
	if len(p.preSandboxCreate) == 0 {
		t.Fatal("sandbox hooks was not added")
	}
}

func TestAddPostCreateSandbox(t *testing.T) {
	p := provider{}
	p.addPostCreateSandbox(createHook)
	if len(p.postSandboxCreate) == 0 {
		t.Fatal("sandbox hooks was not added")
	}

}

func TestAddPreDestroySandbox(t *testing.T) {
	p := provider{}
	p.addPreDestroySandbox(destroyHook)
	if len(p.preSandboxDestroy) == 0 {
		t.Fatal("sandbox hooks was not added")
	}
}

func TestAddPostDestroySandbox(t *testing.T) {
	p := provider{}
	p.addPostDestroySandbox(destroyHook)
	if len(p.postSandboxDestroy) == 0 {
		t.Fatal("sandbox hooks was not added")
	}
}
