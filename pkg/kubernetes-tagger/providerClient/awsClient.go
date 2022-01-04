package providerclient

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	v1 "k8s.io/api/core/v1"
)

const KubernetesAnnotationsNLBValue = "nlb"

// ServiceAnnotationLoadBalancerType is the annotation used on the service
// to indicate what type of Load Balancer we want. Right now, the only accepted
// value is "nlb"
// COPIED FROM https://github.com/kubernetes/kubernetes/blob/d7103187a37dcfff79077c80a151e98571487628/pkg/cloudprovider/providers/aws/aws.go
const ServiceAnnotationLoadBalancerType = "service.beta.kubernetes.io/aws-load-balancer-type"

// ErrLoadBalancerNotFound Load Balancer Not Found.
var ErrLoadBalancerNotFound = errors.New("load balancer not found")

// ErrNoTagsFound No tags found error.
var ErrNoTagsFound = errors.New("no tags found on load balancer")

// AWSProviderClient Aws Provider client.
type AWSProviderClient struct {
	awsConfig   *config.AWSConfig
	ec2client   *ec2.EC2
	elbclient   *elb.ELB
	elbv2client *elbv2.ELBV2
}

func newAWSProviderClient(awsConfig *config.AWSConfig) (*AWSProviderClient, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsConfig.Region)},
	)
	if err != nil {
		return nil, err
	}
	// Create EC2 service client
	ec2client := ec2.New(sess)
	// Create ELB service client
	elbclient := elb.New(sess)
	// Create ELBV2 service client
	elbv2client := elbv2.New(sess)

	// Create aws provider client
	cl := &AWSProviderClient{
		awsConfig:   awsConfig,
		ec2client:   ec2client,
		elbclient:   elbclient,
		elbv2client: elbv2client,
	}

	return cl, nil
}

func getVolumeIDFromPersistentVolume(pv *v1.PersistentVolume) (string, error) {
	url, err := url.Parse(pv.Spec.AWSElasticBlockStore.VolumeID)
	if err != nil {
		return "", fmt.Errorf("cannot parse persistent volume AWS Volume Id: %w", err)
	}

	volumeID := url.Path
	volumeID = strings.Trim(volumeID, "/")

	return volumeID, nil
}

func getAWSLoadBalancerName(svc *v1.Service) string {
	// Split hostname on . and after split the first part on -
	splitHostname := strings.Split(svc.Status.LoadBalancer.Ingress[0].Hostname, ".")
	splitSubDomain := strings.Split(splitHostname[0], "-")
	fromSplit := 0

	if strings.Contains(svc.Status.LoadBalancer.Ingress[0].Hostname, "internal") {
		// ex: internal-acc1b0155441645c6902a362c6821a9e-138903596.eu-west-1.elb.amazonaws.com
		fromSplit = 1
	}

	name := splitSubDomain[fromSplit]
	// Don't take the last one
	for i := fromSplit + 1; i < len(splitSubDomain)-1; i++ {
		name = name + "-" + splitSubDomain[i]
	}

	return name
}

func transformTagsToAwsEC2Tags(tagsList []*tags.Tag) []*ec2.Tag {
	awsEc2Tags := make([]*ec2.Tag, 0)

	for _, tag := range tagsList {
		awsEc2Tags = append(awsEc2Tags, &ec2.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
	}

	return awsEc2Tags
}

// GetActualTagsFromPersistentVolume Get actual tags from persistent volume.
func (apr *AWSProviderClient) GetActualTagsFromPersistentVolume(pv *v1.PersistentVolume) ([]*tags.Tag, error) {
	// Get volume ID from pv
	volumeID, err := getVolumeIDFromPersistentVolume(pv)
	if err != nil {
		return nil, err
	}
	// Prepare data for request
	volumesIDs := make([]*string, 0)
	volumesIDs = append(volumesIDs, aws.String(volumeID))
	// Describe volume to get all information from ec2 volume
	output, err := apr.ec2client.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: volumesIDs,
	})
	if err != nil {
		return nil, err
	}

	volumes := output.Volumes
	if len(volumes) != 1 {
		return nil, fmt.Errorf("can't find volume in AWS from volume id \"%s\"", volumeID)
	}

	volume := volumes[0]

	result := make([]*tags.Tag, 0)
	if volume.Tags == nil {
		return result, nil
	}

	// Transform aws tags in array
	for _, tag := range volume.Tags {
		result = append(result, &tags.Tag{Key: *tag.Key, Value: *tag.Value})
	}

	return result, nil
}

