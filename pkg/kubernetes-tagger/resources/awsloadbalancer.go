package resources

import (
	"strings"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/providerClient"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// AWSLoadBalancer AWS Load Balancer
type AWSLoadBalancer struct {
	resourceType     string
	resourcePlatform string
	awsConfig        *config.AWSConfig
	service          *v1.Service
	k8sClient        kubernetes.Interface
	volumeID         string
	log              *logrus.Entry
	prcl             *providerclient.AWSProviderClient
}

// Type Get type
func (al *AWSLoadBalancer) Type() string {
	return al.resourceType
}

// Platform Get platform
func (al *AWSLoadBalancer) Platform() string {
	return al.resourcePlatform
}

// newAWSLoadBalancer Generate a new AWS Load Balancer
func newAWSLoadBalancer(k8sClient kubernetes.Interface, svc *v1.Service, config *config.Configuration, prcl providerclient.ProviderClient) (*AWSLoadBalancer, error) {
	// Create logger
	log := logrus.WithFields(logrus.Fields{
		"type":        LoadBalancerResourceType,
		"platform":    AWSResourcePlatform,
		"serviceName": svc.Name,
	})

	awsConfig := config.AWS
	instance := AWSLoadBalancer{
		resourceType:     LoadBalancerResourceType,
		resourcePlatform: AWSResourcePlatform,
		awsConfig:        awsConfig,
		service:          svc,
		k8sClient:        k8sClient,
		log:              log,
		prcl:             prcl.(*providerclient.AWSProviderClient),
	}
	return &instance, nil
}

// isAWSLoadBalancerResource returns a boolean to know if a service is an AWS Load Balancer
func isAWSLoadBalancerResource(svc *v1.Service) bool {
	if svc == nil {
		return false
	}
	// Check that svc is a load balancer
	if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
		return false
	}
	if svc.Status.LoadBalancer.Ingress == nil || len(svc.Status.LoadBalancer.Ingress) == 0 {
		return false
	}
	// Get ingress
	ing := svc.Status.LoadBalancer.Ingress[0]
	if ing.Hostname == "" {
		return false
	}
	return strings.HasSuffix(ing.Hostname, "amazonaws.com")
}

// GetAvailableTagValues Get available tag values
func (al *AWSLoadBalancer) GetAvailableTagValues() (map[string]interface{}, error) {
	// Begin to create available tag values
	availableTags := make(map[string]interface{})
	availableTags["type"] = al.Type()
	availableTags["platform"] = al.Platform()
	svcTags := make(map[string]interface{})
	svcTags["name"] = al.service.Name
	svcTags["namespace"] = al.service.Namespace
	svcTags["annotations"] = al.service.Annotations
	svcTags["labels"] = al.service.Labels
	availableTags["service"] = svcTags
	return availableTags, nil
}

// GetActualTags Get actual tags
func (al *AWSLoadBalancer) GetActualTags() ([]*tags.Tag, error) {
	al.log.Info("Get actual tags on resource")
	return al.prcl.GetActualTagsFromService(al.service)
}

// ManageTags Manage tags
func (al *AWSLoadBalancer) ManageTags(delta *tags.TagDelta) error {
	al.log.WithField("delta", delta).Debug("Manage tags on resource")
	al.log.Info("Manage tags on resource")

	// Check if tags needs to be added
	if len(delta.AddList) > 0 {
		al.log.WithField("delta", delta).Debug("Add list detected. Begin request to AWS.")
		err := al.prcl.AddTagsFromService(al.service, delta.AddList)
		if err != nil {
			return err
		}
		al.log.WithField("delta", delta).Debug("Add list successfully managed")
	}

	// Delete case

	// Check if tags needs to be removed
	if len(delta.DeleteList) > 0 {
		al.log.WithField("delta", delta).Debug("Delete list detected. Begin request to AWS.")
		err := al.prcl.DeleteTagsFromService(al.service, delta.DeleteList)
		if err != nil {
			return err
		}
		al.log.WithField("delta", delta).Debug("Delete list successfully managed")
	}

	return nil
}
