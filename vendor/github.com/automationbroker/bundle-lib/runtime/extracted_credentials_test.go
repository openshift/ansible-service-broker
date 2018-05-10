package runtime

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/clients"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateExtractedCredential(t *testing.T) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name           string
		client         *fake.Clientset
		expectedSecret *v1.Secret
		id             string
		namespace      string
		extCreds       map[string]interface{}
		labels         map[string]string
		shouldError    bool
	}{
		{
			name:      "creates secret correctly",
			client:    fake.NewSimpleClientset(),
			namespace: "testing",
			id:        "xxxxx",
			labels: map[string]string{
				"action": "provision",
			},
			extCreds: map[string]interface{}{
				"hello": map[string]string{
					"test": "12",
				},
				"hey": 1,
			},
			expectedSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxxxx",
					Namespace: "testing",
					Labels: map[string]string{
						"action": "provision",
					},
					OwnerReferences: nil,
				},
				Data: map[string][]byte{"credentials": []byte(`{"hello":{"test":"12"},"hey":1}`)},
			},
		},
		{
			name: "unable to create secret",
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxxxx",
					Namespace: "testing",
					Labels: map[string]string{
						"action": "provision",
					},
					OwnerReferences: nil,
				},
				Data: map[string][]byte{"credentials": []byte(`{"hello":{"test":"12"},"hey":1}`)},
			},
			),
			namespace: "testing",
			id:        "xxxxx",
			labels: map[string]string{
				"action": "provision",
			},
			extCreds: map[string]interface{}{
				"hello": map[string]string{
					"test": "12",
				},
				"hey": 1,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := defaultExtractedCredential{}
			k8scli.Client = tc.client
			err := d.CreateExtractedCredential(tc.id, tc.namespace, tc.extCreds, tc.labels)
			if err == nil && tc.shouldError {
				t.Fatalf("error  was expected but not thrown")
				return
			}
			if err != nil && !tc.shouldError {
				t.Fatalf("err occured but was unexpected: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			sec, err := k8scli.Client.CoreV1().Secrets(tc.namespace).Get(tc.id, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("err occured but was unexpected: %v", err)
				return
			}
			fmt.Printf("%q, %q", sec.Data["credentials"], tc.expectedSecret.Data["credentials"])
			if !reflect.DeepEqual(sec.Data["credentials"], tc.expectedSecret.Data["credentials"]) {
				t.Fatalf("expected secret data: %#v did not match \n\nactual: %#v \nsecrets\n", tc.expectedSecret.Data["credentials"], sec.Data["credentials"])
				return
			}
			if !reflect.DeepEqual(sec.ObjectMeta, tc.expectedSecret.ObjectMeta) {
				t.Fatalf("expected object meta data:\n %#v did not match \n\nactual:\n %#v \nsecrets\n", tc.expectedSecret.ObjectMeta, sec.ObjectMeta)
				return
			}
		})
	}
}

