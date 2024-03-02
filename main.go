package main

import (
	"log"

	"client/cmd"
	"client/global"
	"client/internal"
	"client/pkg"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	var err error
	err = cmd.Execute()
	if err != nil {
		log.Fatalf("cmd.Execute err: %v", err)
	}
	err = pkg.ReadYaml(global.ConfigPath, &global.MyConfig)
	if err != nil {
		log.Fatalf("pkg.ReadYaml err: %v", err)
	}
}

func main() {
	// 获取 kubeconfig 配置，优先从文件中读取，若失败则尝试使用集群内部配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)

	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalln("can't get kubeconfig")
		}
		config = inClusterConfig
	}

	// 根据配置创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("can't create client")
	}

	// 创建 SharedInformer 工厂，用于创建 Service 和 Ingress 的 Informer
	factory := informers.NewSharedInformerFactory(clientset, 0)
	serviceInformer := factory.Core().V1().Services()
	ingressInformer := factory.Networking().V1().Ingresses()

	// 创建控制器
	controller := internal.NewController(clientset, serviceInformer, ingressInformer)
	stopCh := make(chan struct{})

	// 启动 Informer 工厂
	factory.Start(stopCh)
	// 等待 Informer 缓存同步完成
	factory.WaitForCacheSync(stopCh)
	// 运行控制器
	controller.Run(stopCh)
}
