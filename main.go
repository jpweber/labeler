package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jpweber/labeler/node"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func Nodes(client *kubernetes.Clientset, excludes map[string]bool) {

	watchlist := cache.NewListWatchFromClient(client.Core().RESTClient(), "nodes", v1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Node{},
		time.Second*600,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Println("Running Add Function")
				K8sNode := obj.(*v1.Node)
				node := node.Node{
					Name:       K8sNode.ObjectMeta.Name,
					ExternalID: K8sNode.Spec.ExternalID,
					Excludes:   excludes,
				}

				go func(K8sNode *v1.Node) {
					// add the tags on to the node struct
					node.ListTags()
					origTagLen := len(K8sNode.GetLabels())
					// apply the tags to labels on the k8s node
					K8sNode = node.AddLabels(K8sNode)
					if origTagLen == len(K8sNode.GetLabels()) {
						log.Println("Not Updating", node.Name, "No changes to apply")
					} else {
						log.Println("Applying new labels to", node.Name)
						// update the actual node
						_, err := client.CoreV1().Nodes().Update(K8sNode)
						if err != nil {
							log.Println("Error updating node", err)
						}
					}
				}(K8sNode)

			},
			// UpdateFunc: func(old, new interface{}) {
			// 	log.Println("Running Update Function")
			// 	oldNode := new.(*v1.Node)
			// 	K8sNode := new.(*v1.Node)
			// 	if oldNode.ResourceVersion == K8sNode.ResourceVersion {
			// 		// Periodic resync will send update events for all known Deployments.
			// 		// Two different versions of the same Deployment will always have different RVs.
			// 		//debug
			// 		log.Println("Same resource version. Returning")
			// 		return
			// 	}

			// 	node := node.Node{
			// 		Name:       K8sNode.ObjectMeta.Name,
			// 		ExternalID: K8sNode.Spec.ExternalID,
			// 		Excludes:   excludes,
			// 	}

			// 	go func(K8sNode *v1.Node) {
			// 		// add the tags on to the node struct
			// 		node.ListTags()
			// 		origTagLen := len(K8sNode.GetLabels())
			// 		// apply the tags to labels on the k8s node
			// 		K8sNode = node.AddLabels(K8sNode)
			// 		if origTagLen == len(K8sNode.GetLabels()) {
			// 			log.Println("Not Updating", node.Name, "No changes to apply")
			// 		} else {
			// 			log.Println("Applying new labels to", node.Name)
			// 			// update the actual node
			// 			_, err := client.CoreV1().Nodes().Update(K8sNode)
			// 			if err != nil {
			// 				log.Println("Error updating node", err)
			// 			}
			// 		}
			// 	}(K8sNode)

			// },
		},
	)
	stop := make(chan struct{})
	done := make(chan bool)
	go controller.Run(stop)
	log.Println("Started Watching Nodes")
	<-done
}
func homeDir() string {
	return os.Getenv("HOME")
}

func main() {

	exludes := map[string]bool{
		"Name": true,
		"aws:autoscaling:groupName": true,
		"kubernetes.io/cluster/jpw": true,
	}
	var kubeconfig *string
	// var namespace *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// namespace = flag.String("namespace", "default", "name space to query")
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

	// Pods(clientset)
	Nodes(clientset, exludes)
}
