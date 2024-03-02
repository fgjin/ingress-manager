package internal

import (
	"context"
	"log"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"

	"client/global"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	informer "k8s.io/client-go/informers/core/v1"
	netInformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	coreLister "k8s.io/client-go/listers/core/v1"
	netLister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	client        kubernetes.Interface
	ingressLister netLister.IngressLister
	serviceLister coreLister.ServiceLister
	queue         workqueue.RateLimitingInterface
}

// 更新 Service 对象时调用，比较旧对象和新对象，若有变化则加入工作队列
func (c *controller) updateService(oldObj interface{}, newObj interface{}) {
	//TODO 比较注解
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	c.enqueue(newObj)
}

// 添加 Service 对象时调用，将新增的对象加入工作队列
func (c *controller) addService(obj interface{}) {
	c.enqueue(obj)
}

// 将对象加入工作队列
func (c *controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.queue.Add(key)
}

// 删除 Ingress 对象时调用，如果对象是 Service 的所有者，则将对应的 Ingress 对象加入工作队列
func (c *controller) deleteIngress(obj interface{}) {
	ingress := obj.(*netv1.Ingress)
	ownerReference := metav1.GetControllerOf(ingress)
	if ownerReference == nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}
	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}

// 运行控制器的主循环，启动工作协程
func (c *controller) Run(stopCh chan struct{}) {
	for i := 0; i < global.WorkNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

// 工作协程函数，处理工作队列中的对象
func (c *controller) worker() {
	for c.processNextItem() {
	}
}

// 处理下一个工作队列中的对象
func (c *controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	key := item.(string)

	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

// 同步 Service 对象的状态
func (c *controller) syncService(key string) error {
	// 解析 key，获取 namespace 和 name
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// 从缓存中获取指定命名空间中指定名称的 Service 对象
	service, err := c.serviceLister.Services(namespaceKey).Get(name)

	if errors.IsNotFound(err) {
		log.Println(err)
		return nil
	}
	if err != nil {
		return err
	}

	// 检查 Service 对象的注解，根据注解的有无执行新增或删除操作
	_, ok := service.GetAnnotations()[global.CustomAnnotation]
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if ok && errors.IsNotFound(err) {
		//create ingress
		ig := c.constructIngress(service)
		_, err := c.client.NetworkingV1().Ingresses(namespaceKey).Create(context.TODO(), ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		log.Printf("ingress创建成功 ingress:%v namespace:%v", service.Name, service.Namespace)

	} else if !ok && ingress != nil {
		//delete ingress
		err := c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		log.Printf("ingress删除成功 ingress:%v namespace:%v", service.Name, service.Namespace)
	}
	return nil
}

// 处理错误情况
func (c *controller) handlerError(key string, err error) {
	if c.queue.NumRequeues(key) <= global.MaxRetry {
		c.queue.AddRateLimited(key)
		return
	}
	runtime.HandleError(err)
	c.queue.Forget(key)
}

// 构造 Ingress 对象
func (c *controller) constructIngress(service *corev1.Service) *netv1.Ingress {
	ingress := netv1.Ingress{}
	// Service 与 Ingress 通过OwnerReference 绑定
	ingress.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(service, corev1.SchemeGroupVersion.WithKind("Service")),
	}
	ingress.Name = service.Name
	ingress.Namespace = service.Namespace

	pathType := netv1.PathTypePrefix
	host := service.Name + global.MyConfig.GetHost()
	ingressClassName := global.MyConfig.GetIngressClassName()

	ingress.Spec = netv1.IngressSpec{
		IngressClassName: &ingressClassName,
		Rules: []netv1.IngressRule{
			{
				Host: host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{
								Path:     global.MyConfig.GetPath(),
								PathType: &pathType,
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: service.Name,
										Port: netv1.ServiceBackendPort{
											Number: global.MyConfig.GetNumber(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &ingress
}

// 创建一个新的控制器实例
func NewController(client kubernetes.Interface, serviceInformer informer.ServiceInformer, ingressInformer netInformer.IngressInformer) controller {
	// 初始化控制器
	c := controller{
		client:        client,
		ingressLister: ingressInformer.Lister(),
		serviceLister: serviceInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}
	// 添加 Service 和 Ingress 的事件处理函数
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
	})
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})
	return c
}