// GetActualTagsFromService Get actual tags from service.
func (apr *AWSProviderClient) GetActualTagsFromService(svc *v1.Service) ([]*tags.Tag, error) {
	// Check if it is a network loadbalancer (elbv2) or classic load balancer (elb)
	if svc.Annotations != nil && svc.Annotations[ServiceAnnotationLoadBalancerType] == KubernetesAnnotationsNLBValue {
		return apr.getActualTagsFromELBV2(svc)
	}

	return apr.getActualTagsFromELB(svc)
}

func (apr *AWSProviderClient) getELBV2ARNFromName(name string) (*string, error) {
	// Get data from AWS
	output, err := apr.elbv2client.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{
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

	return lb.LoadBalancerArn, nil
}

func (apr *AWSProviderClient) getActualTagsFromELBV2(svc *v1.Service) ([]*tags.Tag, error) {
	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(svc)
	// Get load balancer arn
	loadBalancerArn, err := apr.getELBV2ARNFromName(name)
	if err != nil {
		return nil, err
	}

	describeTagsOutput, err := apr.elbv2client.DescribeTags(&elbv2.DescribeTagsInput{
		ResourceArns: []*string{
			loadBalancerArn,
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

	result := make([]*tags.Tag, 0)

	if awsTags == nil {
		return result, nil
	}

	for _, awsTag := range awsTags {
		result = append(result, &tags.Tag{Key: *awsTag.Key, Value: *awsTag.Value})
	}

	return result, nil
}

func (apr *AWSProviderClient) getActualTagsFromELB(svc *v1.Service) ([]*tags.Tag, error) {
	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(svc)
	describeTagsOutput, err := apr.elbclient.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: []*string{
			aws.String(name),
		},
	})
	// Check error
	if err != nil {
		return nil, err
	}

	// Check response size
	if len(describeTagsOutput.TagDescriptions) == 0 {
		return nil, ErrNoTagsFound
	}

	awsTags := describeTagsOutput.TagDescriptions[0].Tags

	result := make([]*tags.Tag, 0)

	if awsTags == nil {
		return result, nil
	}

	for _, awsTag := range awsTags {
		result = append(result, &tags.Tag{Key: *awsTag.Key, Value: *awsTag.Value})
	}

	return result, nil
}

// AddTagsFromPersistentVolume Add Tags from persistent volume.
func (apr *AWSProviderClient) AddTagsFromPersistentVolume(pv *v1.PersistentVolume, tagsList []*tags.Tag) error {
	// Get volume ID from pv
	volumeID, err := getVolumeIDFromPersistentVolume(pv)
	if err != nil {
		return err
	}

	awsEc2Tags := transformTagsToAwsEC2Tags(tagsList)

	// Add tags to the created instance
	_, err = apr.ec2client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(volumeID)},
		Tags:      awsEc2Tags,
	})
	// Check error
	if err != nil {
		return err
	}

	return nil
}

// DeleteTagsFromPersistentVolume Delete tags from persistent volume.
func (apr *AWSProviderClient) DeleteTagsFromPersistentVolume(pv *v1.PersistentVolume, tagsList []*tags.Tag) error {
	// Get volume ID from pv
	volumeID, err := getVolumeIDFromPersistentVolume(pv)
	if err != nil {
		return err
	}

	awsEc2Tags := transformTagsToAwsEC2Tags(tagsList)

	// Add tags to the created instance
	_, err = apr.ec2client.DeleteTags(&ec2.DeleteTagsInput{
		Resources: []*string{aws.String(volumeID)},
		Tags:      awsEc2Tags,
	})
	// Check error
	if err != nil {
		return err
	}

	return nil
}