func TestUpdateExtractedCredential(t *testing.T) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name           string
		client         *fake.Clientset
		expectedSecret *v1.Secret
		id             string
		namespace      string
		extCreds       map[string]interface{}
		labels         map[string]string
		shouldError    bool
	}{
		{
			name: "updates secret correctly",
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxxxx",
					Namespace: "testing",
					Labels: map[string]string{
						"action": "provision",
					},
					OwnerReferences: nil,
				},
				Data: map[string][]byte{"credentials": []byte(`{"hello":{"test":"12"},"hey":1}`)},
			}),
			namespace: "testing",
			id:        "xxxxx",
			labels: map[string]string{
				"action": "provision",
			},
			extCreds: map[string]interface{}{
				"hello": map[string]string{
					"test": "12",
				},
				"hey": 12,
			},
			expectedSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxxxx",
					Namespace: "testing",
					Labels: map[string]string{
						"action": "provision",
					},
					OwnerReferences: nil,
				},
				Data: map[string][]byte{"credentials": []byte(`{"hello":{"test":"12"},"hey":12}`)},
			},
		},
		{
			name:      "unable to update secret",
			client:    fake.NewSimpleClientset(),
			namespace: "testing",
			id:        "xxxxx",
			labels: map[string]string{
				"action": "provision",
			},
			extCreds: map[string]interface{}{
				"hello": map[string]string{
					"test": "12",
				},
				"hey": 1,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := defaultExtractedCredential{}
			k8scli.Client = tc.client
			err := d.UpdateExtractedCredential(tc.id, tc.namespace, tc.extCreds, tc.labels)
			if err == nil && tc.shouldError {
				t.Fatalf("error  was expected but not thrown")
				return
			}
			if err != nil && !tc.shouldError {
				t.Fatalf("err occured but was unexpected: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			sec, err := k8scli.Client.CoreV1().Secrets(tc.namespace).Get(tc.id, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("err occured but was unexpected: %v", err)
				return
			}
			fmt.Printf("%q, %q", sec.Data["credentials"], tc.expectedSecret.Data["credentials"])
			if !reflect.DeepEqual(sec.Data["credentials"], tc.expectedSecret.Data["credentials"]) {
				t.Fatalf("expected secret data: %#v did not match \n\nactual: %#v \nsecrets\n", tc.expectedSecret.Data["credentials"], sec.Data["credentials"])
				return
			}
			if !reflect.DeepEqual(sec.ObjectMeta, tc.expectedSecret.ObjectMeta) {
				t.Fatalf("expected object meta data:\n %#v did not match \n\nactual:\n %#v \nsecrets\n", tc.expectedSecret.ObjectMeta, sec.ObjectMeta)
				return
			}
		})
	}
}

func TestDeleteExtractedCredential(t *testing.T) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name        string
		client      *fake.Clientset
		id          string
		namespace   string
		shouldError bool
	}{
		{
			name: "delete secret correctly",
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxxxx",
					Namespace: "testing",
					Labels: map[string]string{
						"action": "provision",
					},
					OwnerReferences: nil,
				},
				Data: map[string][]byte{"credentials": []byte(`{"hello":{"test":"12"},"hey":1}`)},
			}),
			namespace: "testing",
			id:        "xxxxx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := defaultExtractedCredential{}
			k8scli.Client = tc.client
			err := d.DeleteExtractedCredential(tc.id, tc.namespace)
			if err == nil && tc.shouldError {
				t.Fatalf("error  was expected but not thrown")
				return
			}
			if err != nil && !tc.shouldError {
				t.Fatalf("err occured but was unexpected: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			_, err = k8scli.Client.CoreV1().Secrets(tc.namespace).Get(tc.id, metav1.GetOptions{})
			if err == nil {
				t.Fatalf("retrieved deleted secret")
				return
			}
			if !k8serrors.IsNotFound(err) {
				t.Fatalf("Unknown error occured : %v", err)
				return
			}
		})
	}
}

func TestGetExtractedCredential(t *testing.T) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name        string
		client      *fake.Clientset
		expected    map[string]interface{}
		id          string
		namespace   string
		shouldError bool
	}{
		{
			name: "gets secret correctly",
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xxxxx",
					Namespace: "testing",
					Labels: map[string]string{
						"action": "provision",
					},
					OwnerReferences: nil,
				},
				Data: map[string][]byte{"credentials": []byte(`{"hello":{"test":"12"},"hey":"1"}`)},
			}),
			namespace: "testing",
			id:        "xxxxx",
			expected: map[string]interface{}{
				"hello": map[string]interface{}{
					"test": "12",
				},
				"hey": "1",
			},
		},
		{
			name:        "unable to get secret",
			client:      fake.NewSimpleClientset(),
			namespace:   "testing",
			id:          "xxxxx",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := defaultExtractedCredential{}
			k8scli.Client = tc.client
			act, err := d.GetExtractedCredential(tc.id, tc.namespace)
			if err == nil && tc.shouldError {
				t.Fatalf("error  was expected but not thrown")
				return
			}
			if err != nil && !tc.shouldError {
				t.Fatalf("err occured but was unexpected: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			fmt.Printf("%#+v\n\n%#+v", act, tc.expected)
			if !reflect.DeepEqual(act, tc.expected) {
				t.Fatalf("actual did not match expected\nexpected: %v\nactual: %v", tc.expected, act)
			}
		})
	}
}
