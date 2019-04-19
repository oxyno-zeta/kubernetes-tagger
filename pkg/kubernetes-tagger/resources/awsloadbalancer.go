package resources

import (
	"errors"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ServiceAnnotationLoadBalancerType is the annotation used on the service
// to indicate what type of Load Balancer we want. Right now, the only accepted
// value is "nlb"
// COPIED FROM https://github.com/kubernetes/kubernetes/blob/d7103187a37dcfff79077c80a151e98571487628/pkg/cloudprovider/providers/aws/aws.go
const ServiceAnnotationLoadBalancerType = "service.beta.kubernetes.io/aws-load-balancer-type"

// ErrLoadBalancerNotFound Load Balancer Not Found
var ErrLoadBalancerNotFound = errors.New("load balancer not found")

// ErrNoTagsFound No tags found error
var ErrNoTagsFound = errors.New("no tags found on load balancer")

// AWSLoadBalancer AWS Load Balancer
type AWSLoadBalancer struct {
	resourceType     string
	resourcePlatform string
	awsConfig        *config.AWSConfig
	service          *v1.Service
	k8sClient        kubernetes.Interface
	volumeID         string
	log              *logrus.Entry
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
func newAWSLoadBalancer(k8sClient kubernetes.Interface, svc *v1.Service, config *config.Configuration) (*AWSLoadBalancer, error) {
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
	}
	return &instance, nil
}

// isAWSLoadBalancerResource returns a boolean to know if a service is an AWS Load Balancer
func isAWSLoadBalancerResource(svc *v1.Service) bool {
	// The only thing that can be tested here is the service type
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
	return strings.Contains(ing.Hostname, "amazonaws.com")
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
func (al *AWSLoadBalancer) GetActualTags() ([]*Tag, error) {
	sess, err := getAWSSession(al.awsConfig)
	if err != nil {
		return nil, err
	}

	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(al.service)

	result := make([]*Tag, 0)

	// Check if it is a network loadbalancer (elbv2) or classic load balancer (elb)
	if al.service.Annotations != nil && al.service.Annotations[ServiceAnnotationLoadBalancerType] == "nlb" {
		// elbv2 detected
		svc := elbv2.New(sess)

		output, err := svc.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{
			Names: []*string{
				aws.String(name),
			},
		})
		if err != nil {
			return nil, err
		}

		if len(output.LoadBalancers) == 0 {
			return nil, ErrLoadBalancerNotFound
		}

		lb := output.LoadBalancers[0]

		describeTagsOutput, err := svc.DescribeTags(&elbv2.DescribeTagsInput{
			ResourceArns: []*string{
				lb.LoadBalancerArn,
			},
		})
		if err != nil {
			return nil, err
		}

		// Check response size
		if len(describeTagsOutput.TagDescriptions) == 0 {
			return nil, ErrNoTagsFound
		}

		awsTags := describeTagsOutput.TagDescriptions[0].Tags

		if awsTags == nil {
			return result, nil
		}

		for i := 0; i < len(awsTags); i++ {
			awsTag := awsTags[i]
			result = append(result, &Tag{Key: *awsTag.Key, Value: *awsTag.Value})
		}

		return result, nil
	}

	// elb normal detected
	svc := elb.New(sess)

	describeTagsOutput, err := svc.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: []*string{
			aws.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	// Check response size
	if len(describeTagsOutput.TagDescriptions) == 0 {
		return nil, ErrNoTagsFound
	}

	awsTags := describeTagsOutput.TagDescriptions[0].Tags

	if awsTags == nil {
		return result, nil
	}

	for i := 0; i < len(awsTags); i++ {
		awsTag := awsTags[i]
		result = append(result, &Tag{Key: *awsTag.Key, Value: *awsTag.Value})
	}

	return result, nil
}

// ManageTags Manage tags
func (al *AWSLoadBalancer) ManageTags(delta *TagDelta) error {
	al.log.WithField("delta", delta).Debug("Manage tags on resource")
	al.log.Info("Manage tags on resource")

	sess, err := getAWSSession(al.awsConfig)
	if err != nil {
		return err
	}

	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(al.service)

	// Check if it is a network loadbalancer (elbv2) or classic load balancer (elb)
	if al.service.Annotations != nil && al.service.Annotations[ServiceAnnotationLoadBalancerType] == "nlb" {
		// elbv2 detected
		svc := elbv2.New(sess)

		output, err := svc.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{
			Names: []*string{
				aws.String(name),
			},
		})
		if err != nil {
			return err
		}

		if len(output.LoadBalancers) == 0 {
			return ErrLoadBalancerNotFound
		}

		lb := output.LoadBalancers[0]

		// Add case

		// Check if tags needs to be added
		if len(delta.AddList) > 0 {
			al.log.WithField("delta", delta).Debug("Add list detected. Begin request to AWS.")
			awsAddTags := make([]*elbv2.Tag, 0)
			for i := 0; i < len(delta.AddList); i++ {
				tag := delta.AddList[i]
				awsAddTags = append(awsAddTags, &elbv2.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
			}

			// Add tags to the created instance
			_, err = svc.AddTags(&elbv2.AddTagsInput{
				ResourceArns: []*string{
					lb.LoadBalancerArn,
				},
				Tags: awsAddTags,
			})
			if err != nil {
				return err
			}
			al.log.WithField("delta", delta).Debug("Add list successfully managed")
		}

		// Delete case

		// Check if tags needs to be removed
		if len(delta.DeleteList) > 0 {
			al.log.WithField("delta", delta).Debug("Delete list detected. Begin request to AWS.")
			awsDeleteTags := make([]*string, 0)
			for i := 0; i < len(delta.DeleteList); i++ {
				tag := delta.DeleteList[i]
				awsDeleteTags = append(awsDeleteTags, aws.String(tag.Key))
			}

			// Delete tags to the created instance
			_, err = svc.RemoveTags(&elbv2.RemoveTagsInput{
				ResourceArns: []*string{
					lb.LoadBalancerArn,
				},
				TagKeys: awsDeleteTags,
			})
			if err != nil {
				return err
			}
			al.log.WithField("delta", delta).Debug("Delete list successfully managed")
		}

		return nil
	}

	svc := elb.New(sess)
	// Add case

	// Check if tags needs to be added
	if len(delta.AddList) > 0 {
		al.log.WithField("delta", delta).Debug("Add list detected. Begin request to AWS.")
		awsAddTags := make([]*elb.Tag, 0)
		for i := 0; i < len(delta.AddList); i++ {
			tag := delta.AddList[i]
			awsAddTags = append(awsAddTags, &elb.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
		}

		// Add tags to the created instance
		_, err = svc.AddTags(&elb.AddTagsInput{
			LoadBalancerNames: []*string{
				aws.String(name),
			},
			Tags: awsAddTags,
		})
		if err != nil {
			return err
		}
		al.log.WithField("delta", delta).Debug("Add list successfully managed")
	}

	// Delete case

	// Check if tags needs to be removed
	if len(delta.DeleteList) > 0 {
		al.log.WithField("delta", delta).Debug("Delete list detected. Begin request to AWS.")
		awsDeleteTags := make([]*elb.TagKeyOnly, 0)
		for i := 0; i < len(delta.DeleteList); i++ {
			tag := delta.DeleteList[i]
			awsDeleteTags = append(awsDeleteTags, &elb.TagKeyOnly{Key: aws.String(tag.Key)})
		}

		// Delete tags to the created instance
		_, err = svc.RemoveTags(&elb.RemoveTagsInput{
			LoadBalancerNames: []*string{
				aws.String(name),
			},
			Tags: awsDeleteTags,
		})
		if err != nil {
			return err
		}
		al.log.WithField("delta", delta).Debug("Delete list successfully managed")
	}

	return nil
}

func getAWSLoadBalancerName(svc *v1.Service) string {
	// Split hostname on . and after split the first part on -
	splitHostname := strings.Split(svc.Status.LoadBalancer.Ingress[0].Hostname, ".")
	splitSubDomain := strings.Split(splitHostname[0], "-")
	name := splitSubDomain[0]
	// Don't take the last one
	for i := 1; i < len(splitSubDomain)-1; i++ {
		name = name + "-" + splitSubDomain[i]
	}
	return name
}
