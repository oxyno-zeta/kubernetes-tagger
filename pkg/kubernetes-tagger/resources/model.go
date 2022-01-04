package resources

import (
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	providerclient "github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/providerClient"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Resource Resource interface for all type of data.
type Resource interface {
	Type() string
	Platform() string
	GetAvailableTagValues() (map[string]interface{}, error)
	GetActualTags() ([]*tags.Tag, error)
	ManageTags(delta *tags.TagDelta) error
}

// NewFromPersistentVolume New resource instance from persistent volume.
func NewFromPersistentVolume(k8sClient kubernetes.Interface, pv *v1.PersistentVolume, cfg *config.Configuration) (Resource, error) {
	// Check if AWS provider is enabled
	if cfg.Provider == config.AWSProviderName {
		// Create Provider client
		prcl, err := providerclient.NewProviderClient(cfg)
		if err != nil {
			return nil, err
		}
		// Check if it is an aws volume resource
		if isAWSVolumeResource(pv) {
			res, err := newAWSVolume(k8sClient, pv, cfg, prcl)
			if err != nil {
				return nil, err
			}

			return res, nil
		}
	}

	return nil, nil //nolint:nilnil // Not needed
}

// NewFromService New resource instance from service.
func NewFromService(k8sClient kubernetes.Interface, svc *v1.Service, cfg *config.Configuration) (Resource, error) {
	// Check if AWS provider is enabled
	if cfg.Provider == config.AWSProviderName {
		// Create Provider client
		prcl, err := providerclient.NewProviderClient(cfg)
		if err != nil {
			return nil, err
		}
		// Check if it is an aws volume resource
		if isAWSLoadBalancerResource(svc) {
			res, err := newAWSLoadBalancer(k8sClient, svc, cfg, prcl)
			if err != nil {
				return nil, err
			}

			return res, nil
		}
	}

	return nil, nil //nolint:nilnil // Not needed
}
