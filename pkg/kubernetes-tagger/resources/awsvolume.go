package resources

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
	k8sClient        *kubernetes.Clientset
	volumeID         string
	log              *logrus.Entry
}

// Type Get type
func (av *AWSVolume) Type() string {
	return av.resourceType
}

// Platform Get platform
func (av *AWSVolume) Platform() string {
	return av.resourcePlatform
}

// CanBeProcessed Can be processed ?
func (av *AWSVolume) CanBeProcessed() bool {
	// It is always true in this case
	return true
}

// newAWSVolume Generate a new AWS Volume
func newAWSVolume(k8sClient *kubernetes.Clientset, pv *v1.PersistentVolume, config *config.Configuration) (*AWSVolume, error) {
	url, err := url.Parse(pv.Spec.AWSElasticBlockStore.VolumeID)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse persistent volume AWS Volume Id: %v", err)
	}
	volumeID := url.Path
	volumeID = strings.Trim(volumeID, "/")

	// Create logger
	log := logrus.WithFields(logrus.Fields{
		"type":                 AWSVolumeResourceType,
		"platform":             AWSResourcePlatform,
		"persistentVolumeName": pv.Name,
	})

	awsConfig := config.AWS
	instance := AWSVolume{
		resourceType:     AWSVolumeResourceType,
		resourcePlatform: AWSResourcePlatform,
		awsConfig:        awsConfig,
		persistentVolume: pv,
		k8sClient:        k8sClient,
		volumeID:         volumeID,
		log:              log,
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

// ManageTags Manage tags on resource
func (av *AWSVolume) ManageTags(delta *TagDelta) error {
	// TODO Need to check AWS Limits before sending
	av.log.WithField("delta", delta).Debug("Manage tags on resource")
	av.log.Info("Manage tags on resource")
	// Get EC2 AWS Client
	svc, err := av.getAWSEC2Client()
	if err != nil {
		return err
	}

	// Add case

	// Check if tags needs to be added
	if len(delta.AddList) > 0 {
		av.log.WithField("delta", delta).Debug("Add list detected. Begin request to AWS.")
		awsAddTags := make([]*ec2.Tag, 0)
		for i := 0; i < len(delta.AddList); i++ {
			tag := delta.AddList[i]
			awsAddTags = append(awsAddTags, &ec2.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
		}

		// Add tags to the created instance
		_, err = svc.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{aws.String(av.volumeID)},
			Tags:      awsAddTags,
		})
		if err != nil {
			return err
		}
		av.log.WithField("delta", delta).Debug("Add list successfully managed")
	}

	// Delete case

	// Check if tags needs to be removed
	if len(delta.DeleteList) > 0 {
		av.log.WithField("delta", delta).Debug("Delete list detected. Begin request to AWS.")
		awsDeleteTags := make([]*ec2.Tag, 0)
		for i := 0; i < len(delta.DeleteList); i++ {
			tag := delta.DeleteList[i]
			awsDeleteTags = append(awsDeleteTags, &ec2.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
		}

		// Delete tags to the created instance
		_, err = svc.DeleteTags(&ec2.DeleteTagsInput{
			Resources: []*string{aws.String(av.volumeID)},
			Tags:      awsDeleteTags,
		})
		if err != nil {
			return err
		}
		av.log.WithField("delta", delta).Debug("Delete list successfully managed")
	}

	return nil
}

// GetActualTags Get actual tags.
func (av *AWSVolume) GetActualTags() ([]*Tag, error) {
	av.log.Info("Get actual tags on resource")
	// Get EC2 AWS Client
	svc, err := av.getAWSEC2Client()
	if err != nil {
		return nil, err
	}

	volumesIDs := make([]*string, 0)
	volumesIDs = append(volumesIDs, aws.String(av.volumeID))
	output, err := svc.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: volumesIDs,
	})
	if err != nil {
		return nil, err
	}
	volumes := output.Volumes
	if len(volumes) != 1 {
		return nil, fmt.Errorf("Can't find volume in AWS from volume id \"%s\"", av.volumeID)
	}
	volume := volumes[0]

	result := make([]*Tag, 0)
	if volume.Tags == nil {
		return result, nil
	}

	for i := 0; i < len(volume.Tags); i++ {
		awsTag := volume.Tags[i]
		result = append(result, &Tag{Key: *awsTag.Key, Value: *awsTag.Value})
	}
	return result, nil
}

func (av *AWSVolume) getAWSEC2Client() (*ec2.EC2, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(av.awsConfig.Region)},
	)
	if err != nil {
		return nil, err
	}
	// Create EC2 service client
	svc := ec2.New(sess)
	return svc, nil
}
