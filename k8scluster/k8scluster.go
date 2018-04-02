package k8scluster

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jpweber/labeler/configReader"
	"github.com/jpweber/labeler/provider"
	log "github.com/sirupsen/logrus"
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
func Watcher(client *kubernetes.Clientset, appConfig configReader.Config) {

	watchlist := cache.NewListWatchFromClient(client.Core().RESTClient(), "nodes", v1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Node{},
		time.Second*300,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Println("Observed new node being added to cluster")
				K8sNode := obj.(*v1.Node)

				go func(K8sNode *v1.Node) {
					n := Node{}
					n.Name = K8sNode.ObjectMeta.Name
					n.ExternalID = K8sNode.Spec.ExternalID
					n.Excludes = appConfig.Excludes
					// add the tags on to the node struct
					n.ProviderTags()

					// filter out any tags we do not want added as labels
					newLabels := n.filterExcludes()
					if len(newLabels) <= 0 {
						log.Printf("No new labels to apply on node %s", K8sNode.Name)
						return
					}

					// apply the tags to labels on the k8s node
					K8sNode = n.GenNewLabelSet(K8sNode, appConfig, newLabels)

					b := backoff.NewExponentialBackOff()
					b.MaxElapsedTime = 2 * time.Minute
					op := func() error {
						ret := ApplyLabels(client, K8sNode)
						return ret
					}
					err := backoff.Retry(op, b)
					if err != nil {
						log.Printf("error after retrying: %v", err)
					}
					// ApplyLabels(client, K8sNode)
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

func (n *Node) filterExcludes() map[string]string {
	newlabels := make(map[string]string)
	// filter out any explicitly ecluded labels
	for k, v := range n.Tags {
		if n.Excludes[k] != true {
			newlabels[k] = v
		}
	}

	return newlabels
}

func (n *Node) GenNewLabelSet(K8sNode *v1.Node, appConfig configReader.Config, newLabels map[string]string) *v1.Node {

	// fetch existing node labels
	labels := K8sNode.GetLabels()
	// add the new labels
	for k, v := range newLabels {
		labels[appConfig.Namespace+"/"+k] = v
	}

	K8sNode.SetLabels(labels)

	return K8sNode
}

func ApplyLabels(client *kubernetes.Clientset, K8sNode *v1.Node) error {
	log.Println("Applying new labels to", K8sNode.ObjectMeta.Name)
	// Always update the k8s node to the current revision of said node.
	currentNode, _ := client.CoreV1().Nodes().Get(K8sNode.Name, meta_v1.GetOptions{})

	// Get the labels from our original revision of the K8S node and apply them to the current
	// revision of the K8s node and update the node
	currentNode.SetLabels(K8sNode.GetLabels())
	_, err := client.CoreV1().Nodes().Update(currentNode)
	if err != nil {
		log.Println("Error updating node", err)
		return err
	}

	return nil
}
