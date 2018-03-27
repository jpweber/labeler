package node

import (
	"fmt"
	"log"
	"time"

	"github.com/jpweber/labeler/provider"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Node struct {
	Name       string
	ExternalID string
	Tags       map[string]string
	Excludes   map[string]bool
}

// Watcher - starts the watcher for nodes joining the cluster
// and triggers the adding of labels on new node connection
func (n *Node) Watcher(client *kubernetes.Clientset, excludes map[string]bool) {

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

				n.Name = K8sNode.ObjectMeta.Name
				n.ExternalID = K8sNode.Spec.ExternalID
				n.Excludes = excludes

				go func(K8sNode *v1.Node) {
					// add the tags on to the node struct
					n.ProviderTags()

					// apply the tags to labels on the k8s node
					K8sNode = n.GenNewLabelSet(K8sNode)
					// log.Println("Applying new labels to", n.Name)
					// // update the actual node
					// _, err := client.CoreV1().Nodes().Update(K8sNode)
					// if err != nil {
					// 	log.Println("Error updating node", err)
					// 	updatedNode, _ := client.CoreV1().Nodes().Get(K8sNode.Name, meta_v1.GetOptions{})
					// 	log.Println(updatedNode)
					// }
					ApplyLabels(client, K8sNode)
				}(K8sNode)

			},
		},
	)
	stop := make(chan struct{})
	done := make(chan bool)
	go controller.Run(stop)
	log.Println("Started Watching Nodes")
	<-done
}

func (n *Node) ProviderTags() {
	n.Tags = provider.EC2Tags(n.ExternalID)
	// as we add different cloud providrers
	// create a swtich statement to fetch from different sources
}

func (n *Node) GenNewLabelSet(K8sNode *v1.Node) *v1.Node {
	newlabels := make(map[string]string)

	// filter out any explicitly ecluded labels
	for k, v := range n.Tags {
		if n.Excludes[k] != true {
			newlabels[k] = v
		}
	}

	// fetch existing node labels
	labels := K8sNode.GetLabels()
	objMeta := K8sNode.GetResourceVersion()
	fmt.Println(objMeta)
	// add the new labels
	for k, v := range newlabels {
		labels[k] = v
	}

	K8sNode.SetLabels(labels)

	return K8sNode
}

func ApplyLabels(client *kubernetes.Clientset, K8sNode *v1.Node) {
	log.Println("Applying new labels to", K8sNode.ObjectMeta.Name)
	// update the actual node
	_, err := client.CoreV1().Nodes().Update(K8sNode)
	if err != nil {
		log.Println("Error updating node", err)
		updatedNode, _ := client.CoreV1().Nodes().Get(K8sNode.Name, meta_v1.GetOptions{})
		ApplyLabels(client, updatedNode)
	}
}
