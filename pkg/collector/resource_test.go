package collector

import (
	"context"
	"testing"

	helmapi "github.com/fluxcd/helm-controller/api/v2beta1"
	ksapi "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type NullLogger struct{}

func (l NullLogger) Debugf(format string, a ...interface{}) {
}

func (l NullLogger) Failuref(format string, a ...interface{}) {
}

func strPtr(str string) *string {
	s := str
	return &s
}

type test struct {
	name         string
	filters      func() []FilterFunc
	list         func() *unstructured.UnstructuredList
	expectedPass int
}

func TestDiscovery(t *testing.T) {
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
				helmReleases := []helmapi.HelmRelease{}
				hr := helmapi.HelmRelease{}
				hr.SetName("release")
				hr.SetNamespace("test")

				helmReleases = append(helmReleases, hr)

				return []FilterFunc{IgnoreIfHelmReleaseFound(helmReleases)}
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
			name: "A resource which is part of a kustomization but without a matching inventory entry is not ignored",
			filters: func() []FilterFunc {
				kustomizations := &ksapi.KustomizationList{}
				ks := ksapi.Kustomization{}
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
					fluxKustomizeNamespaceLabel: "test",
				})

				list.Items = append(list.Items, expected, alsoExpected)
				return list
			},
			expectedPass: 2,
		},
		{
			name: "A resource which is part of a kustomization and has a valid matching inventory entry is ignored",
			filters: func() []FilterFunc {
				kustomizations := &ksapi.KustomizationList{}
				ks := ksapi.Kustomization{}
				ks.SetName("release")
				ks.SetNamespace("test")
				ks.Status.Inventory = &ksapi.ResourceInventory{
					Entries: []ksapi.ResourceRef{
						{
							ID: "test_service-account-secret__Secret",
						},
					},
				}

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
					fluxKustomizeNamespaceLabel: "test",
				})

				notExpected := unstructured.Unstructured{}
				notExpected.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Secret",
				})
				notExpected.SetNamespace("test")
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
		{
			name: "A resource which is part of a kustomization but the kustomization was not found",
			filters: func() []FilterFunc {
				kustomizations := &ksapi.KustomizationList{}
				ks := ksapi.Kustomization{}
				ks.SetName("release")
				ks.SetNamespace("test")
				ks.Status.Inventory = &ksapi.ResourceInventory{
					Entries: []ksapi.ResourceRef{
						{
							ID: "test_service-account-secret__Secret",
						},
					},
				}

				kustomizations.Items = append(kustomizations.Items, ks)

				return []FilterFunc{IgnoreIfKustomizationFound(kustomizations.Items)}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("service-account-secret")
				expected.SetLabels(map[string]string{
					fluxKustomizeNameLabel:      "does-not-exists",
					fluxKustomizeNamespaceLabel: "does-not-exists",
				})

				notExpected := unstructured.Unstructured{}
				notExpected.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Secret",
				})
				notExpected.SetNamespace("test")
				notExpected.SetName("service-account-secret")
				notExpected.SetLabels(map[string]string{
					fluxKustomizeNameLabel:      "release",
					fluxKustomizeNamespaceLabel: "test",
				})

				list.Items = append(list.Items, expected, notExpected)
				return list
			},
			expectedPass: 1,
		},
		{
			name: "Resources are excluded from conf",
			filters: func() []FilterFunc {
				return []FilterFunc{IgnoreRuleExclusions([]Exclusion{
					{
						Name:        strPtr("default"),
						Description: strPtr("exclude default service account"),
						Kind: &schema.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "ServiceAccount",
						},
					},
					{
						Name:        strPtr("velero-capi-backup-.*"),
						Namespace:   strPtr("velero"),
						Description: strPtr("velero"),
						Kind: &schema.GroupVersionKind{
							Group:   "velero.io",
							Version: "v1",
							Kind:    "Backup",
						},
					},
				})}
			},
			list: func() *unstructured.UnstructuredList {
				list := &unstructured.UnstructuredList{}
				expected := unstructured.Unstructured{}
				expected.SetName("default")
				expected.SetNamespace("default")
				expected.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "ServiceAccount",
				})

				alsoExpected := unstructured.Unstructured{}
				alsoExpected.SetName("velero-capi-backup-20230119201916")
				alsoExpected.SetNamespace("velero")
				alsoExpected.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "velero.io",
					Version: "v1",
					Kind:    "Backup",
				})

				notExpected := unstructured.Unstructured{}
				notExpected.SetName("Deployment")
				notExpected.SetNamespace("default")
				notExpected.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				})

				list.Items = append(list.Items, expected, alsoExpected, notExpected)
				return list
			},
			expectedPass: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan unstructured.Unstructured, test.expectedPass+1)
			discovery := NewDiscovery(klog.NewKlogr(), test.filters()...)
			err := discovery.Discover(context.TODO(), test.list(), ch)
			require.NoError(t, err)
			assert.Equal(t, test.expectedPass, len(ch))
		})
	}
}
