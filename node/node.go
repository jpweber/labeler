package node

import (
	"github.com/jpweber/labeler/provider"
	"k8s.io/api/core/v1"
)

type Node struct {
	Name       string
	ExternalID string
	Tags       map[string]string
	Excludes   map[string]bool
}

func (n *Node) ListTags() {
	n.Tags = provider.EC2Tags(n.ExternalID)
}

func (n *Node) AddLabels(K8sNode *v1.Node) *v1.Node {
	newlabels := make(map[string]string)

	// filter out any explicitly ecluded labels
	for k, v := range n.Tags {
		if n.Excludes[k] != true {
			newlabels[k] = v
		}
	}
	labels := K8sNode.GetLabels()
	// add the new labels
	for k, v := range newlabels {
		labels[k] = v
	}
	// fmt.Println("old")
	// fmt.Printf("%+v", K8sNode.GetLabels())
	K8sNode.SetLabels(labels)
	// fmt.Println("here be labels")
	// fmt.Println("new")
	// fmt.Printf("%+v", K8sNode.GetLabels())

	return K8sNode
}