// AddTagsFromService Add tags from service.
func (apr *AWSProviderClient) AddTagsFromService(svc *v1.Service, tagsList []*tags.Tag) error {
	// Check if it is a network loadbalancer (elbv2) or classic load balancer (elb)
	if svc.Annotations != nil && svc.Annotations[ServiceAnnotationLoadBalancerType] == "nlb" {
		return apr.addTagsToELBV2(svc, tagsList)
	}

	return apr.addTagsToELB(svc, tagsList)
}

func (apr *AWSProviderClient) addTagsToELBV2(svc *v1.Service, tagsList []*tags.Tag) error {
	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(svc)

	// Get load balancer arn
	loadBalancerArn, err := apr.getELBV2ARNFromName(name)
	if err != nil {
		return err
	}

	awsAddTags := make([]*elbv2.Tag, 0)

	for _, tag := range tagsList {
		awsAddTags = append(awsAddTags, &elbv2.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
	}

	// Add tags to the created instance
	_, err = apr.elbv2client.AddTags(&elbv2.AddTagsInput{
		ResourceArns: []*string{
			loadBalancerArn,
		},
		Tags: awsAddTags,
	})

	return err
}

func (apr *AWSProviderClient) addTagsToELB(svc *v1.Service, tagsList []*tags.Tag) error {
	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(svc)

	awsAddTags := make([]*elb.Tag, 0)
	for _, tag := range tagsList {
		awsAddTags = append(awsAddTags, &elb.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)})
	}

	// Add tags to the created instance
	_, err := apr.elbclient.AddTags(&elb.AddTagsInput{
		LoadBalancerNames: []*string{
			aws.String(name),
		},
		Tags: awsAddTags,
	})

	return err
}

// DeleteTagsFromService Delete tags from service.
func (apr *AWSProviderClient) DeleteTagsFromService(svc *v1.Service, tagsList []*tags.Tag) error {
	// Check if it is a network loadbalancer (elbv2) or classic load balancer (elb)
	if svc.Annotations != nil && svc.Annotations[ServiceAnnotationLoadBalancerType] == "nlb" {
		return apr.deleteTagsToELBV2(svc, tagsList)
	}

	return apr.deleteTagsToELB(svc, tagsList)
}

func (apr *AWSProviderClient) deleteTagsToELBV2(svc *v1.Service, tagsList []*tags.Tag) error {
	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(svc)

	// Get load balancer arn
	loadBalancerArn, err := apr.getELBV2ARNFromName(name)
	if err != nil {
		return err
	}

	awsDeleteTags := make([]*string, 0)

	for _, tag := range tagsList {
		awsDeleteTags = append(awsDeleteTags, aws.String(tag.Key))
	}

	// Delete tags to the created instance
	_, err = apr.elbv2client.RemoveTags(&elbv2.RemoveTagsInput{
		ResourceArns: []*string{
			loadBalancerArn,
		},
		TagKeys: awsDeleteTags,
	})

	return err
}

func (apr *AWSProviderClient) deleteTagsToELB(svc *v1.Service, tagsList []*tags.Tag) error {
	// Get aws load balancer name from service
	name := getAWSLoadBalancerName(svc)

	awsDeleteTags := make([]*elb.TagKeyOnly, 0)
	for _, tag := range tagsList {
		awsDeleteTags = append(awsDeleteTags, &elb.TagKeyOnly{Key: aws.String(tag.Key)})
	}

	// Delete tags to the created instance
	_, err := apr.elbclient.RemoveTags(&elb.RemoveTagsInput{
		LoadBalancerNames: []*string{
			aws.String(name),
		},
		Tags: awsDeleteTags,
	})

	return err
}
