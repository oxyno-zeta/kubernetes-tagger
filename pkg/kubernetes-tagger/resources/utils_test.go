package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestGetPersistentVolumeClaimWithoutClaimRef(t *testing.T) {
	spec := v1.PersistentVolumeSpec{ClaimRef: nil}
	pv := v1.PersistentVolume{Spec: spec}

	// Call code
	res, err := getPersistentVolumeClaim(&pv, nil)

	assert.Nil(t, res)
	assert.Nil(t, err)
}

func TestGetPersistentVolumeClaimWithClaimNotFound(t *testing.T) {
	claimRef := v1.ObjectReference{Namespace: "test-claim-ref-namespace", Name: "test-claim-ref-name"}
	spec := v1.PersistentVolumeSpec{ClaimRef: &claimRef}
	pv := v1.PersistentVolume{Spec: spec}

	client := testclient.NewSimpleClientset()

	// Call code
	res, err := getPersistentVolumeClaim(&pv, client)

	assert.Nil(t, res)
	assert.Nil(t, err)
}

func TestGetPersistentVolumeClaimWithClaimFound(t *testing.T) {
	claimRefNamespace := "test-claim-ref-namespace"
	claimRefName := "test-claim-ref-name"
	claimRef := v1.ObjectReference{Namespace: claimRefNamespace, Name: claimRefName}
	spec := v1.PersistentVolumeSpec{ClaimRef: &claimRef}
	pv := v1.PersistentVolume{Spec: spec}

	// PVC
	pvc := &v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: claimRefName, Namespace: claimRefNamespace}}

	client := testclient.NewSimpleClientset(pvc)

	// Call code
	res, err := getPersistentVolumeClaim(&pv, client)

	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, pvc, res)
}
