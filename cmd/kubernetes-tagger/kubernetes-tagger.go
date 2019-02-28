package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/business"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/rules"

	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Project name used for configuration path
const projectName = "kubernetes-tagger"

// Kubernetes configuration home path
const kubeConfig = ".kube/config"

var context = &business.BusinessContext{}

func main() {
	// Viper configuration
	viper.SetConfigName(config.RecommendedConfigFileName)
	viper.AddConfigPath("/etc/" + projectName + "/")
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// Event only say that the file is reloading
		fmt.Println("Config file changed:", e.Name)
		// Reload configuration
		readConfiguration()
	})

	readConfiguration()

	kubeClient, err := getKubernetesClient()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	context.KubernetesClient = kubeClient
	business.WatchPersistentVolumes(context)
}

func readConfiguration() {
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Error(fmt.Errorf("Fatal error config file: %s \n", err))
		os.Exit(1)
	}
	var cfg config.MainConfiguration
	err = viper.Unmarshal(&cfg)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Generate rules from rules declared in configuration
	rules, err := rules.New(cfg.Rules)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Update context
	context.Rules = rules
	context.MainConfiguration = &cfg
}

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
	logrus.Info("Created Kubernetes client %s", config.Host)
	return kubernetes.NewForConfig(config)
}
