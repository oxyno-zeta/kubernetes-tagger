package resources

import (
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/providerClient"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"

	"github.com/Sirupsen/logrus"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// AWSVolume AWS Volume
type AWSVolume struct {
	resourceType     string
	resourcePlatform string
	awsConfig        *config.AWSConfig
	persistentVolume *v1.PersistentVolume
	k8sClient        kubernetes.Interface
	log              *logrus.Entry
	prcl             *providerclient.AWSProviderClient
}

// Type Get type
func (av *AWSVolume) Type() string {
	return av.resourceType
}

// Platform Get platform
func (av *AWSVolume) Platform() string {
	return av.resourcePlatform
}

// newAWSVolume Generate a new AWS Volume
func newAWSVolume(k8sClient kubernetes.Interface, pv *v1.PersistentVolume, config *config.Configuration, prcl providerclient.ProviderClient) (*AWSVolume, error) {
	// Create logger
	log := logrus.WithFields(logrus.Fields{
		"type":                 VolumeResourceType,
		"platform":             AWSResourcePlatform,
		"persistentVolumeName": pv.Name,
	})

	awsConfig := config.AWS
	instance := AWSVolume{
		resourceType:     VolumeResourceType,
		resourcePlatform: AWSResourcePlatform,
		awsConfig:        awsConfig,
		persistentVolume: pv,
		k8sClient:        k8sClient,
		log:              log,
		prcl:             prcl.(*providerclient.AWSProviderClient),
	}
	return &instance, nil
}

// isAWSVolumeResource returns a boolean to know if a persistent volume is an AWS Volume
func isAWSVolumeResource(pv *v1.PersistentVolume) bool {
	return pv.Spec.AWSElasticBlockStore != nil
}

// GetAvailableTagValues Get available tags
func (av *AWSVolume) GetAvailableTagValues() (map[string]interface{}, error) {
	pvc, err := getPersistentVolumeClaim(av.persistentVolume, av.k8sClient)
	if err != nil {
		return nil, err
	}

	// Begin to create available tag values
	availableTags := make(map[string]interface{})
	availableTags["type"] = av.Type()
	availableTags["platform"] = av.Platform()
	pvTags := make(map[string]interface{})
	pvTags["labels"] = av.persistentVolume.Labels
	pvTags["annotations"] = av.persistentVolume.Annotations
	pvTags["name"] = av.persistentVolume.Name
	pvTags["phase"] = av.persistentVolume.Status.Phase
	pvTags["reclaimpolicy"] = av.persistentVolume.Spec.PersistentVolumeReclaimPolicy
	pvTags["storageclassname"] = av.persistentVolume.Spec.StorageClassName
	availableTags["persistentvolume"] = pvTags

	// If pvc exists, create tag values
	if pvc != nil {
		pvcTags := make(map[string]interface{})
		pvcTags["labels"] = pvc.Labels
		pvcTags["annotations"] = pvc.Annotations
		pvcTags["namespace"] = pvc.Namespace
		pvcTags["name"] = pvc.Name
		pvcTags["phase"] = pvc.Status.Phase
		availableTags["persistentvolumeclaim"] = pvcTags
	}

	return availableTags, nil
}

// GetActualTags Get actual tags.
func (av *AWSVolume) GetActualTags() ([]*tags.Tag, error) {
	av.log.Info("Get actual tags on resource")
	return av.prcl.GetActualTagsFromPersistentVolume(av.persistentVolume)
}

// ManageTags Manage tags on resource
func (av *AWSVolume) ManageTags(delta *tags.TagDelta) error {
	// TODO Need to check AWS Limits before sending
	av.log.WithField("delta", delta).Debug("Manage tags on resource")
	av.log.Info("Manage tags on resource")

	// Add case

	// Check if tags needs to be added
	if len(delta.AddList) > 0 {
		av.log.WithField("delta", delta).Debug("Add list detected. Begin request to AWS.")
		err := av.prcl.AddTagsFromPersistentVolume(av.persistentVolume, delta.AddList)
		if err != nil {
			return err
		}
		av.log.WithField("delta", delta).Debug("Add list successfully managed")
	}

	// Delete case

	// Check if tags needs to be removed
	if len(delta.DeleteList) > 0 {
		av.log.WithField("delta", delta).Debug("Delete list detected. Begin request to AWS.")
		err := av.prcl.DeleteTagsFromPersistentVolume(av.persistentVolume, delta.DeleteList)
		if err != nil {
			return err
		}
		av.log.WithField("delta", delta).Debug("Delete list successfully managed")
	}

	return nil
}
