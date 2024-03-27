package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/tiggoins/ingress-validator/client"
	"github.com/tiggoins/ingress-validator/server"
	"github.com/tiggoins/ingress-validator/sets"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pflag.CommandLine.AddFlagSet(client.NewClientFlagSet())
	pflag.CommandLine.AddFlagSet(server.NewServerFlagSet())
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.BoolP("help", "h", false, "Usage of ingress-validator")
	pflag.Parse()
	if pflag.CommandLine.Changed("help") || !pflag.CommandLine.Changed("tls-cert-file") || !pflag.CommandLine.Changed("tls-key-file") {
		pflag.Usage()
		os.Exit(0)
	}

	gHosts := sets.NewIngressHost()

	server := server.NewServer(ctx, gHosts)
	// 1. start server to serve ValidatingWebhook
	go server.Start()

	client := client.NewK8sClient(ctx, gHosts)
	// 2. list ingress from cluster
	client.ListIngress()
	// 3. start informer to watch the event of delete ingress
	go client.StartInformer()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	// stop the server gracefully
	if err := server.Shutdown(); err != nil {
		klog.Error(err)
	}
	os.Exit(0)
}
