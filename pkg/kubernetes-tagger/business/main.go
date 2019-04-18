package business

import (
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Watch Watch Kubernetes
func Watch(context *Context) {
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(context.KubernetesClient,
		0, kubeinformers.WithNamespace(""))

	persistentVolumeInformer := informerFactory.Core().V1().PersistentVolumes()
	serviceInformer := informerFactory.Core().V1().Services()

	persistentVolumeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    context.handlePersistentVolumeAdd,
		UpdateFunc: context.handlePersistentVolumeUpdate,
		DeleteFunc: context.handlePersistentVolumeDelete,
	})

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    context.handleServiceAdd,
		UpdateFunc: context.handleServiceUpdate,
		DeleteFunc: context.handleServiceDelete,
	})

	informerFactory.Start(wait.NeverStop)
}
