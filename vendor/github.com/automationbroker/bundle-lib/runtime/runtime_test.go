package runtime

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime/mocks"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	rest "k8s.io/client-go/rest"
	fakerest "k8s.io/client-go/rest/fake"
)

type fakeClientSet struct {
	*fake.Clientset
	rest.Interface
}

type fakeCoreV1 struct {
	fakecorev1.FakeCoreV1
	rest.Interface
}

func (f fakeClientSet) CoreV1() corev1.CoreV1Interface {
	return &fakeCoreV1{
		FakeCoreV1: fakecorev1.FakeCoreV1{Fake: &f.Clientset.Fake},
		Interface:  f.Interface,
	}
}

func (f fakeCoreV1) RESTClient() rest.Interface {
	return f.Interface
}

func sandboxCreateHook(pod, ns string, targetNS []string, role string) error {
	return nil
}

func sandboxDestroyHook(pod, ns string, targetNS []string) error {
	return nil
}

func TestNewRuntime(t *testing.T) {
	testCases := []struct {
		name             string
		config           Configuration
		client           *fake.Clientset
		response         *http.Response
		expectedProvider *provider
		shouldPanic      bool
		dontUseDeepEqual bool
	}{
		{
			name:   "New Default Openshift Runtime",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name:   "New Default Kubernetes Runtime not found",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			expectedProvider: &provider{
				coe:                 newKubernetes(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name:   "New Default Kubernetes Runtime unauth",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			expectedProvider: &provider{
				coe:                 newKubernetes(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name:   "New Default Kubernetes Runtime forbidden",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			expectedProvider: &provider{
				coe:                 newKubernetes(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name:   "Panic on finding cluster error",
			config: Configuration{},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			shouldPanic: true,
			expectedProvider: &provider{
				coe:                 newKubernetes(),
				ExtractedCredential: defaultExtractedCredential{},
			},
		},
		{
			name: "New Default Openshift Runtime with mock extracted credentials",
			config: Configuration{
				ExtractedCredential: &mocks.ExtractedCredential{},
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: &mocks.ExtractedCredential{},
			},
		},
		{
			name: "New Default Openshift Runtime with pre sandbox hooks",
			config: Configuration{
				PreCreateSandboxHooks:  []PreSandboxCreate{sandboxCreateHook},
				PreDestroySandboxHooks: []PreSandboxDestroy{sandboxDestroyHook},
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: defaultExtractedCredential{},
				preSandboxCreate:    []PreSandboxCreate{sandboxCreateHook},
				preSandboxDestroy:   []PreSandboxDestroy{sandboxDestroyHook},
			},
			dontUseDeepEqual: true,
		},
		{
			name: "New Default Openshift Runtime with pre sandbox hooks",
			config: Configuration{
				PostCreateSandboxHooks:  []PostSandboxCreate{sandboxCreateHook},
				PostDestroySandboxHooks: []PostSandboxDestroy{sandboxDestroyHook},
			},
			client: fake.NewSimpleClientset(),
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
			},
			expectedProvider: &provider{
				coe:                 newOpenshift(),
				ExtractedCredential: defaultExtractedCredential{},
				postSandboxCreate:   []PostSandboxCreate{sandboxCreateHook},
				postSandboxDestroy:  []PostSandboxDestroy{sandboxDestroyHook},
			},
			dontUseDeepEqual: true,
		},
	}
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Fatalf("test panic unexpectedly: %#+v", r)
				}
			}()
			k.Client = &fakeClientSet{
				tc.client,
				&fakerest.RESTClient{
					Resp:                 tc.response,
					NegotiatedSerializer: scheme.Codecs,
				},
			}
			NewRuntime(tc.config)
			if !tc.dontUseDeepEqual {
				if !reflect.DeepEqual(tc.expectedProvider, Provider) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
			} else {
				p := Provider.(*provider)
				if len(p.preSandboxCreate) != len(tc.expectedProvider.preSandboxCreate) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
				if len(p.preSandboxDestroy) != len(tc.expectedProvider.preSandboxDestroy) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
				if len(p.postSandboxDestroy) != len(tc.expectedProvider.postSandboxDestroy) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
				if len(p.postSandboxCreate) != len(tc.expectedProvider.postSandboxCreate) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
				if !reflect.DeepEqual(tc.expectedProvider.coe, p.coe) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
				if !reflect.DeepEqual(tc.expectedProvider.ExtractedCredential, p.ExtractedCredential) {
					t.Fatalf("invalid provider for configuration: %#+v \n\n got: %#+v \n\n exp: %#+v", tc.config, Provider, tc.expectedProvider)
				}
			}
		})
	}
}
