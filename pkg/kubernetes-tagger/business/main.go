package business

import (
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// WatchPersistentVolumes Watch persistent volumes
func WatchPersistentVolumes(context *Context) {

	// Watch for persistent volume
	watchList := cache.NewListWatchFromClient(
		context.KubernetesClient.CoreV1().RESTClient(),
		"persistentvolumes",
		"",
		fields.Everything())

	// Create informer
	_, controller := cache.NewInformer(
		watchList,
		&v1.PersistentVolume{},
		time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    context.handlePersistentVolumeAdd,
			UpdateFunc: context.handlePersistentVolumeUpdate,
			DeleteFunc: context.handlePersistentVolumeDelete,
		},
	)

	stop := make(chan struct{})
	// Launch in a sub routine
	go controller.Run(stop)
	<-stop
}
