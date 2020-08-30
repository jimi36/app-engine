package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"

	engine "github.com/jimi36/app-engine"
)

func BasePath(basePath string) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.basePath = basePath
		return nil
	}
}

func KubeNamespace(ns string) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.ns = ns
		return nil
	}
}

func InKubeCluster(inCluster bool) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.inKubeCluster = inCluster
		return nil
	}
}

func KubeConfPath(path string) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.kubeConfigPath = path
		return nil
	}
}

func Store(store engine.Store) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.store = store
		return nil
	}
}

func openKubeClient(cli engine.ClientImpl) error {
	c, _ := cli.(*Client)
	kubeConfig, err := func() (*rest.Config, error) {
		if c.inKubeCluster {
			return rest.InClusterConfig()
		}
		return clientcmd.BuildConfigFromFlags("", c.kubeConfigPath)
	}()
	if err != nil {
		return err
	}

	kubeCli, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	metricsCli, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	c.kubeCli = kubeCli
	c.metricsCli = metricsCli

	return nil
}

func applyOptions(cli *Client, opts []engine.Option) error {
	opts = append(opts, openKubeClient)
	for _, opt := range opts {
		if err := opt(cli); err != nil {
			return err
		}
	}
	return nil
}
