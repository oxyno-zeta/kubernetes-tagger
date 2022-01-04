package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
)

// Kubernetes configuration home path
const kubeConfig = ".kube/config"

func configureViper(onChange func(e fsnotify.Event)) {
	kubeConfigPath := filepath.Join(os.Getenv("HOME"), kubeConfig)
	// Flags
	flag.String("namespace", "kube-system", "Namespace where "+projectName+" is deployed")
	flag.String("kubeconfig", kubeConfigPath, "Kubernetes configuration file path")
	flag.String("address", ":8085", "The address to expose health and prometheus metrics")
	flag.String("loglevel", "info", "Log level")
	flag.String("logformat", "json", "Log format")
	flag.String("provider", "aws", "Kubernetes Provider")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	// Add config file name
	viper.SetConfigName(config.RecommendedConfigFileName)
	// Add config possible path
	viper.AddConfigPath("/etc/" + projectName + "/")
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")
	// Watch configuration file change
	viper.WatchConfig()
	viper.OnConfigChange(onChange)
}

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

func configureLogger() {
	// Log level
	lvl, err := logrus.ParseLevel(context.Configuration.LogLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.SetLevel(lvl)

	// Log Formatter
	switch context.Configuration.LogFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	default:
		logrus.Fatalf("Log format not supported: %s", context.Configuration.LogFormat)
	}
}
