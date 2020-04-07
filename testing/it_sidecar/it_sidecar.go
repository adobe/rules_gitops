package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/adobe/rules_gitops/testing/it_sidecar/stern"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/homedir"
)

type portForwardConf struct {
	services map[string][]uint16
}

func (i *portForwardConf) String() string {
	return fmt.Sprintf("%v", i.services)
}

func (i *portForwardConf) Set(value string) error {
	v := strings.SplitN(value, ":", 2)
	if len(v) != 2 {
		return fmt.Errorf("incorrect portforward '%s': must be in form of service:port", value)
	}
	port, err := strconv.ParseUint(v[1], 10, 16)
	if err != nil {
		return fmt.Errorf("incorrect port in portforward '%s': %v", value, err)
	}
	i.services[v[0]] = append(i.services[v[0]], uint16(port))
	return nil
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	namespace       = flag.String("namespace", os.Getenv("NAMESPACE"), "kubernetes namespace")
	timeout         = flag.Duration("timeout", time.Second*30, "execution timeout")
	deleteNamespace = flag.Bool("delete_namespace", false, "delete namespace as part of the cleanup")
	pfconfig        = portForwardConf{services: make(map[string][]uint16)}
	signalChannel   chan os.Signal
	kubeconfig      string
	waitForApps     arrayFlags
)

func init() {
	flag.Var(&pfconfig, "portforward", "set a port forward item in form of servicename:port")
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "path to kubernetes config file")
	flag.Var(&waitForApps, "waitforapp", "wait for pods with label app=<this parameter>")
}

// contains returns true if slice v contains an item
func contains(v []string, item string) bool {
	for _, s := range v {
		if s == item {
			return true
		}
	}
	return false
}

// listReadyApps converts a list returned from podsInformer.GetStore().List() to a map containing apps with ready status
// app is determined by app label
func listReadyApps(list []interface{}) (readypods, notReady []string) {
	var readyApps []string
	for _, it := range list {
		pod, ok := it.(*v1.Pod)
		if !ok {
			panic(errors.New("expected pod in informer"))
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v1.PodReady {
				if cond.Status == v1.ConditionTrue {
					readypods = append(readypods, pod.Name)
					app := pod.GetLabels()["app"]
					if app != "" {
						readyApps = append(readyApps, app)
					}
					app = pod.GetLabels()["app.kubernetes.io/name"]
					if app != "" {
						readyApps = append(readyApps, app)
					}

				}
			}
		}
	}
	for _, app := range waitForApps {
		if !contains(readyApps, app) {
			notReady = append(notReady, app)
		}
	}
	return
}

func waitForPods(ctx context.Context, clientset *kubernetes.Clientset) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	events := make(chan interface{})
	fn := func(obj interface{}) {
		events <- obj
	}

	handler := &cache.ResourceEventHandlerFuncs{
		AddFunc:    fn,
		DeleteFunc: fn,
		UpdateFunc: func(old interface{}, new interface{}) {
			fn(new)
		},
	}

	kubeInformerFactory := informers.NewFilteredSharedInformerFactory(clientset, time.Second*30, *namespace, nil)
	podsInformer := kubeInformerFactory.Core().V1().Pods().Informer()
	podsInformer.AddEventHandler(handler)
	go kubeInformerFactory.Start(ctx.Done())

waitForPodsUp:
	for {
		select {
		case <-events:
			v := podsInformer.GetStore().List()
			ready, notReady := listReadyApps(v)
			log.Print("ready pods:", ready)
			if len(notReady) != 0 {
				log.Print("waiting for apps:", notReady)
			} else {
				log.Println("all apps are ready")
				break waitForPodsUp
			}
		case <-ctx.Done():
			return errors.New("timed out waiting for apps")
		}
	}
	return nil
}

// listReadyServices converts a list returned from endpointsInformer.GetStore().List() to a list of services with ready status
func listReadyServices(list []interface{}) (ready, notReady []string) {
	for _, it := range list {
		ep, ok := it.(*v1.Endpoints)
		if !ok {
			panic(errors.New("expected EndpointsList in informer"))
		}
		for _, subset := range ep.Subsets {
			if len(subset.Addresses) > 0 {
				ready = append(ready, ep.Name)
				break
			}
		}
	}
	for service, _ := range pfconfig.services {
		if !contains(ready, service) {
			notReady = append(notReady, service)
		}
	}
	return
}

