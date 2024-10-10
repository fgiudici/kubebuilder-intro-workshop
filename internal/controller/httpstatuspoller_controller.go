/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	healthcheckerv1 "cattle.io/workshop/api/v1"
)

// HttpStatusPollerReconciler reconciles a HttpStatusPoller object
type HttpStatusPollerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=healthchecker.cattle.io,resources=httpstatuspollers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=healthchecker.cattle.io,resources=httpstatuspollers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=healthchecker.cattle.io,resources=httpstatuspollers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HttpStatusPoller object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *HttpStatusPollerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	httpStatusPoller := &healthcheckerv1.HttpStatusPoller{}
	if err := r.Get(ctx, req.NamespacedName, httpStatusPoller); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patch := client.MergeFrom(httpStatusPoller.DeepCopy())

	// Poll the URLs specified in the CR
	result := pollURLs(ctx, httpStatusPoller.Spec.URLs)

	// Update the status of the CR with the results of the polling
	httpStatusPoller.Status.StatusCodes = result

	if err := r.Status().Patch(ctx, httpStatusPoller, patch); err != nil {
		return ctrl.Result{}, err
	}

	pollingInterval := time.Duration(httpStatusPoller.Spec.IntervalSeconds) * time.Second
	if httpStatusPoller.Spec.IntervalSeconds == 0 {
		pollingInterval = time.Minute
	}

	return ctrl.Result{RequeueAfter: pollingInterval}, nil
}

func pollURLs(ctx context.Context, urls []string) map[string]int {
	log := log.FromContext(ctx)

	result := make(map[string]int)
	http.DefaultClient.Timeout = 2 * time.Second

	for _, u := range urls {
		resp, err := http.Get(u)
		if err != nil {
			if err.(*url.Error).Timeout() {
				log.Info("URL is unreachable", "url", u)
			} else {
				log.Error(err, "Failed to poll URL", "url", u)
			}
			result[u] = -1
			continue
		}
		defer resp.Body.Close()
		log.Info("URL is reachable", "url", u, "HTTP Status code", resp.Status)
		result[u] = resp.StatusCode
	}

	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *HttpStatusPollerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&healthcheckerv1.HttpStatusPoller{}).
		Complete(r)
}
