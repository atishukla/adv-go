package main

import (
	"adv-go/model"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	clientset *kubernetes.Clientset
	wg        sync.WaitGroup
)

func main() {
	// Load Kubernetes configuration
	config, err := loadKubeConfig()
	if err != nil {
		log.Fatalf("Failed to load Kubernetes config: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Determine if we're running inside a Kubernetes cluster
	isInCluster := isRunningInCluster()

	// Start leader election if in a Kubernetes cluster, otherwise directly log pod statuses
	if isInCluster {
		startLeaderElection(clientset)
	} else {
		fmt.Println("Running locally, skipping leader election.")
		logPodStatus(clientset)
	}

	// Block the program so it doesn’t exit immediately. Useful to test leadership
	select {}

}

// Function to check if the app is running inside a Kubernetes cluster
func isRunningInCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}

// logPodStatus retrieves the pod statuses and logs them
func logPodStatus(clientset *kubernetes.Clientset) {
	pods, err := getAllPods(clientset)
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	logFile, err := openLogFile("pod_status.log")
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create a mutex for thread-safe logging
	var mu sync.Mutex
	statusChannel := make(chan string, len(pods.Items))

	// Log each pod's status asynchronously
	for _, pod := range pods.Items {
		wg.Add(1)
		go logPodInfo(pod, logFile, statusChannel, &mu)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(statusChannel)
	}()

	// Collect and print results from the channel
	for status := range statusChannel {
		fmt.Println(status)
	}
}

// getAllPods fetches all pods in the cluster
func getAllPods(clientset *kubernetes.Clientset) (*v1.PodList, error) {
	return clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
}

// logPodInfo logs the status of a single pod
func logPodInfo(pod v1.Pod, logFile *os.File, statusChannel chan<- string, mu *sync.Mutex) {
	defer wg.Done()

	// Create an instance of the Pod struct from the model package
	podModel := model.NewPod(&pod)

	// Format pod information
	status := fmt.Sprintf("Pod Name: %s, Node: %s, Phase: %s",
		podModel.Name(), podModel.NodeName(), podModel.Phase())

	// Log to the file with mutex for thread safety
	mu.Lock()
	if _, err := logFile.WriteString(status + "\n"); err != nil {
		log.Printf("Error writing to log file: %v", err)
	} else {
		log.Println("Logged:", status)
	}
	mu.Unlock()

	// Send the status to the status channel
	statusChannel <- status
}

func startLeaderElection(clientset *kubernetes.Clientset) {
	// Use a leader election
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "leader-election",
			Namespace: "default",
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: os.Getenv("POD_NAME"), // Identity of the POD
		},
	}

	// Leader election callback functions
	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second, // Duration of the leadership
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// Start logging pod status only when this instance is the leader
				log.Println("I am the leader, starting to log pod statuses.")
				logPodStatus(clientset) // Call your function to log pod statuses
			},
			OnStoppedLeading: func() {
				log.Println("Lost leadership, stopping pod status logging.")
			},
			OnNewLeader: func(identity string) {
				// Not necessary but useful for logging purposes
				if identity == os.Getenv("POD_NAME") {
					log.Println("I am still the leader!")
				} else {
					log.Printf("New leader elected: %s\n", identity)
				}
			},
		},
	})
}

// openLogFile opens or creates a log file for writing
func openLogFile(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// loadKubeConfig loads the Kubernetes configuration based on the environment
func loadKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		fmt.Println("Using in-cluster config")
		return config, nil
	} else {
		// Use local kubeconfig for development
		fmt.Println("Using local kubeconfig")
		kubeconfig := flag.String("kubeconfig", filepath.Join(homeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		return clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
}

// homeDir returns the home directory for the user running the program.
func homeDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	return dirname
}
