package business

import (
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/resources"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/rules"
	"github.com/sirupsen/logrus"
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

func (context *Context) handleServiceAdd(obj interface{}) {
	svc := obj.(*v1.Service)
	log := logrus.WithFields(logrus.Fields{
		"serviceName": svc.Name,
		"namespace":   svc.Namespace,
	})
	log.Debug("New service added detected")
	err := context.runForService(svc)
	if err != nil {
		log.Errorf("Error managing service: %v", err)
	}
}
func (context *Context) handleServiceDelete(obj interface{}) {
	// Nothing to do
	svc := obj.(*v1.Service)
	log := logrus.WithFields(logrus.Fields{
		"serviceName": svc.Name,
		"namespace":   svc.Namespace,
	})
	log.Debug("New service deleted detected")
}

func (context *Context) handleServiceUpdate(old, current interface{}) {
	currentService := current.(*v1.Service)
	log := logrus.WithFields(logrus.Fields{
		"serviceName": currentService.Name,
		"namespace":   currentService.Namespace,
	})
	log.Debug("New service updated detected")
	err := context.runForService(currentService)
	if err != nil {
		logrus.WithField("serviceName", currentService.Name).Errorf("Error managing service: %v", err)
	}
}

func (context *Context) runForPV(pv *v1.PersistentVolume) error {
	resource, err := resources.NewFromPersistentVolume(context.KubernetesClient, pv, context.Configuration)
	if err != nil {
		return err
	}
	return context.runForResource(resource)
}

func (context *Context) runForService(svc *v1.Service) error {
	resource, err := resources.NewFromService(context.KubernetesClient, svc, context.Configuration)
	if err != nil {
		return err
	}
	return context.runForResource(resource)
}

func (context *Context) runForResource(resource resources.Resource) error {
	if resource == nil {
		// No resource available
		return nil
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
