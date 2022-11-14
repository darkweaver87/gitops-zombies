package collector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type NullLogger struct{}

func (l NullLogger) Debugf(format string, a ...interface{}) {
}

func (l NullLogger) Failuref(format string, a ...interface{}) {
}

type test struct {
	name         string
	filters      func() []FilterFunc
	list         func() *unstructured.UnstructuredList
	expectedPass int
}

func TestDisovery(t *testing.T) {
	tests := []test{
		{
			name: "A resource which has owner references is skipped",
			filters: func() []FilterFunc {
				return []FilterFunc{IgnoreOwnedResource()}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("resource-without-owner")

				notExpected := unstructured.Unstructured{}
				notExpected.SetName("resource-with-owner")
				notExpected.SetOwnerReferences([]v1.OwnerReference{
					{
						Name: "owner",
					},
				})

				list.Items = append(list.Items, expected, notExpected)
				return list
			},
			expectedPass: 1,
		},
		{
			name: "A secret which belongs to a service account is ignored",
			filters: func() []FilterFunc {
				return []FilterFunc{IgnoreServiceAccountSecret()}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("secret")
				expected.SetAPIVersion("v1")
				expected.SetKind("Secret")

				notExpected := unstructured.Unstructured{}
				notExpected.SetName("service-account-secret")
				notExpected.SetAPIVersion("v1")
				notExpected.SetKind("Secret")
				notExpected.SetAnnotations(map[string]string{
					"kubernetes.io/service-account.name": "sa",
				})

				list.Items = append(list.Items, expected, notExpected)
				return list
			},
			expectedPass: 1,
		},
		{
			name: "A secret which is labeled as a helm owner is ignored",
			filters: func() []FilterFunc {
				return []FilterFunc{IgnoreHelmSecret()}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("secret")
				expected.SetAPIVersion("v1")
				expected.SetKind("Secret")

				notExpected := unstructured.Unstructured{}
				notExpected.SetName("service-account-secret")
				notExpected.SetAPIVersion("v1")
				notExpected.SetKind("Secret")
				notExpected.SetLabels(map[string]string{
					"owner": "helm",
				})

				list.Items = append(list.Items, expected, notExpected)
				return list
			},
			expectedPass: 1,
		},
		{
			name: "A resource which is part of a helmrelease is ignored",
			filters: func() []FilterFunc {
				helmReleases := &unstructured.UnstructuredList{}
				hr := unstructured.Unstructured{}
				hr.SetName("release")
				hr.SetNamespace("test")

				helmReleases.Items = append(helmReleases.Items, hr)

				return []FilterFunc{IgnoreIfHelmReleaseFound(helmReleases.Items)}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("resource")

				alsoExpected := unstructured.Unstructured{}
				alsoExpected.SetName("service-account-secret")
				alsoExpected.SetLabels(map[string]string{
					fluxHelmNameLabel:      "release",
					fluxHelmNamespaceLabel: "not-existing",
				})

				notExpected := unstructured.Unstructured{}
				notExpected.SetName("service-account-secret")
				notExpected.SetLabels(map[string]string{
					fluxHelmNameLabel:      "release",
					fluxHelmNamespaceLabel: "test",
				})

				list.Items = append(list.Items, expected, alsoExpected, notExpected)
				return list
			},
			expectedPass: 2,
		},
		{
			name: "A resource which is part of a kustomization is ignored",
			filters: func() []FilterFunc {
				kustomizations := &unstructured.UnstructuredList{}
				ks := unstructured.Unstructured{}
				ks.SetName("release")
				ks.SetNamespace("test")

				kustomizations.Items = append(kustomizations.Items, ks)

				return []FilterFunc{IgnoreIfKustomizationFound(kustomizations.Items)}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("resource")

				alsoExpected := unstructured.Unstructured{}
				alsoExpected.SetName("service-account-secret")
				alsoExpected.SetLabels(map[string]string{
					fluxKustomizeNameLabel:      "release",
					fluxKustomizeNamespaceLabel: "not-existing",
				})

				notExpected := unstructured.Unstructured{}
				notExpected.SetName("service-account-secret")
				notExpected.SetLabels(map[string]string{
					fluxKustomizeNameLabel:      "release",
					fluxKustomizeNamespaceLabel: "test",
				})

				list.Items = append(list.Items, expected, alsoExpected, notExpected)
				return list
			},
			expectedPass: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan unstructured.Unstructured, test.expectedPass+1)
			discovery := NewDiscovery(&NullLogger{}, test.filters()...)
			err := discovery.Discover(context.TODO(), test.list(), ch)
			require.NoError(t, err)
			assert.Equal(t, test.expectedPass, len(ch))
		})
	}
}
