/*
Copyright 2026.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	computev1 "github.com/xonas1101/controller-test/api/v1"
)

// EC2InstanceReconciler reconciles a EC2Instance object
type EC2InstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=compute.example.com,resources=ec2instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.example.com,resources=ec2instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=compute.example.com,resources=ec2instances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EC2Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *EC2InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := logf.FromContext(ctx)
	l.Info("=== RECONCILE LOOP STARTED ===", "namespace", req.Namespace, "name", req.Name)
	ec2Instance := &computev1.EC2Instance{}

	if err:= r.Get(ctx, req.NamespacedName, ec2Instance); err!=nil{
		if errors.IsNotFound(err){
			l.Info("Instance deleted. No need to reconcile.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !ec2Instance.DeletionTimestamp.IsZero(){
		l.Info("Has Deletion Timestamp, Instance is being deleted")
		_,err:= deleteEc2Instance(ctx, ec2Instance)
		if err!=nil{
			l.Error(err, "Failed to delete EC2 Instance")
			return ctrl.Result{Requeue: true}, err
		}

		controllerutil.RemoveFinalizer(ec2Instance, "ec2instance.compute.example.com")
		if err := r.Update(ctx, ec2Instance); err!=nil{
			l.Error(err, "Failed to remove finalizer")
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	l.Info("=== ABOUT TO ADD FINALIZER ===")
	ec2Instance.Finalizers = append(ec2Instance.Finalizers, "ec2instance.compute.exmaple.com")
	if err := r.Update(ctx, ec2Instance); err!=nil{
		l.Error(err, "failed to add finalizer")
		return ctrl.Result{Requeue: true}, err
	}
	l.Info("=== FINALIZER ADDED ===")
	l.Info("=== CONTINUING WITH EC2 INSTANCE CREATION IN CURRENT RECONCILE")

	createdInstanceInfo, err:= createEc2Instance(ec2Instance)
	if err!=nil{
		l.Error(err, "Failed to create EC2 Instance")
		return ctrl.Result{}, err
	}

	l.Info("=== ABOUT TO UPDATE STATUS ===")
}

// SetupWithManager sets up the controller with the Manager.
func (r *EC2InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1.EC2Instance{}).
		Named("ec2instance").
		Complete(r)
}
