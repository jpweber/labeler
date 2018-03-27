package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/jpweber/labeler/config"
	"github.com/jpweber/labeler/node"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	return os.Getenv("HOME")
}

func main() {

	configPath := "config.yaml"
	appConfig := config.ReadConfig(configPath)

	var kubeconfig *string
	// var namespace *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// ns := flag.String("namespace", "default", "name space to query")
	// not used yet
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Start node watcher
	n := node.Node{}
	n.Watcher(clientset, appConfig.Excludes)
}
