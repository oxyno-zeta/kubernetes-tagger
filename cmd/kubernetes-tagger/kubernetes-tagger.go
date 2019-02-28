package main

import (
	ctx "context"
	"os"
	"time"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/version"

	"github.com/fsnotify/fsnotify"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/business"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/rules"

	"path/filepath"

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

// Kubernetes configuration home path
const kubeConfig = ".kube/config"

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
	// Add default values
	viper.SetDefault("config.namespace", "kube-system")
	// Add config file name
	viper.SetConfigName(config.RecommendedConfigFileName)
	// Add config possible path
	viper.AddConfigPath("/etc/" + projectName + "/")
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")
	// Watch configuration file change
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
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
		context.MainConfiguration.Config.Namespace,
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
	var cfg config.MainConfiguration
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
	context.MainConfiguration = &cfg
}

// TODO Improve this to use configuration flags
// Take example on this: https://github.com/kubernetes/autoscaler/blob/8944afd9016dbda091066d2c6500af526caf6315/cluster-autoscaler/main.go
func getKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config

	kubeConfigPath := filepath.Join(os.Getenv("HOME"), kubeConfig)

	_, err := os.Stat(kubeConfigPath)

	if os.IsNotExist(err) {
		logrus.Info("Using in cluster config")
		config, err = rest.InClusterConfig()
	} else {
		logrus.WithFields(logrus.Fields{"config": kubeConfigPath}).Info("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}
	if err != nil {
		return nil, err
	}
	// TODO Add request timeout in the configuration (using flags)
	logrus.WithField("host", config.Host).Info("Create Kubernetes client")
	return kubernetes.NewForConfig(config)
}
