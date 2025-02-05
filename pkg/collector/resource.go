package collector

import (
	"context"
	"fmt"

	ksapi "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	fluxHelmNameLabel           = "helm.toolkit.fluxcd.io/name"
	fluxHelmNamespaceLabel      = "helm.toolkit.fluxcd.io/namespace"
	fluxKustomizeNameLabel      = "kustomize.toolkit.fluxcd.io/name"
	fluxKustomizeNamespaceLabel = "kustomize.toolkit.fluxcd.io/namespace"
)

// FilterFunc is a function that filters resources.
type FilterFunc func(res unstructured.Unstructured, logger logger) bool

// Interface represents collector interface.
type Interface interface {
	Discover(ctx context.Context, list *unstructured.UnstructuredList, ch chan unstructured.Unstructured) error
}

type logger interface {
	Debugf(format string, a ...interface{})
}

type discovery struct {
	filters []FilterFunc
	logger  logger
}

// NewDiscovery returns a new discovery instance.
func NewDiscovery(logger logger, filters ...FilterFunc) Interface {
	return &discovery{
		logger:  logger,
		filters: filters,
	}
}

// Discover validates discovered resources against all filters and adds it to consumer channel.
func (d *discovery) Discover(ctx context.Context, list *unstructured.UnstructuredList, ch chan unstructured.Unstructured) error {
RESOURCES:
	for _, res := range list.Items {
		d.logger.Debugf("validate resource %s %s %s", res.GetName(), res.GetNamespace(), res.GetAPIVersion())

		for _, filter := range d.filters {
			if filter(res, d.logger) {
				continue RESOURCES
			}
		}

		ch <- res
	}

	return nil
}

// IgnoreOwnedResource returns a FilterFunc which filters resources owner by parents ones.
func IgnoreOwnedResource() FilterFunc {
	return func(res unstructured.Unstructured, logger logger) bool {
		if refs := res.GetOwnerReferences(); len(refs) > 0 {
			logger.Debugf("ignore resource owned by parent %s %s %s", res.GetName(), res.GetNamespace(), res.GetAPIVersion())
			return true
		}

		return false
	}
}

// IgnoreServiceAccountSecret returns a FilterFunc which filters secrets linked to a service account.
func IgnoreServiceAccountSecret() FilterFunc {
	return func(res unstructured.Unstructured, logger logger) bool {
		if res.GetKind() == "Secret" && res.GetAPIVersion() == "v1" {
			if _, ok := res.GetAnnotations()["kubernetes.io/service-account.name"]; ok {
				return true
			}
		}

		return false
	}
}

// IgnoreHelmSecret returns a FilterFunc which filters secrets owned by helm.
func IgnoreHelmSecret() FilterFunc {
	return func(res unstructured.Unstructured, logger logger) bool {
		if res.GetKind() == "Secret" && res.GetAPIVersion() == "v1" {
			if v, ok := res.GetLabels()["owner"]; ok && v == "helm" {
				return true
			}
		}

		return false
	}
}

// IgnoreIfHelmReleaseFound returns a FilterFunc which filters resources part of an helm release.
func IgnoreIfHelmReleaseFound(helmReleases []unstructured.Unstructured) FilterFunc {
	return func(res unstructured.Unstructured, logger logger) bool {
		labels := res.GetLabels()
		if helmName, ok := labels[fluxHelmNameLabel]; ok {
			if helmNamespace, ok := labels[fluxHelmNamespaceLabel]; ok {
				if hasResource(helmReleases, helmName, helmNamespace) {
					return true
				}
				logger.Debugf("helmrelease [%s.%s] not found from resource  %s %s %s\n", helmName, helmNamespace, res.GetName(), res.GetNamespace(), res.GetAPIVersion())
			}
		}

		return false
	}
}

// IgnoreIfKustomizationFound returns a FilterFunc which filters resources part of a flux kustomization.
func IgnoreIfKustomizationFound(kustomizations []ksapi.Kustomization) FilterFunc {
	return func(res unstructured.Unstructured, logger logger) bool {
		labels := res.GetLabels()
		ksName, okKsName := labels[fluxKustomizeNameLabel]
		ksNamespace, okKsNamespace := labels[fluxKustomizeNamespaceLabel]
		if !okKsName || !okKsNamespace {
			return false
		}

		if ks := findKustomization(kustomizations, ksName, ksNamespace); ks != nil {
			id := fmt.Sprintf("%s_%s_%s_%s", res.GetNamespace(), res.GetName(), res.GroupVersionKind().Group, res.GroupVersionKind().Kind)
			logger.Debugf("lookup kustomization [%s.%s] inventory for %s", ksName, ksNamespace, id)

			if ks.Status.Inventory != nil {
				for _, entry := range ks.Status.Inventory.Entries {
					if entry.ID == id {
						return true
					}
				}
			}

			logger.Debugf("resource %s %s %s is not part of the kustomization [%s.%s] inventory", res.GetName(), res.GetNamespace(), res.GetAPIVersion(), ksName, ksNamespace)
			return false
		}
		logger.Debugf("kustomization [%s.%s] not found from resource  %s %s %s\n", ksName, ksNamespace, res.GetName(), res.GetNamespace(), res.GetAPIVersion())
		return false
	}
}

func hasResource(pool []unstructured.Unstructured, name, namespace string) bool {
	for _, res := range pool {
		if res.GetName() == name && res.GetNamespace() == namespace {
			return true
		}
	}

	return false
}

func findKustomization(pool []ksapi.Kustomization, name, namespace string) *ksapi.Kustomization {
	for _, res := range pool {
		if res.GetName() == name && res.GetNamespace() == namespace {
			return &res
		}
	}

	return nil
}
