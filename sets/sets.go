package sets

import (
	"k8s.io/klog/v2"
)

type HostAction struct {
	Action string
	Host   string
	Result chan bool
}

type IngressHost struct {
	actionCh chan HostAction
	hosts    map[string]int
}

func NewIngressHost() *IngressHost {
	ingressHost := &IngressHost{
		actionCh: make(chan HostAction, 1024),
		hosts:    make(map[string]int, 1024),
	}
	go ingressHost.processActions()
	return ingressHost
}

func (ih *IngressHost) processActions() {
	for action := range ih.actionCh {
		switch action.Action {
		case "Add":
			ih.hosts[action.Host]++
			action.Result <- true
		case "Delete":
			if v, ok := ih.hosts[action.Host]; ok {
				if v > 1 {
					ih.hosts[action.Host]--
				} else {
					delete(ih.hosts, action.Host)
				}
			}
			action.Result <- true
		default:
			action.Result <- false
		}
	}
}

func (i *IngressHost) Add(host string) bool {
	resultCh := make(chan bool)
	i.actionCh <- HostAction{Action: "Add", Host: host, Result: resultCh}

	return <-resultCh
}

func (i *IngressHost) Delete(host string) bool {
	resultCh := make(chan bool)
	i.actionCh <- HostAction{Action: "Delete", Host: host, Result: resultCh}

	return <-resultCh
}

func (i *IngressHost) Has(host string) bool {
	for k := range i.hosts {
		if k == host {
			return true
		}
	}
	return false
}

func (i *IngressHost) Len() int {
	return len(i.hosts)
}

func (i *IngressHost) List() {
	if i.Len() > 0 {
		for k, v := range i.hosts {
			klog.V(2).Infof("Host=%s Count=%d", k, v)
		}
	}
}
