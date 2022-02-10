/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/nstogner/gcp-ilb-redirect-controller/cloud"
	"github.com/nstogner/gcp-ilb-redirect-controller/cloudresources"

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	clog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// StaticIPAnnotation is used by gce-ingress.
	StaticIPAnnotation = "kubernetes.io/ingress.regional-static-ip-name"

	RedirectAnnotation = "networking.gke.io/ilb-https-redirect"
	Finalizer          = "networking.gke.io/ilb-redirect"
)

// IngressReconciler reconciles a PlaceholderKind object
type IngressReconciler struct {
	Project string
	Region  string

	client.Client
	Scheme *runtime.Scheme

	CloudClient *cloud.Client
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx)

	log.Info("Reconciling")
	defer log.Info("Done reconciling")

	var ing netv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &ing); err != nil {
		// Ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ignore Ingress resources that do not have the redirect annotation.
	if ing.GetAnnotations() == nil {
		return ctrl.Result{}, nil
	}
	if _, ok := ing.GetAnnotations()[RedirectAnnotation]; !ok {
		return ctrl.Result{}, nil
	}
	if _, ok := ing.GetAnnotations()[StaticIPAnnotation]; !ok {
		return ctrl.Result{}, fmt.Errorf("a shared static IP is needed to create the https redirect forwarding rule, missing annotation: %s", StaticIPAnnotation)
	}

	if len(ing.Spec.Rules) == 0 {
		log.Info("No rules in ingress, ignoring")
		return ctrl.Result{}, nil
	}

	info := cloudresources.Info{
		GeneratedName: externalName(&ing),
		Project:       r.Project,
		Region:        r.Region,
		Hostname:      ing.Spec.Rules[0].Host,
	}

	cloudResources := []cloudResource{
		&cloudresources.BackendServices{Client: r.CloudClient},
		&cloudresources.UrlMaps{Client: r.CloudClient},
		&cloudresources.TargetHttpProxies{Client: r.CloudClient},
		&cloudresources.ForwardingRules{Client: r.CloudClient},
	}

	if ing.GetDeletionTimestamp() == nil {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(ing.GetFinalizers(), Finalizer) {
			log.Info("Adding finalizer", "finalizer", Finalizer)
			controllerutil.AddFinalizer(&ing, Finalizer)
			if err := r.Update(ctx, &ing); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating ingress with finalizer: %w", err)
			}
		}
	} else {
		log.Info("Ingress is in deletion state")

		// Remove cloud resources in reverse order.
		for i := len(cloudResources) - 1; i >= 0; i-- {
			cr := cloudResources[i]
			if err := remove(ctx, info, cr); err != nil {
				if err == errNotReady {
					wait := 3 * time.Second
					log.Info("Resource not ready for deletion, requeuing", "wait", wait)
					return ctrl.Result{RequeueAfter: wait}, nil
				}
				return ctrl.Result{}, fmt.Errorf("removing %T: %w", cr, err)
			}
		}

		if containsString(ing.GetFinalizers(), Finalizer) {
			log.Info("Done with resource deletions, removing finalizer")
			controllerutil.RemoveFinalizer(&ing, Finalizer)
			if err := r.Update(ctx, &ing); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if ingStatus := ing.Status.LoadBalancer.Ingress; len(ingStatus) > 0 && ingStatus[0].IP != "" {
		info.IP = ingStatus[0].IP
	} else {
		log.Info("Waiting for Ingress to be updated with an IP address")
		return ctrl.Result{}, nil
	}

	for _, cr := range cloudResources {
		if err := ensure(ctx, info, cr); err != nil {
			if err == errNotReady {
				wait := 3 * time.Second
				log.Info("Resource not ready, requeuing", "wait", wait)
				return ctrl.Result{RequeueAfter: wait}, nil
			}
			return ctrl.Result{}, fmt.Errorf("ensuring %T: %w", cr, err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Ingress{}).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Complete(r)
}
