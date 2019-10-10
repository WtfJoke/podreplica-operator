package podreplica

import (
	"context"
	"reflect"

	appv1alpha1 "github.com/wtfjoke/pod-replica/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_podreplica")

// Add creates a new PodReplica Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodReplica{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podreplica-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PodReplica
	err = c.Watch(&source.Kind{Type: &appv1alpha1.PodReplica{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PodReplica
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.PodReplica{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePodReplica implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePodReplica{}

// ReconcilePodReplica reconciles a PodReplica object
type ReconcilePodReplica struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PodReplica object and makes changes based on the state read
// and what is in the PodReplica.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePodReplica) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("⚖️ Reconciling PodReplica")

	// Fetch the PodReplica instance
	podReplica := &appv1alpha1.PodReplica{}
	err := r.client.Get(context.TODO(), request.NamespacedName, podReplica)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Fetch existing replica pods
	existingReplicaPods := &corev1.PodList{}
	existingReplicaPodsQuery := &client.ListOptions{
		Namespace:     request.Namespace,
		LabelSelector: labels.SelectorFromSet(labelsForPodReplicas(request.Name)),
	}

	err = r.client.List(context.TODO(), existingReplicaPodsQuery, existingReplicaPods)
	if err != nil {
		reqLogger.Error(err, "💥 Failed to list existing pods in the podreplica")
		return reconcile.Result{}, err
	}

	runningReplicaPodNames := getRunningReplicaPodNames(existingReplicaPods.Items)

	// Update pod names (in status)
	if !reflect.DeepEqual(podReplica.Status.Replicas, runningReplicaPodNames) {
		podReplica.Status.Replicas = runningReplicaPodNames
		err := r.client.Update(context.TODO(), podReplica)
		if err != nil {
			reqLogger.Error(err, "💥 Failed to update podreplica")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("👀 Checking podreplicas", "expected replicas", podReplica.Spec.Size, "Pod.Names", runningReplicaPodNames)

	// scale down replicas
	if int32(len(runningReplicaPodNames)) > podReplica.Spec.Size {
		podToBeDeleted := existingReplicaPods.Items[0]
		reqLogger.Info("💀 Deleting a pod", "Pod.Namespace", podToBeDeleted.Namespace, "Pod.Name", podToBeDeleted.Name)
		err := r.client.Delete(context.TODO(), &podToBeDeleted)
		if err != nil {
			reqLogger.Error(err, "💥 Failed to delete a pod")
			return reconcile.Result{}, err
		}
	}

	// scale up replicas
	if int32(len(runningReplicaPodNames)) > podReplica.Spec.Size {
		// Define a new Pod object
		pod := newPodForCR(podReplica)
		reqLogger.Info("👶 Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "💥 Failed to create a pod")
			return reconcile.Result{}, err
		}

		// Set PodReplica instance as the owner and controller
		if err := controllerutil.SetControllerReference(podReplica, pod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
	}

	if int32(len(runningReplicaPodNames)) == podReplica.Spec.Size {
		reqLogger.Info("⏭️ Skip reconcile: Enough replicas")
		return reconcile.Result{}, nil
	}

	// requeue for additional scaling or updating pod names
	return reconcile.Result{Requeue: true}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *appv1alpha1.PodReplica) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func labelsForPodReplicas(name string) map[string]string {
	return map[string]string{"app": name}
}

func getRunningReplicaPodNames(existingReplicaPods []corev1.Pod) []string {
	runningReplicaPodNames := []string{}
	for _, replicaPod := range existingReplicaPods {
		if replicaPod.DeletionTimestamp != nil { // ignore pods beeing terminated
			continue
		}
		if replicaPod.Status.Phase == corev1.PodPending || replicaPod.Status.Phase == corev1.PodRunning {
			runningReplicaPodNames = append(runningReplicaPodNames, replicaPod.Name)
		}
	}
	return runningReplicaPodNames
}
