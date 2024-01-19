package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/ssst0n3/awesome_libs"
	"github.com/ssst0n3/awesome_libs/awesome_error"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

type HungryHelmsman struct {
	Config *rest.Config
	Client *kubernetes.Clientset
}

func (h HungryHelmsman) CreateNamespaceFlagReceiver() (err error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "flag-receiver",
			Labels: map[string]string{
				"pod-security.kubernetes.io/enforce": "restricted",
			},
		},
	}

	namespace, err = h.Client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}

	fmt.Printf("Created Namespace %s\n", namespace.ObjectMeta.Name)
	return
}

func (h HungryHelmsman) CreateNamespaceFlagSender() (err error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "flag-sender",
		},
	}

	namespace, err = h.Client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}

	fmt.Printf("Created Namespace %s\n", namespace.ObjectMeta.Name)
	return
}

func (h HungryHelmsman) CreateServiceAccount() (err error) {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-player",
			Namespace: "default",
		},
	}
	sa, err = h.Client.CoreV1().ServiceAccounts(sa.Namespace).Create(context.TODO(), sa, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created serviceaccount %s\n", sa.ObjectMeta.Name)
	return
}

func (h HungryHelmsman) CreateFlag() (err error) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flag",
			Namespace: "flag-sender",
		},
		Immutable: nil,
		Data: map[string][]byte{
			"flag": []byte("cG90bHVja3trdWJlcm5ldGVzX2Nhbl9iZV9hX2JpdF93ZWlyZH0"),
		},
		Type: "Opaque",
	}
	secret, err = h.Client.CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created secret %s\n", secret.ObjectMeta.Name)
	return
}

func int32Ptr(i int32) *int32 { return &i }

func (h HungryHelmsman) CreateDeploymentFlagSender() (err error) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flag-sender",
			Namespace: "flag-sender",
			Labels: map[string]string{
				"app": "flag-sender",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "flag-sender"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "flag-sender"},
				},
				Spec: corev1.PodSpec{
					Volumes:        nil,
					InitContainers: nil,
					Containers: []corev1.Container{
						{
							Name:  "container",
							Image: "busybox",
							Command: []string{
								"sh",
							},
							Args: []string{
								"-c", "while true; do echo $FLAG | nc 1.1.1.1 80 || continue; echo 'Flag Send'; sleep 10; done",
							},
							Env: []corev1.EnvVar{
								{
									Name: "FLAG",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "flag",
											},
											Key: "flag",
										},
									},
								},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: "default",
					DNSPolicy:          corev1.DNSClusterFirst,
				},
			},
		},
	}
	deployment, err = h.Client.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created deployment %s\n", deployment.ObjectMeta.Name)
	return
}

func (h HungryHelmsman) CreateNetworkPolicy() (err error) {
	networkPolicy := &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flag-receiver",
			Namespace: "flag-receiver",
		},
		Spec: networkingv1.NetworkPolicySpec{
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "flag-sender"},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"ns": "flag-sender"},
							},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				"Ingress", "Egress",
			},
		},
	}
	networkPolicy, err = h.Client.NetworkingV1().NetworkPolicies(networkPolicy.Namespace).Create(context.TODO(), networkPolicy, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created network policy %s\n", networkPolicy.ObjectMeta.Name)
	return
}

func (h HungryHelmsman) SetupRbacAllowCtfPlayerListResources() (err error) {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ctf-player",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"list"},
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
			},
			{
				Verbs:     []string{"list"},
				APIGroups: []string{""},
				Resources: []string{"pods"},
			},
			{
				Verbs:     []string{"list"},
				APIGroups: []string{""},
				Resources: []string{"services"},
			},
		},
	}
	clusterRole, err = h.Client.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created ClusterRole %s\n", clusterRole.ObjectMeta.Name)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ctf-player",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "ctf-player",
				Namespace: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "ctf-player",
		},
	}
	clusterRoleBinding, err = h.Client.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created ClusterRoleBinding %s\n", clusterRole.ObjectMeta.Name)

	return
}

