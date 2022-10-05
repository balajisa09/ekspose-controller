package ekspose

import (
	"context"
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Controller struct{
	Clientset kubernetes.Interface
	DepLister applisters.DeploymentLister
	DepCacheSysnced cache.InformerSynced
	Queue workqueue.RateLimitingInterface 
}
func NewController( clientset kubernetes.Interface, depInformer appsinformer.DeploymentInformer) *Controller{

	c := &Controller{
		Clientset: clientset,
		DepLister: depInformer.Lister(),
		DepCacheSysnced: depInformer.Informer().HasSynced,
		Queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),"ekspose"),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.handleAdd,
			DeleteFunc: c.handleDelete,
		},
	)

	return c 	
}

func (c *Controller) Run(ch <- chan struct{}){

	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.DepCacheSysnced){
		fmt.Println("waiting for cache to be synced")}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *Controller) worker(){
	for c.processItem(){

	}
}

func (c *Controller) processItem() bool{
	item, shutdown := c.Queue.Get()
	if shutdown{
		return false
	}
	key,err := cache.MetaNamespaceKeyFunc(item)

	if(err != nil){
		fmt.Println("getting key from cache: ",err.Error())
	}
	ns,name,err := cache.SplitMetaNamespaceKey(key)

	if err !=nil {
		fmt.Println("splitting key in to namespace and name :",err.Error())
		return false
	}

	err = c.syncDeployment(ns,name)
	if err != nil{
		// re-try
		fmt.Println("Syncing deployments : ", err.Error())
		return false
	}

	return true
}


func (c *Controller) syncDeployment(ns, name string) error{

	//get deployment from informer

	dep,err := c.DepLister.Deployments(ns).Get(name)

	if err !=nil{
		fmt.Println("getting deployment from informer: ",err.Error())
	}

	//create service
	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: dep.Name,
			Namespace: ns,
		},
		Spec: v1.ServiceSpec{
			Selector: deplabels(*dep),
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	ctx := context.Background()
	_, err = c.Clientset.CoreV1().Services(ns).Create(ctx,&svc,metav1.CreateOptions{})
	if err != nil{
		fmt.Println("creating service : ",err.Error())
		return err
	}
	return nil

	//create ingress
}

func deplabels(d appsv1.Deployment) map[string]string{
	return d.Spec.Template.Labels
}

func (c *Controller) handleAdd(obj interface{}){
	fmt.Println("add was called")
	c.Queue.Add(obj)
}


func (c *Controller) handleDelete(obj interface{}){
	fmt.Println("delete was called")
	c.Queue.Add(obj)
}