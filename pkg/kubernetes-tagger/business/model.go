package business

import (
	"github.com/Sirupsen/logrus"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/resources"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/rules"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Context Business context
type Context struct {
	KubernetesClient *kubernetes.Clientset
	Configuration    *config.Configuration
	Rules            []*rules.Rule
}

func (context *Context) handlePersistentVolumeAdd(obj interface{}) {
	pv := obj.(*v1.PersistentVolume)
	log := logrus.WithField("persistentVolumeName", pv.Name)
	log.Debug("New persistent volume added detected")
	err := context.runForPV(pv)
	if err != nil {
		log.Errorf("Error managing persistent volume: %v", err)
	}
}
func (context *Context) handlePersistentVolumeDelete(obj interface{}) {
	// Nothing to do
	pv := obj.(*v1.PersistentVolume)
	log := logrus.WithField("persistentVolumeName", pv.Name)
	log.Debug("New persistent volume deleted detected")
}

func (context *Context) handlePersistentVolumeUpdate(old, current interface{}) {
	currentPersistentVolume := current.(*v1.PersistentVolume)
	log := logrus.WithField("persistentVolumeName", currentPersistentVolume.Name)
	log.Debug("New persistent volume updated detected")
	err := context.runForPV(currentPersistentVolume)
	if err != nil {
		logrus.WithField("persistentVolumeName", currentPersistentVolume.Name).Errorf("Error managing persistent volume: %v", err)
	}
}

func (context *Context) runForPV(pv *v1.PersistentVolume) error {
	resource, err := resources.New(context.KubernetesClient, pv, context.Configuration)
	if err != nil {
		return err
	}
	if resource == nil {
		// No resource available
		return nil
	}
	// Check if resource can be processed
	if !resource.CanBeProcessed() {
		return nil
	}

	// Check if configuration is valid before continue
	err = resource.CheckIfConfigurationValid()
	if err != nil {
		return err
	}

	// Get actual tags
	actualTags, err := resource.GetActualTags()
	if err != nil {
		return err
	}

	availableTagValues, err := resource.GetAvailableTagValues()
	if err != nil {
		return err
	}

	delta, err := rules.CalculateTags(actualTags, availableTagValues, context.Rules)
	if err != nil {
		return err
	}
	err = resource.ManageTags(delta)
	if err != nil {
		return err
	}
	return nil
}
