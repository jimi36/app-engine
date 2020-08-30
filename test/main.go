package main

import (
	"time"

	coreV1 "k8s.io/api/core/v1"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/kube"
	"github.com/jimi36/app-engine/log"
	"github.com/jimi36/app-engine/native"
)

func main() {
	//runKubeTest()
	runNativeTest()
}

func runKubeTest() {
	var opts []engine.Option
	opts = append(opts, kube.InKubeCluster(false))
	opts = append(opts, native.BasePath("/home/zht/kube"))
	opts = append(opts, kube.KubeConfPath("/home/zht/.kube/config"))

	impl, err := kube.NewClient(opts...)
	if err != nil {
		log.Errorf("new kube client error: %s", err.Error())
		return
	}

	cli := engine.NewClient(impl)
	if err := cli.Start(); err != nil {
		return
	}

	config := engine.Config{
		Name: "my-config",
		Data: map[string]string{
			"test.txt": "hello world",
		},
	}
	if err := cli.CreateConfig(&config); err != nil {
		log.Errorf("kube client new config error: %s", err.Error())
		//return
	}

	app := engine.Application{
		ApplicationTag: engine.ApplicationTag{
			Name:    "test",
			Version: "1.0.0",
		},
		Type: engine.Kube,
		KubeSpec: &engine.KubeAppSpec{
			Image: "nginx:latest",
			Ports: []engine.KubePort{
				{
					Name:          "p80",
					ContainerPort: 80,
				},
			},
			Service: &engine.KubeService{
				Type: coreV1.ServiceTypeNodePort,
				Ports: []engine.KubeServicePort{
					engine.KubeServicePort{
						Port:       8080,
						TargetPort: 80,
						NodePort:   30001,
					},
				},
			},
		},
	}
	cli.CreateApplication(&app)

	if err := cli.StartApplication(&engine.ApplicationTag{"test", "1.0.0"}); err != nil {
		log.Errorf("kube client new application  error: %s", err.Error())
		//return
	}

	time.Sleep(time.Second * 5)

	state, err := cli.GetApplicationStates([]*engine.ApplicationTag{&engine.ApplicationTag{"test", "1.0.0"}})
	if err != nil {
		log.Errorf("kube client get application state  error: %s", err.Error())
		return
	}

	log.Infof("%v", state[0])

	ch := make(chan struct{})
	<-ch
}

func runNativeTest() {
	var opts []engine.Option
	opts = append(opts, native.BasePath("/home/zht/native"))
	impl, err := native.NewClient(opts...)
	if err != nil {
		log.Errorf("new native client new error: %s", err.Error())
	}

	cli := engine.NewClient(impl)
	if err := cli.Start(); err != nil {
		return
	}
	/*
		config := engine.Config{
			Name: "test",
			Data: map[string]string{
				"test.txt": "hello world",
			},
		}
		if err := cli.CreateConfig(&config); err != nil {
			log.Errorf("native client create config error: %s", err.Error())
			return
		}
	*/
	app := engine.Application{
		ApplicationTag: engine.ApplicationTag{
			Name:    "test",
			Version: "1.0.0",
		},
		Type: engine.Native,
		NativeSpec: &engine.NativeAppSpec{
			//Command: []string{"cat", "config/test/test.txt"},
			Command: []string{"test"},
		},
	}
	cli.CreateApplication(&app)

	if err := cli.StartApplication(&engine.ApplicationTag{"test", "1.0.0"}); err != nil {
		log.Errorf("native client start application error: %s", err.Error())
		//return
	}

	time.Sleep(time.Second * 2)

	state, err := cli.GetApplicationStates([]*engine.ApplicationTag{&engine.ApplicationTag{"test", "1.0.0"}})
	if err != nil {
		log.Errorf("native client get application state error: %s", err.Error())
		return
	}
	if len(state) > 0 {
		log.Infoln(state[0])
	}

	if err := cli.StopApplication(&engine.ApplicationTag{"test", "1.0.0"}); err != nil {
		log.Errorf("native client delete application error: %s", err.Error())
		return
	}

	ch := make(chan struct{})
	<-ch
}
