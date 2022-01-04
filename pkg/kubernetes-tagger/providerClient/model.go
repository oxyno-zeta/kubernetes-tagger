package providerclient

import (
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"
	v1 "k8s.io/api/core/v1"
)

// ProviderClient Provider Client.
type ProviderClient interface {
	GetActualTagsFromPersistentVolume(pv *v1.PersistentVolume) ([]*tags.Tag, error)
	GetActualTagsFromService(svc *v1.Service) ([]*tags.Tag, error)
	AddTagsFromPersistentVolume(pv *v1.PersistentVolume, tagsList []*tags.Tag) error
	DeleteTagsFromPersistentVolume(pv *v1.PersistentVolume, tagsList []*tags.Tag) error
	AddTagsFromService(svc *v1.Service, tagsList []*tags.Tag) error
	DeleteTagsFromService(svc *v1.Service, tagsList []*tags.Tag) error
}

// NewProviderClient New Provider client.
func NewProviderClient(cfg *config.Configuration) (ProviderClient, error) {
	return newAWSProviderClient(cfg.AWS)
}
