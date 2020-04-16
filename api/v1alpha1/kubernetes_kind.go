package v1alpha1

//KubernetesKind comes from the Object spinnakerKindMap in call: http://localhost:8084/credentials?expand=true.
//Also this can be found in class  /clouddriver/clouddriver-kubernetes-v2/src/main/java/com/netflix/spinnaker/clouddriver/kubernetes/v2/description/manifest/KubernetesKind.java
// +kubebuilder:validation:Enum=apiService;clusterRole;clusterRoleBinding;configMap;controllerRevision;cronJob;customResourceDefinition;daemonSet;deployment;event;horizontalpodautoscaler;ingress;job;limitRange;mutatingWebhookConfiguration;namespace;networkPolicy;persistentVolume;persistentVolumeClaim;pod;podDisruptionBudget;podPreset;podSecurityPolicy;replicaSet;role;roleBinding;secret;service;serviceAccount;statefulSet;storageClass;validatingWebhookConfiguration
type KubernetesKind string

const (
	ApiService                     KubernetesKind = "apiService"
	ClusterRole                    KubernetesKind = "clusterRole"
	ClusterRoleBinding             KubernetesKind = "clusterRoleBinding"
	ConfigMap                      KubernetesKind = "configMap"
	ControllerRevision             KubernetesKind = "controllerRevision"
	CronJob                        KubernetesKind = "cronJob"
	CustomResourceDefinition       KubernetesKind = "customResourceDefinition"
	DaemonSet                      KubernetesKind = "daemonSet"
	Deployment                     KubernetesKind = "deployment"
	Event                          KubernetesKind = "event"
	Horizontalpodautoscaler        KubernetesKind = "horizontalpodautoscaler"
	Ingress                        KubernetesKind = "ingress"
	Job                            KubernetesKind = "job"
	LimitRange                     KubernetesKind = "limitRange"
	MutatingWebhookConfiguration   KubernetesKind = "mutatingWebhookConfiguration"
	Namespace                      KubernetesKind = "namespace"
	NetworkPolicy                  KubernetesKind = "networkPolicy"
	PersistentVolume               KubernetesKind = "persistentVolume"
	PersistentVolumeClaim          KubernetesKind = "persistentVolumeClaim"
	Pod                            KubernetesKind = "pod"
	PodDisruptionBudget            KubernetesKind = "podDisruptionBudget"
	PodPreset                      KubernetesKind = "podPreset"
	PodSecurityPolicy              KubernetesKind = "podSecurityPolicy"
	ReplicaSet                     KubernetesKind = "replicaSet"
	Role                           KubernetesKind = "role"
	RoleBinding                    KubernetesKind = "roleBinding"
	Secret                         KubernetesKind = "secret"
	Service                        KubernetesKind = "service"
	ServiceAccount                 KubernetesKind = "serviceAccount"
	StatefulSet                    KubernetesKind = "statefulSet"
	StorageClass                   KubernetesKind = "storageClass"
	ValidatingWebhookConfiguration KubernetesKind = "validatingWebhookConfiguration"
)
