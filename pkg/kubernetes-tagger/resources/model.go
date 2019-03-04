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
	CanBeProcessed() bool
	CheckIfConfigurationValid() error
	GetAvailableTagValues() (map[string]interface{}, error)
	GetActualTags() ([]*Tag, error)
	ManageTags(delta *TagDelta) error
}

// New New resource instance
func New(k8sClient *kubernetes.Clientset, pv *v1.PersistentVolume, config *config.Configuration) (Resource, error) {
	if isAWSVolumeResource(pv) {
		res, err := newAWSVolume(k8sClient, pv, config)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, nil
}
