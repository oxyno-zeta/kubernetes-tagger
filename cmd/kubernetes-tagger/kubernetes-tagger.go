package main

import (
	ctx "context"
	"os"
	"time"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/utils"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/version"

	"github.com/fsnotify/fsnotify"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/business"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/rules"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"

	kube_record "k8s.io/client-go/tools/record"
)

// Project name used for configuration path
const projectName = "kubernetes-tagger"

var context = &business.Context{}

// Default values for leader election
const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

func defaultLeaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:   true,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.EndpointsResourceLock,
	}
}

func main() {
	// Get Hostname to have unique id for container
	id, err := os.Hostname()
	if err != nil {
		logrus.Fatal(err)
	}

	// Viper configuration
	configureViper(func(e fsnotify.Event) {
		// Event only say that the file is reloading
		logrus.WithField("file", e.Name).Info("Configuration file changed")
		// Reload configuration
		readConfiguration()
	})

	readConfiguration()

	versionObj := version.GetVersion()
	logrus.WithFields(logrus.Fields{
		"build-date": versionObj.BuildDate,
		"git-commit": versionObj.GitCommit,
		"version":    versionObj.Version,
	}).Infof("Starting %s", projectName)

	kubeClient, err := getKubernetesClient()
	if err != nil {
		logrus.Fatalf("Cannot create a Kubernetes client: %v", err)
	}

	// Add Kubernetes client to context
	context.KubernetesClient = kubeClient

	// Create default leader election configuration
	leaderElection := defaultLeaderElectionConfiguration()

	// Create event broadcaster
	eventBroadcaster := kube_record.NewBroadcaster()
	// Add logger to event broadcaster
	eventBroadcaster.StartLogging(logrus.Infof)
	// Create event recorder from event broadcaster
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: projectName})

	// Create new resource lock
	lock, err := resourcelock.New(
		leaderElection.ResourceLock,
		context.Configuration.Namespace,
		projectName,
		kubeClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: eventRecorder,
		},
	)
	if err != nil {
		logrus.Fatalf("Unable to create leader election lock: %v", err)
	}

	// Leader election
	leaderelection.RunOrDie(ctx.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: leaderElection.LeaseDuration.Duration,
		RenewDeadline: leaderElection.RenewDeadline.Duration,
		RetryPeriod:   leaderElection.RetryPeriod.Duration,
		Callbacks: leaderelection.LeaderCallbacks{
			OnNewLeader: func(identity string) {
				if identity != id {
					logrus.WithField("leader", identity).Info("Other leader detected")
				}
			},
			OnStartedLeading: func(_ ctx.Context) {
				// Since we are committing a suicide after losing
				// mastership, we can safely ignore the argument.
				logrus.WithField("leader", id).Info("Start leading")
				run()
			},
			OnStoppedLeading: func() {
				logrus.Fatal("Lost master")
			},
		},
	})
}

func run() {
	logrus.Info("Launch business")
	business.WatchPersistentVolumes(context)
}

func readConfiguration() {
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("Fatal error reading configuration file: %v", err)
	}
	var cfg config.Configuration
	err = viper.Unmarshal(&cfg)
	if err != nil {
		logrus.Fatalf("Error marshalling configuration: %v", err)
	}

	// Generate rules from rules declared in configuration
	rules, err := rules.New(cfg.Rules)
	if err != nil {
		logrus.Fatal(err)
	}

	// Update context
	context.Rules = rules
	context.Configuration = &cfg
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	kubeConfigPath := context.Configuration.Kubeconfig

	exists, err := utils.Exists(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	if exists {
		logrus.WithFields(logrus.Fields{"config": kubeConfigPath}).Info("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	} else {
		logrus.Info("Using in cluster config")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}
	logrus.WithField("host", config.Host).Info("Create Kubernetes client")
	return kubernetes.NewForConfig(config)
}
