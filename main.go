package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"github.com/balajisa09/ekspose-controller/ekspose"
)


var config *rest.Config

func main(){

	
	var kubeconfig *string

	//Set kube config path
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	//Build config
	var err error
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		fmt.Println("Error while building kubeconfig ",err.Error())
		return 
	} 

	// Get clientset 
	clientset ,err := kubernetes.NewForConfig(config)

	if err != nil{
		fmt.Println("Error while creating a clientset from config : ",err.Error())
		return 
	}

	//Start the informer and the controller
	ch := make(chan struct{})
	informers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	c :=  ekspose.NewController(clientset,informers.Apps().V1().Deployments())
	informers.Start(ch)
	c.Run(ch)
}
