package resources

import (
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AWSVolumeResourceType AWS Volume Resource Type
const AWSVolumeResourceType = "volume"

// AWSResourcePlatform AWS Resource Platform
const AWSResourcePlatform = "aws"

func getPersistentVolumeClaim(persistentVolume *v1.PersistentVolume, k8sClient kubernetes.Interface) (*v1.PersistentVolumeClaim, error) {
	claimRef := persistentVolume.Spec.ClaimRef
	if claimRef == nil {
		return nil, nil
	}

	pvc, err := k8sClient.CoreV1().PersistentVolumeClaims(claimRef.Namespace).Get(claimRef.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}
	return pvc, nil
}