func waitForEndpoints(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config) error {
	events := make(chan interface{})
	fn := func(obj interface{}) {
		events <- obj
	}

	handler := &cache.ResourceEventHandlerFuncs{
		AddFunc:    fn,
		DeleteFunc: fn,
		UpdateFunc: func(old interface{}, new interface{}) {
			fn(new)
		},
	}

	kubeInformerFactory := informers.NewFilteredSharedInformerFactory(clientset, time.Second*30, *namespace, nil)
	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints().Informer()
	endpointsInformer.AddEventHandler(handler)
	go kubeInformerFactory.Start(ctx.Done())

	allReadyServices := make(map[string]bool)
waitForServicesUp:
	for {
		select {
		case <-events:
			v := endpointsInformer.GetStore().List()
			ready, notReady := listReadyServices(v)
			log.Print("ready services:", ready)
			for _, svc := range ready {
				if !allReadyServices[svc] {
					allReadyServices[svc] = true
					log.Print("SERVICE_READY ", svc)
					if ports := pfconfig.services[svc]; len(ports) > 0 {
						err := portForward(ctx, clientset, config, svc, ports)
						if err != nil {
							return err
						}
					}
				}
			}
			if len(notReady) != 0 {
				log.Print("waiting for endpoints:", notReady)
			} else {
				log.Println("all services are ready")
				break waitForServicesUp
			}
		case <-ctx.Done():
			return errors.New("timed out waiting for services")
		}
	}
	return nil
}

func portForward(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, serviceName string, ports []uint16) error {
	// port forward
	var wg sync.WaitGroup
	wg.Add(len(ports))
	for _, port := range ports {
		ep, err := clientset.CoreV1().Endpoints(*namespace).Get(serviceName, meta_v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error listing endpoints for service %s: %v", serviceName, err)
		}
		var podnamespace, podname string
		for _, subset := range ep.Subsets {
			if len(subset.Addresses) == 0 {
				continue
			}
			podnamespace = subset.Addresses[0].TargetRef.Namespace
			podname = subset.Addresses[0].TargetRef.Name
			break
		}
		if podnamespace == "" || podname == "" {
			return fmt.Errorf("no pods are available for service %s", serviceName)
		}
		log.Printf("%s -> %s/%s", serviceName, podnamespace, podname)

		url := clientset.CoreV1().RESTClient().Post().Resource("pods").Namespace(podnamespace).Name(podname).SubResource("portforward").URL()
		transport, upgrader, err := spdy.RoundTripperFor(config)
		if err != nil {
			return fmt.Errorf("Could not create round tripper: %v", err)
		}
		dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
		ports := []string{fmt.Sprintf(":%d", port)}
		readyChan := make(chan struct{}, 1)
		pf, err := portforward.New(dialer, ports, ctx.Done(), readyChan, os.Stderr, os.Stderr)
		if err != nil {
			return fmt.Errorf("Could not port forward into pod: %v", err)
		}
		go func(port uint16) {
			err := pf.ForwardPorts()
			if err != nil {
				log.Fatalf("Could not forward ports for %s:%d : %v", serviceName, port, err)
			}
		}(port)
		go func(port uint16) {
			<-pf.Ready
			ports, err := pf.GetPorts()
			if err != nil {
				log.Fatalf("Could not get forwarded ports for %s:%d : %v", serviceName, port, err)
			}
			for _, port := range ports {
				fmt.Printf("FORWARD %s:%d:%d\n", serviceName, port.Remote, port.Local)
			}
			wg.Done()
		}(port)
	}
	wg.Wait()
	return nil
}

func cleanup(clientset *kubernetes.Clientset) {
	log.Print("Cleanup")
	if *deleteNamespace && *namespace != "" {
		log.Printf("deleting namespace %s", *namespace)
		s := meta_v1.DeletePropagationBackground
		err := clientset.CoreV1().Namespaces().Delete(*namespace, &meta_v1.DeleteOptions{PropagationPolicy: &s})
		if err != nil {
			log.Printf("Unable to delete namespace %s: %v", *namespace, err)
		}
	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)

	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(signalChannel)
		cancel()
	}()
	// cancel context if signal is received
	go func() {
		select {
		case <-signalChannel:
			cancel()
		case <-ctx.Done():
		}
	}()
	// cancel context if stdin is closed
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			_, _, err := reader.ReadRune()
			if err != nil && err == io.EOF {
				cancel()
				break
			}
		}
	}()

	var clientset *kubernetes.Clientset
	if kubeconfig == "" {
		_, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST")
		if !ok {
			kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset = kubernetes.NewForConfigOrDie(config)
	defer cleanup(clientset)

	go stern.Run(ctx, *namespace, clientset)

	if len(waitForApps) > 0 {
		err = waitForPods(ctx, clientset)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if len(pfconfig.services) > 0 {
		err = waitForEndpoints(ctx, clientset, config)
		if err != nil {
			log.Print(err)
			return
		}
	}

	fmt.Println("READY")
	<-ctx.Done()
}