func (h HungryHelmsman) SetupRbacAllowCtfPlayerCreatePodsServices() (err error) {
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-player",
			Namespace: "flag-receiver",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"create", "delete"},
				APIGroups: []string{""},
				Resources: []string{"pods.*"},
			},
			{
				Verbs:     []string{"get", "create", "delete"},
				APIGroups: []string{""},
				Resources: []string{"pods"},
			},
			{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
			},
			{
				Verbs:     []string{"create", "delete"},
				APIGroups: []string{""},
				Resources: []string{"services.*"},
			},
			{
				Verbs:     []string{"get", "create", "delete"},
				APIGroups: []string{""},
				Resources: []string{"services"},
			},
			{
				Verbs:     []string{"list", "get"},
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"networkpolicies"},
			},
		},
	}
	role, err = h.Client.RbacV1().Roles(role.Namespace).Create(context.TODO(), role, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created role %s\n", role.ObjectMeta.Name)

	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-player",
			Namespace: "flag-receiver",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "ctf-player",
				Namespace: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "ctf-player",
		},
	}
	roleBinding, err = h.Client.RbacV1().RoleBindings(roleBinding.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created RoleBinding %s\n", roleBinding.ObjectMeta.Name)

	return
}

func (h HungryHelmsman) SetupRbacAllowCtfPlayerGetPodLog() (err error) {
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-player",
			Namespace: "flag-sender",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
			},
		},
	}
	role, err = h.Client.RbacV1().Roles(role.Namespace).Create(context.TODO(), role, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created role %s\n", role.ObjectMeta.Name)

	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-player",
			Namespace: "flag-sender",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "ctf-player",
				Namespace: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "ctf-player",
		},
	}
	roleBinding, err = h.Client.RbacV1().RoleBindings(roleBinding.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created RoleBinding %s\n", roleBinding.ObjectMeta.Name)

	return
}

func (h HungryHelmsman) PrintPlayerConfig() (err error) {
	tokenSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-player-token",
			Namespace: metav1.NamespaceDefault,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": "ctf-player",
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	tokenSecret, err = h.Client.CoreV1().Secrets(tokenSecret.Namespace).Create(context.TODO(), tokenSecret, metav1.CreateOptions{})
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	fmt.Printf("Created token %s\n", tokenSecret.ObjectMeta.Name)

	var token string
	for {
		secret, err := h.Client.CoreV1().Secrets(tokenSecret.Namespace).Get(context.TODO(), tokenSecret.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			awesome_error.CheckErr(err)
			time.Sleep(time.Second)
			continue
		}
		data, ok := secret.Data["token"]
		if !ok {
			time.Sleep(time.Second)
			continue
		}
		token = string(data)
		break
	}
	template := `apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: {.server}
  name: ctf-cluster
contexts:
- context:
    cluster: ctf-cluster
    user: ctf-player
  name: ctf-cluster
current-context: ctf-cluster
kind: Config
users:
- name: ctf-player
  user:
    token: {.token}
EOF`
	spew.Dump(h.Config.Host)
	fmt.Println(awesome_libs.Format(template, awesome_libs.Dict{
		"server": h.Config.Host,
		"token":  token,
	}))
	return
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	h := HungryHelmsman{Config: config, Client: clientset}
	awesome_error.CheckFatal(h.CreateNamespaceFlagReceiver())
	awesome_error.CheckFatal(h.CreateNamespaceFlagSender())
	awesome_error.CheckFatal(h.CreateServiceAccount())
	awesome_error.CheckFatal(h.CreateFlag())
	awesome_error.CheckFatal(h.CreateDeploymentFlagSender())
	awesome_error.CheckFatal(h.CreateNetworkPolicy())
	awesome_error.CheckFatal(h.SetupRbacAllowCtfPlayerListResources())
	awesome_error.CheckFatal(h.SetupRbacAllowCtfPlayerCreatePodsServices())
	awesome_error.CheckFatal(h.SetupRbacAllowCtfPlayerGetPodLog())
	awesome_error.CheckFatal(h.PrintPlayerConfig())
}
