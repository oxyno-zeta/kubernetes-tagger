package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Kubernetes configuration home path
const kubeConfig = ".kube/config"

func configureViper(onChange func(e fsnotify.Event)) {
	kubeConfigPath := filepath.Join(os.Getenv("HOME"), kubeConfig)
	// Flags
	flag.String("namespace", "kube-system", "Namespace where "+projectName+" is deployed")
	flag.String("kubeconfig", kubeConfigPath, "Kubernetes configuration file path")
	flag.String("address", ":8085", "The address to expose health and prometheus metrics")
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
