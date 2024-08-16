/*


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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cronhpav1beta1 "github.com/wosai/elastic-env-operator/api/cronhpa/v1beta1"
	qav1alpha1 "github.com/wosai/elastic-env-operator/api/v1alpha1"
	"github.com/wosai/elastic-env-operator/domain/common"
	"github.com/wosai/elastic-env-operator/domain/handler"
)

var cronHPARefIndexFunc = func(obj interface{}) ([]string, error) {
	cronHPA := obj.(*cronhpav1beta1.CronHorizontalPodAutoscaler)

	return []string{fmt.Sprintf("%s.%s.%s.%s", cronHPA.Spec.ScaleTargetRef.ApiVersion, cronHPA.Spec.ScaleTargetRef.Kind, cronHPA.GetNamespace(), cronHPA.Spec.ScaleTargetRef.Name)}, nil
}

// sqbDeploymentReconciler reconciles a SQBDeployment object
type sqbDeploymentReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	cronHPAIndexer cache.Indexer
}

func NewSQBDeploymentReconciler(mgr ctrl.Manager) error {
	cacher := mgr.GetCache()
	cronHPAInformer, err := cacher.GetInformerForKind(context.TODO(), cronhpav1beta1.SchemeGroupVersion.WithKind("CronHorizontalPodAutoscaler"))
	if err != nil {
		return err
	}

	if innerErr := cronHPAInformer.AddIndexers(map[string]cache.IndexFunc{
		common.CronHPAIndexByRef: cronHPARefIndexFunc,
	}); innerErr != nil {
		return innerErr
	}

	reconciler := &sqbDeploymentReconciler{
		Client:         mgr.GetClient(),
		Log:            ctrl.Log.WithName("controllers").WithName("SQBDeployment"),
		Scheme:         mgr.GetScheme(),
		cronHPAIndexer: cronHPAInformer.(cache.SharedIndexInformer).GetIndexer(),
	}

	return reconciler.setupWithManager(mgr)
}

// +kubebuilder:rbac:groups=qa.shouqianba.com,resources=sqbdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qa.shouqianba.com,resources=sqbdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaling.alibabacloud.com,resources=cronhorizontalpodautoscalers,verbs=get;list;watch

func (r *sqbDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return handler.HandleReconcile(handler.NewSqbDeploymentHandler(req, ctx, r.cronHPAIndexer))
}

func (r *sqbDeploymentReconciler) setupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qav1alpha1.SQBDeployment{}, builder.WithPredicates(GenerationAnnotationPredicate)).
		Complete(r)
}
