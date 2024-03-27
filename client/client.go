package client

import (
	"context"
	"time"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/tiggoins/ingress-validator/sets"
)

const EmptyHost = "EMPTY_HOST"

type Client struct {
	master     string
	kubeconfig string
	ctx        context.Context
	clientset  *kubernetes.Clientset
	ghosts     *sets.IngressHost
}

func NewClientFlagSet() *pflag.FlagSet {
	clientFlags := pflag.NewFlagSet("client", pflag.ExitOnError)
	clientFlags.String("master", "", "Master ip to use for CLI requests")
	clientFlags.String("kubeconfig", "", "Path to the certificate file")

	return clientFlags
}

func NewK8sClient(ctx context.Context, ghosts *sets.IngressHost) *Client {
	c := new(Client)
	// pflag.CommandLine.StringVar(&c.master, "master", "", "Master ip to use for CLI requests")
	c.master, _ = pflag.CommandLine.GetString("master")
	// pflag.CommandLine.StringVar(&c.kubeconfig, "kubeconfig", "", "Path to the certificate file")
	c.kubeconfig, _ = pflag.CommandLine.GetString("kubeconfig")

	if c.kubeconfig == "" {
		c.kubeconfig = clientcmd.RecommendedHomeFile
	}
	config, err := clientcmd.BuildConfigFromFlags(c.master, c.kubeconfig)
	if err != nil {
		klog.V(2).Infof("Cannot build rest.Config,will use rest.InClusterConfig()")
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Still cannot creating in-cluster config: %v,exit the program\n", err)
		}
	}

	c.ctx = ctx
	c.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating Kubernetes client: %v\n", err)
	}

	c.ghosts = ghosts
	return c
}

func (c *Client) ListIngress() {
	networkIngress, err := c.clientset.NetworkingV1().Ingresses(corev1.NamespaceAll).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Fatalln(err)
	}
	klog.V(2).Infof("Got %d items networking/v1.Ingress in the cluster", len(networkIngress.Items))

	var hosts []string
	for _, ingress := range networkIngress.Items {
		for _, host := range ingress.Spec.Rules {
			if host.Host == "" {
				host.Host = EmptyHost
			}
			hosts = append(hosts, host.Host) // 将主机信息添加到切片中
		}
	}
	// 将切片中的所有主机信息添加到 IngressHost 对象中
	for _, host := range hosts {
		if status := c.ghosts.Add(host); !status {
			klog.Error("Error in add host to gHosts")
		}
	}
	// list the domain->count
	c.ghosts.List()
}

func (c *Client) StartInformer() {
	ingressListWatcher := cache.NewListWatchFromClient(c.clientset.NetworkingV1().RESTClient(), "ingresses", metav1.NamespaceAll, fields.Everything())
	_, informer := cache.NewIndexerInformer(ingressListWatcher, &networkingv1.Ingress{}, time.Second*5, cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.DeleteIngressHost,
	}, cache.Indexers{})
	
	klog.Info("Starting informer to watch the event of deletion of Ingress")
	go informer.Run(c.ctx.Done())

	if !cache.WaitForCacheSync(c.ctx.Done(), informer.HasSynced) {
		klog.Error("Timed out waiting for caches to sync")
		return
	}
}

func (c *Client) DeleteIngressHost(obj interface{}) {
	ingress, ok := obj.(*networkingv1.Ingress)
	if !ok {
		klog.Errorf("Failed to cast object to *networking.Ingress")
		return
	}
	klog.V(2).Infof("Ingress %s/%s has been deleted", ingress.Namespace, ingress.Name)

	var hosts []string
	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			hosts = append(hosts, EmptyHost)
		} else {
			hosts = append(hosts, rule.Host)
		}
	}

	for _, host := range hosts {
		klog.V(2).Infof("removed hosts %v from GlobalIngressHosts", host)
		if status := c.ghosts.Delete(host); !status {
			klog.Error("Error in delete host from gHosts")
		}
	}
	klog.V(2).Infof("Now GlobalIngressHosts have %d items", c.ghosts.Len())
}