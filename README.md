# Labeler

Labeler runs in your Kubernetes cluster and applies labels to nodes based on that Tags are set on an instance in a cloud provider. Some cloud provider specific labels already are applied to a node when using the cloud provider subsystem. This continues that idea and allows user specific tags on instances to also get added as node labels. 

## Running Labeler
See the example config in the `deploy/` directory and modify values as needed or desired.   

Deploy to cluster:  
`kubectl apply -f deploy/`  

Remove from cluster:  
`kubectl delete -f deploy/`

View output logs:  
`kubectl logs deployment/labeler -n kube-system`

## Configuration
The configuration is a yaml formatted file that is saved as a config map and then read in at deploy time. Below is an exmaple configuration file

```

namespace: labeler.io
region: us-east-2    
provider: aws         
excludes:
    Name: true
    aws:autoscaling:groupName: true
    kubernetes.io/cluster/<clusterName> : true
```
`namespace` - Labels can be namespaced in kubernetes. This takes the form of `namespace/key=value`. If you wish to namespace your labels you can provide that string here.   

`region` - Specify the region of your cloud provider your nodes are running in.  

`provider` - the name of your cloud provider. (aws, gcp, azure)  

`excludes` - tags that may be applied to an instance that you wish to exclude from being applied as labels.  In the example some common starter values are provided. 



