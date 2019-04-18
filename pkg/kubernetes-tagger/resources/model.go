package resources

import (
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Tag Tag structure
type Tag struct {
	Key   string
	Value string
}

// TagDelta Tag delta with to add and to delete tag lists
type TagDelta struct {
	AddList    []*Tag
	DeleteList []*Tag
}

// Resource Resource interface for all type of data
type Resource interface {
	Type() string
	Platform() string
	GetAvailableTagValues() (map[string]interface{}, error)
	GetActualTags() ([]*Tag, error)
	ManageTags(delta *TagDelta) error
}

// NewFromPersistentVolume New resource instance from persistent volume
func NewFromPersistentVolume(k8sClient *kubernetes.Clientset, pv *v1.PersistentVolume, cfg *config.Configuration) (Resource, error) {
	// Check if AWS provider is enabled and if it is an aws volume resource
	if cfg.Provider == config.AWSProviderName && isAWSVolumeResource(pv) {
		res, err := newAWSVolume(k8sClient, pv, cfg)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, nil
}

// NewFromService New resource instance from service
func NewFromService(k8sClient *kubernetes.Clientset, svc *v1.Service, cfg *config.Configuration) (Resource, error) {
	// Check if AWS provider is enabled and if it is an aws volume resource
	if cfg.Provider == config.AWSProviderName && isAWSLoadBalancerResource(svc) {
		res, err := newAWSLoadBalancer(k8sClient, svc, cfg)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, nil
}
