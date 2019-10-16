package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	apis "github.com/wtfjoke/pod-replica/pkg/apis"
	v1alpha "github.com/wtfjoke/pod-replica/pkg/apis/app/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
	ctx                  *framework.TestCtx
)

func TestPodReplica(t *testing.T) {
	setUp(t)

	t.Run("PodReplica-ScaleTest", scaleUpAndDownPodReplicaTest)
}

func setUp(t *testing.T) {
	registerCRD(t)
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	InitializeCluster(t)
}

func registerCRD(t *testing.T) {
	podReplicaList := &v1alpha.PodReplicaList{}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, podReplicaList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
}

func scaleUpAndDownPodReplicaTest(t *testing.T) {
	f := framework.Global
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(fmt.Errorf("could not get namespace: %v", err))
	}
	// create pod replica custom resource
	examplePodReplica := &v1alpha.PodReplica{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-podreplica",
			Namespace: namespace,
		},
		Spec: v1alpha.PodReplicaSpec{
			Size: 3,
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(context.TODO(), examplePodReplica, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}
	// wait for example-podreplica to reach 3 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-podreplica", 3, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "example-podreplica", Namespace: namespace}, examplePodReplica)
	if err != nil {
		t.Fatal(err)
	}
	examplePodReplica.Spec.Size = 4
	err = f.Client.Update(context.TODO(), examplePodReplica)
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-podreplica to reach 4 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-podreplica", 4, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

}

func InitializeCluster(t *testing.T) {
	t.Parallel()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for podreplica-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "podreplica-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

}

func waitForPodCreation(t *testing.T, kubeclient kubernetes.Interface, namespace, nameprefix string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		pod, err := kubeclient.CoreV1().Pods(namespace).Get(nameprefix, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s deployment\n", nameprefix)
				return false, nil
			}
			return false, err
		}
		t.Log(pod.Name)
		// if int(deployment.Status.AvailableReplicas) == replicas {
		// 	return true, nil
		// }
		// t.Logf("Waiting for full availability of %s deployment (%d/%d)\n", nameprefix, deployment.Status.AvailableReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Deployment available (%d/%d)\n", replicas, replicas)
	return nil
}
