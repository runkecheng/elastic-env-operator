package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cronhpav1beta1 "github.com/wosai/elastic-env-operator/api/cronhpa/v1beta1"
	qav1alpha1 "github.com/wosai/elastic-env-operator/api/v1alpha1"
	"github.com/wosai/elastic-env-operator/domain/entity"
	"github.com/wosai/elastic-env-operator/domain/handler"
	"github.com/wosai/elastic-env-operator/domain/util"
)

var _ = Describe("Controller", func() {
	namespace := "default"
	applicationName := "default-app"
	planeName := "base"
	var deploymentName = util.GetSubsetName(applicationName, planeName)
	image := "busybox:1.32"
	ctx := context.Background()
	var err error
	var sqbdeployment *qav1alpha1.SQBDeployment
	var sqbplane *qav1alpha1.SQBPlane
	var sqbapplication *qav1alpha1.SQBApplication

	It("create plane success", func() {
		sqbplane := &qav1alpha1.SQBPlane{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      planeName,
			},
			Spec: qav1alpha1.SQBPlaneSpec{
				Description: "test",
			},
		}
		err = k8sClient.Create(ctx, sqbplane)
		time.Sleep(time.Second)
		Expect(err).NotTo(HaveOccurred())
		instance := &qav1alpha1.SQBPlane{}
		err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: planeName}, instance)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Delete(ctx, sqbplane)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(time.Second)
	})

	It("create application success,create service", func() {
		// create application
		sqbapplication := &qav1alpha1.SQBApplication{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      applicationName,
			},
			Spec: qav1alpha1.SQBApplicationSpec{
				ServiceSpec: qav1alpha1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "http-80",
							Port:       int32(80),
							TargetPort: intstr.FromInt(8080),
							Protocol:   "TCP",
						},
					},
				},
				DeploySpec: qav1alpha1.DeploySpec{
					Image: image,
				},
			},
		}
		err = k8sClient.Create(ctx, sqbapplication)
		time.Sleep(time.Second)
		Expect(err).NotTo(HaveOccurred())
		instance := &qav1alpha1.SQBApplication{}
		err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, instance)
		// service success
		service := &corev1.Service{}
		err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, service)
		Expect(err).NotTo(HaveOccurred())
		Expect(service.Spec.Selector[entity.AppKey]).To(Equal(applicationName))
		Expect(service.Spec.Type, corev1.ServiceTypeClusterIP)
		port := service.Spec.Ports[0]
		Expect(port.Name).To(Equal("http-80"))
		Expect(port.Protocol).To(Equal(corev1.ProtocolTCP))
		Expect(port.Port).To(Equal(int32(80)))
		Expect(port.TargetPort).To(Equal(intstr.FromInt(8080)))

		_ = k8sClient.Delete(ctx, sqbapplication)
		_ = k8sClient.Delete(ctx, service)

		time.Sleep(time.Second)
	})

	Describe("CronHPA enabled", func() {
		var cronHPA *cronhpav1beta1.CronHorizontalPodAutoscaler

		BeforeEach(func() {
			// 创建默认的application
			sqbapplication = &qav1alpha1.SQBApplication{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      applicationName,
					Annotations: map[string]string{
						entity.IstioInjectAnnotationKey: "false",
					},
				},
				Spec: qav1alpha1.SQBApplicationSpec{
					IngressSpec: qav1alpha1.IngressSpec{
						Domains: []qav1alpha1.Domain{
							{
								Class: "nginx",
								Host:  fmt.Sprintf("%s.iwosai.com", applicationName),
							},
							{
								Class: "nginx-vpc",
								Host:  fmt.Sprintf("%s.beta.iwosai.com", applicationName),
							},
						},
					},
					ServiceSpec: qav1alpha1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "http-80",
								Port:       int32(80),
								TargetPort: intstr.FromInt(8080),
								Protocol:   "TCP",
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			sqbplane = &qav1alpha1.SQBPlane{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      planeName,
				},
				Spec: qav1alpha1.SQBPlaneSpec{
					Description: planeName,
				},
			}
			err = k8sClient.Create(ctx, sqbplane)
			Expect(err).NotTo(HaveOccurred())
			sqbdeployment = &qav1alpha1.SQBDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      deploymentName,
					Labels: map[string]string{
						entity.AppKey:   applicationName,
						entity.PlaneKey: planeName,
					},
				},
				Spec: qav1alpha1.SQBDeploymentSpec{
					Selector: qav1alpha1.Selector{
						App:   applicationName,
						Plane: planeName,
					},
					DeploySpec: qav1alpha1.DeploySpec{
						Image:    image,
						Replicas: proto.Int32(1),
					},
				},
			}
			err = k8sClient.Create(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())

			sqbdeployment = &qav1alpha1.SQBDeployment{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)

			sqbplane = &qav1alpha1.SQBPlane{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: planeName}, sqbplane)

			deployment := &appv1.Deployment{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, sqbapplication)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, sqbdeployment)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, sqbplane)).Should(Succeed())
			if cronHPA != nil {
				Expect(k8sClient.Delete(ctx, cronHPA)).Should(Succeed())
				cronHPA = nil
			}
			// 删除service,ingress,deployment
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: applicationName}}
			_ = k8sClient.Delete(ctx, service)
			ingress := &v1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: applicationName}}
			_ = k8sClient.Delete(ctx, ingress)
			deployment := &appv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: deploymentName}}
			_ = k8sClient.Delete(ctx, deployment)
			time.Sleep(time.Second)
		})

		It("replicas update should not take effect", func() {
			deployment := &appv1.Deployment{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(deployment.Spec.Replicas).To(Equal(proto.Int32(1)))

			By("Create CronHPA")
			// Create CronHPA
			cronHPA = &cronhpav1beta1.CronHorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      deploymentName,
				},
				Spec: cronhpav1beta1.CronHorizontalPodAutoscalerSpec{
					Jobs: []cronhpav1beta1.Job{
						{
							Name:       "每天中午10：12扩容",
							Schedule:   "0 12 10 * * *",
							TargetSize: 10,
						},
					},
					ScaleTargetRef: cronhpav1beta1.ScaleTargetRef{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       deploymentName,
					},
				},
			}
			err = k8sClient.Create(ctx, cronHPA)
			Expect(err).NotTo(HaveOccurred())

			cronHPA = &cronhpav1beta1.CronHorizontalPodAutoscaler{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, cronHPA)

			sqbdeployment = &qav1alpha1.SQBDeployment{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			sqbdeployment.Spec.Replicas = proto.Int32(2)
			Expect(k8sClient.Update(ctx, sqbdeployment)).Should(Succeed())

			Consistently(func() int32 {
				_ = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
				return *deployment.Spec.Replicas
			}).Should(BeNumerically("==", 1))
		})

		It("replicas update should take effect without CronHPA", func() {
			deployment := &appv1.Deployment{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(deployment.Spec.Replicas).To(Equal(proto.Int32(1)))

			sqbdeployment = &qav1alpha1.SQBDeployment{}
			testObjExistence(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			sqbdeployment.Spec.Replicas = proto.Int32(2)
			Expect(k8sClient.Update(ctx, sqbdeployment)).Should(Succeed())

			Eventually(func() int32 {
				_ = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
				return *deployment.Spec.Replicas
			}).Should(BeNumerically("==", 2))
		})
	})

	Describe("istio disabled", func() {
		BeforeEach(func() {
			// 创建默认的application
			sqbapplication = &qav1alpha1.SQBApplication{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      applicationName,
					Annotations: map[string]string{
						entity.IstioInjectAnnotationKey: "false",
					},
				},
				Spec: qav1alpha1.SQBApplicationSpec{
					IngressSpec: qav1alpha1.IngressSpec{
						Domains: []qav1alpha1.Domain{
							{
								Class: "nginx",
								Host:  fmt.Sprintf("%s.iwosai.com", applicationName),
							},
							{
								Class: "nginx-vpc",
								Host:  fmt.Sprintf("%s.beta.iwosai.com", applicationName),
							},
						},
					},
					ServiceSpec: qav1alpha1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "http-80",
								Port:       int32(80),
								TargetPort: intstr.FromInt(8080),
								Protocol:   "TCP",
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			sqbplane = &qav1alpha1.SQBPlane{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      planeName,
				},
				Spec: qav1alpha1.SQBPlaneSpec{
					Description: planeName,
				},
			}
			err = k8sClient.Create(ctx, sqbplane)
			Expect(err).NotTo(HaveOccurred())
			sqbdeployment = &qav1alpha1.SQBDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      deploymentName,
					Labels: map[string]string{
						entity.AppKey:   applicationName,
						entity.PlaneKey: planeName,
					},
				},
				Spec: qav1alpha1.SQBDeploymentSpec{
					Selector: qav1alpha1.Selector{
						App:   applicationName,
						Plane: planeName,
					},
					DeploySpec: qav1alpha1.DeploySpec{
						Image: image,
					},
				},
			}
			err = k8sClient.Create(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			sqbdeployment = &qav1alpha1.SQBDeployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			sqbplane = &qav1alpha1.SQBPlane{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: planeName}, sqbplane)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_ = k8sClient.Delete(ctx, sqbapplication)
			_ = k8sClient.Delete(ctx, sqbdeployment)
			_ = k8sClient.Delete(ctx, sqbplane)
			// 删除service,ingress,deployment
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: applicationName}}
			_ = k8sClient.Delete(ctx, service)
			ingress := &v1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: applicationName}}
			_ = k8sClient.Delete(ctx, ingress)
			deployment := &appv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: deploymentName}}
			_ = k8sClient.Delete(ctx, deployment)
			time.Sleep(time.Second)
		})

		It("deployment reconcile success", func() {
			sqbdeployment.Spec = qav1alpha1.SQBDeploymentSpec{
				Selector: qav1alpha1.Selector{
					App:   applicationName,
					Plane: planeName,
				},
				DeploySpec: qav1alpha1.DeploySpec{
					Replicas: proto.Int32(2),
					Image:    image,
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: *resource.NewQuantity(2, resource.DecimalSI),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: *resource.NewQuantity(1, resource.DecimalSI),
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "env1",
							Value: "value1",
						},
					},
					HealthCheck: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Port: intstr.FromInt(8080),
								Path: "/healthy",
							},
						},
						InitialDelaySeconds: 10,
						TimeoutSeconds:      10,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						FailureThreshold:    1,
					},
					NodeAffinity: &qav1alpha1.NodeAffinity{
						Prefer: []qav1alpha1.NodeSelector{
							{
								Weight: 100,
								NodeSelectorRequirement: corev1.NodeSelectorRequirement{
									Key:      "node",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"qa"},
								},
							},
						},
						Require: []qav1alpha1.NodeSelector{
							{
								Weight: 100,
								NodeSelectorRequirement: corev1.NodeSelectorRequirement{
									Key:      "node",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"qa"},
								},
							},
						},
					},
					Lifecycle: &qav1alpha1.Lifecycle{
						Init: &qav1alpha1.InitHandler{Exec: &corev1.ExecAction{Command: []string{"sleep", "1"}}},
						Lifecycle: corev1.Lifecycle{
							PostStart: &corev1.LifecycleHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Port: intstr.FromInt(8080),
									Path: "/poststart",
								},
							},
							PreStop: &corev1.LifecycleHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.FromInt(8080),
								},
							},
						},
					},
					Volumes: []*qav1alpha1.VolumeSpec{
						{
							MountPath: "/tmp",
							HostPath:  "/tmp",
						},
					},
				},
			}
			err = k8sClient.Update(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)

			deployment := &appv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(image))
			Expect(deployment.Labels[entity.AppKey]).To(Equal(applicationName))
			Expect(deployment.Labels[entity.PlaneKey]).To(Equal(planeName))
			Expect(deployment.Spec.Template.Labels[entity.AppKey]).To(Equal(applicationName))
			Expect(deployment.Spec.Template.Labels[entity.PlaneKey]).To(Equal(planeName))
			Expect(deployment.Spec.Selector.MatchLabels[entity.AppKey]).To(Equal(applicationName))
			Expect(deployment.Spec.Replicas).To(Equal(proto.Int32(2)))
			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container.Resources.Limits.Cpu().String()).To(Equal("2"))
			Expect(container.Resources.Requests.Cpu().String()).To(Equal("1"))
			Expect(container.Env[0].Name).To(Equal("env1"))
			Expect(container.Env[0].Value).To(Equal("value1"))
			Expect(container.LivenessProbe.InitialDelaySeconds).To(Equal(int32(10)))
			Expect(container.LivenessProbe.PeriodSeconds).To(Equal(int32(10)))
			Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/healthy"))
			Expect(container.LivenessProbe.HTTPGet.Port).To(Equal(intstr.FromInt(8080)))
			Expect(container.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(10)))
			Expect(container.ReadinessProbe.PeriodSeconds).To(Equal(int32(10)))
			Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/healthy"))
			Expect(container.ReadinessProbe.HTTPGet.Port).To(Equal(intstr.FromInt(8080)))
			Expect(container.Lifecycle.PostStart.HTTPGet.Port).To(Equal(intstr.FromInt(8080)))
			Expect(container.Lifecycle.PostStart.HTTPGet.Path).To(Equal("/poststart"))
			Expect(container.Lifecycle.PreStop.TCPSocket.Port).To(Equal(intstr.FromInt(8080)))
			initContainer := deployment.Spec.Template.Spec.InitContainers[0]
			Expect(initContainer.Image).To(Equal(image))
			Expect(initContainer.Command).To(Equal([]string{"sleep", "1"}))
			Expect(container.VolumeMounts[0].Name).To(Equal("volume-0"))
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/tmp"))
			required := deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			Expect(required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("node"))
			Expect(required.NodeSelectorTerms[0].MatchExpressions[0].Values[0]).To(Equal("qa"))
			preferred := deployment.Spec.Template.Spec.Affinity.NodeAffinity.
				PreferredDuringSchedulingIgnoredDuringExecution[0]
			Expect(preferred.Preference.MatchExpressions[0].Key).To(Equal("node"))
			Expect(preferred.Preference.MatchExpressions[0].Values[0]).To(Equal("qa"))
			Expect(preferred.Weight).To(Equal(int32(100)))
			Expect(deployment.Spec.Template.Spec.Volumes[0].Name).To(Equal("volume-0"))
			Expect(deployment.Spec.Template.Spec.Volumes[0].HostPath.Path).To(Equal("/tmp"))
		})

		It("ingress close", func() {
			_ = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			sqbapplication.Annotations[entity.IngressOpenAnnotationKey] = "false"
			err = k8sClient.Update(ctx, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			ingress := &v1.Ingress{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, ingress)
			Expect(err).To(HaveOccurred())
		})

		It("ingress open", func() {
			_ = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			sqbapplication.Annotations[entity.IngressOpenAnnotationKey] = "true"
			sqbapplication.Spec.Subpaths = []qav1alpha1.Subpath{
				{
					Path:        "/v1",
					ServiceName: "version1",
					ServicePort: 8080,
				},
			}
			err = k8sClient.Update(ctx, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			for _, domain := range sqbapplication.Spec.Domains {
				ingress := &v1.Ingress{}
				err = k8sClient.Get(
					ctx,
					types.NamespacedName{
						Namespace: namespace,
						Name:      handler.GetIngressName(applicationName, domain.Class, domain.Host),
					},
					ingress,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(ingress.Spec.Rules[0].Host).To(Equal(entity.ConfigMapData.GetDomainNameByClass(applicationName, domain.Class)))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal("version1"))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/v1"))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[1].Backend.Service.Name).To(Equal(applicationName))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[1].Path).To(Equal(""))
			}
		})

		It("pass deployment annotation,pod annotation", func() {
			if sqbdeployment.Annotations == nil {
				sqbdeployment.Annotations = make(map[string]string)
			}
			sqbdeployment.Annotations[entity.DeploymentAnnotationKey] = `{"type":"deployment"}`
			sqbdeployment.Annotations[entity.PodAnnotationKey] = `{"type":"pod"}`
			err = k8sClient.Update(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			deployment := &appv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.Annotations["type"]).To(Equal("deployment"))
			Expect(deployment.Spec.Template.Annotations["type"]).To(Equal("pod"))
		})

		It("pass ingress annotation,service annotation", func() {
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			sqbapplication.Annotations[entity.ServiceAnnotationKey] = `{"type":"service"}`
			sqbapplication.Annotations[entity.IngressOpenAnnotationKey] = "true"
			for i, domain := range sqbapplication.Spec.Domains {
				domain.Annotation = map[string]string{
					"type": "ingress",
				}
				sqbapplication.Spec.Domains[i] = domain
			}
			err = k8sClient.Update(ctx, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			service := &corev1.Service{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, service)
			Expect(service.Annotations["type"]).To(Equal("service"))
			for _, domain := range sqbapplication.Spec.Domains {
				ingress := &v1.Ingress{}
				err = k8sClient.Get(ctx,
					types.NamespacedName{
						Namespace: namespace,
						Name:      handler.GetIngressName(sqbapplication.Name, domain.Class, domain.Host),
					},
					ingress,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(ingress.Annotations["type"]).To(Equal("ingress"))
			}
		})

		It("delete sqbapplication without password", func() {
			// sqbdeployment不删除
			err := k8sClient.Delete(ctx, &qav1alpha1.SQBApplication{ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace, Name: applicationName,
			}})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			Expect(err).To(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
		})

		It("delete sqbdeployment without password", func() {
			// deployment不删除，sqbapplication的status不变
			err = k8sClient.Delete(ctx, &qav1alpha1.SQBDeployment{
				ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).To(HaveOccurred())
			deployment := &appv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			_, ok := sqbapplication.Status.Mirrors[deploymentName]
			Expect(ok).To(BeTrue())
		})

		It("delete sqbplane without password", func() {
			// sqbdeployment不删除
			err = k8sClient.Delete(ctx, &qav1alpha1.SQBPlane{
				ObjectMeta: metav1.ObjectMeta{Name: planeName, Namespace: namespace},
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: planeName}, sqbplane)
			Expect(err).To(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
		})

		It("delete sqbapplication with password", func() {
			_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, sqbapplication, func() error {
				sqbapplication.Annotations[entity.ExplicitDeleteAnnotationKey] = util.GetDeleteCheckSum(sqbapplication.Name)
				sqbapplication.Annotations[entity.IngressOpenAnnotationKey] = "true"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			// sqbapplication，sqbdeployment被删除
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			Expect(err).To(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).To(HaveOccurred())
			// deployment,ingress和service被删除
			deployment := &appv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(err).To(HaveOccurred())
			ingress := &v1.Ingress{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, ingress)
			Expect(err).To(HaveOccurred())
			service := &corev1.Service{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, service)
			Expect(err).To(HaveOccurred())
		})
	})

	// deprecated
	XDescribe("istio enabled", func() {
		BeforeEach(func() {
			// 创建默认的application
			sqbapplication = &qav1alpha1.SQBApplication{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      applicationName,
					Annotations: map[string]string{
						entity.IstioInjectAnnotationKey: "true",
					},
				},
				Spec: qav1alpha1.SQBApplicationSpec{
					IngressSpec: qav1alpha1.IngressSpec{
						Domains: []qav1alpha1.Domain{
							{
								Class: "nginx",
								Host:  fmt.Sprintf("%s.iwosai.com", applicationName),
							},
							{
								Class: "nginx-vpc",
								Host:  fmt.Sprintf("%s.beta.iwosai.com", applicationName),
							},
						},
					},
					ServiceSpec: qav1alpha1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "http-80",
								Port:       int32(80),
								TargetPort: intstr.FromInt(8080),
								Protocol:   "TCP",
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			sqbplane = &qav1alpha1.SQBPlane{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      planeName,
				},
				Spec: qav1alpha1.SQBPlaneSpec{
					Description: planeName,
				},
			}
			err = k8sClient.Create(ctx, sqbplane)
			Expect(err).NotTo(HaveOccurred())
			sqbdeployment = &qav1alpha1.SQBDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      deploymentName,
					Labels: map[string]string{
						entity.AppKey:   applicationName,
						entity.PlaneKey: planeName,
					},
				},
				Spec: qav1alpha1.SQBDeploymentSpec{
					Selector: qav1alpha1.Selector{
						App:   applicationName,
						Plane: planeName,
					},
					DeploySpec: qav1alpha1.DeploySpec{
						Image: image,
					},
				},
			}
			err = k8sClient.Create(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			sqbdeployment = &qav1alpha1.SQBDeployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			sqbplane = &qav1alpha1.SQBPlane{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "base"}, sqbplane)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_ = k8sClient.Delete(ctx, sqbdeployment)
			_ = k8sClient.Delete(ctx, sqbapplication)
			_ = k8sClient.Delete(ctx, sqbplane)
			// 删除service,ingress,deployment,virtualservice,destinationrule
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: applicationName}}
			_ = k8sClient.Delete(ctx, service)
			ingress := &v1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: applicationName}}
			_ = k8sClient.Delete(ctx, ingress)
			_ = k8sClient.DeleteAllOf(ctx, &appv1.Deployment{}, &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{entity.AppKey: applicationName}),
					Namespace:     namespace,
				},
			})
			time.Sleep(time.Second)
		})

		It("ingress open", func() {
			// ingress指向istio-ingressgateway
			_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, sqbapplication, func() error {
				sqbapplication.Annotations[entity.IngressOpenAnnotationKey] = "true"
				sqbapplication.Annotations[entity.IstioInjectAnnotationKey] = "true"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			for _, domain := range sqbapplication.Spec.Domains {
				ingress := &v1.Ingress{}
				err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName + "-" + domain.Class}, ingress)
				Expect(ingress.Spec.Rules[0].Host).To(Equal(entity.ConfigMapData.GetDomainNameByClass(applicationName, domain.Class)))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal("istio-ingressgateway-" + namespace))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal(intstr.FromInt(80)))
			}
		})

		It("public entry", func() {
			_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, sqbdeployment, func() error {
				if sqbdeployment.Annotations == nil {
					sqbdeployment.Annotations = make(map[string]string)
				}
				sqbdeployment.Annotations[entity.PublicEntryAnnotationKey] = "true"
				return nil
			})
			err = k8sClient.Update(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			// 断言ingress
			ingress := &v1.Ingress{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName + "-" + handler.SpecialVirtualServiceIngress(sqbdeployment)}, ingress)
			Expect(len(ingress.Spec.Rules)).To(Equal(2))
			// 关闭入口
			_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, sqbdeployment, func() error {
				sqbdeployment.Annotations[entity.PublicEntryAnnotationKey] = "false"
				return nil
			})
			err = k8sClient.Update(ctx, sqbdeployment)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName + "-" + handler.SpecialVirtualServiceIngress(sqbdeployment)}, ingress)
			Expect(len(ingress.Spec.Rules)).To(Equal(1))
		})

		It("multi subpaths,multi plane", func() {
			_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, sqbapplication, func() error {
				sqbapplication.Spec.Subpaths = []qav1alpha1.Subpath{
					{
						Path:        "/v2",
						ServiceName: "version2",
						ServicePort: 82,
					},
				}
				sqbapplication.Annotations[entity.IstioInjectAnnotationKey] = "true"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			// 加一个sqbplane和sqbdeployment
			plane2 := &qav1alpha1.SQBPlane{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "test"},
				Spec:       qav1alpha1.SQBPlaneSpec{Description: "test"},
			}
			err = k8sClient.Create(ctx, plane2)
			Expect(err).NotTo(HaveOccurred())
			sqbdeployment2 := &qav1alpha1.SQBDeployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: util.GetSubsetName(applicationName, "test")},
				Spec: qav1alpha1.SQBDeploymentSpec{
					Selector: qav1alpha1.Selector{
						App:   applicationName,
						Plane: "test",
					},
					DeploySpec: qav1alpha1.DeploySpec{
						Image: image,
					},
				},
			}
			err = k8sClient.Create(ctx, sqbdeployment2)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			// sqbapplication的status正确
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: applicationName}, sqbapplication)
			Expect(err).NotTo(HaveOccurred())
			_, ok := sqbapplication.Status.Planes["base"]
			Expect(ok).To(BeTrue())
			_, ok = sqbapplication.Status.Planes["test"]
		})

		It("tcp service port", func() {
			_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, sqbapplication, func() error {
				sqbapplication.Spec.Ports = []corev1.ServicePort{
					{
						Name:       "tcp-3306",
						Port:       int32(3306),
						TargetPort: intstr.FromInt(3306),
						Protocol:   "TCP",
					},
				}
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			service := &corev1.Service{}
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: applicationName}, service)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(service.Spec.Ports)).To(Equal(1))
			Expect(service.Spec.Ports[0].Name).To(Equal("tcp-3306"))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(3306)))
		})

		It("delete sqbapplication with password", func() {
			_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, sqbapplication, func() error {
				sqbapplication.Annotations[entity.ExplicitDeleteAnnotationKey] = util.GetDeleteCheckSum(sqbapplication.Name)
				sqbapplication.Annotations[entity.IstioInjectAnnotationKey] = "true"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
		})

		It("istio open then close", func() {
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).To(BeNil())
			Expect(sqbdeployment.Annotations[entity.IstioInjectAnnotationKey]).To(Equal("true"))
			deployment := &appv1.Deployment{}
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(err).To(BeNil())
			Expect(deployment.Spec.Template.Annotations[entity.IstioSidecarInjectKey]).To(Equal("true"))

			_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, sqbapplication, func() error {
				sqbapplication.Annotations[entity.IstioInjectAnnotationKey] = "false"
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: deploymentName}, sqbdeployment)
			Expect(err).To(BeNil())
			Expect(sqbdeployment.Annotations[entity.IstioInjectAnnotationKey]).To(Equal("false"))

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: deploymentName}, deployment)
			Expect(err).To(BeNil())
			Expect(deployment.Spec.Template.Annotations[entity.IstioSidecarInjectKey]).To(Equal("false"))

		})

	})
})

func testObjExistence(ctx context.Context, name types.NamespacedName, obj client.Object) {
	Eventually(func() error {
		return k8sClient.Get(ctx, name, obj)
	}).WithPolling(500 * time.Millisecond).WithTimeout(5 * time.Second).Should(Succeed())
}
