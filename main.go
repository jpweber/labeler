package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/jpweber/labeler/configReader"
	"github.com/jpweber/labeler/k8scluster"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func homeDir() string {
	return os.Getenv("HOME")
}

func main() {

	configPath := flag.String("config", "/etc/labeler/config.yaml", "Path to config file")
	// not used yet
	flag.Parse()
	// configPath := "/etc/labeler/config.yaml"
	appConfig := configReader.Read(*configPath)

	// var kubeconfig *string

	// if home := homeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }

	// use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// creates the in-cluster config
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		panic(err.Error())
	}

	// Start node watcher
	k8scluster.Watcher(clientset, appConfig)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutdown signal received, exiting...")
}
