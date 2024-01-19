package main

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
)

var (
	h = HungryHelmsman{}
)

func init() {
	config, err := clientcmd.BuildConfigFromFlags("", "test/test_data/admin")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	h = HungryHelmsman{Config: config, Client: clientset}

}

func TestHungryHelmsman_CreateNamespaceFlagReceiver(t *testing.T) {
	assert.NoError(t, h.CreateNamespaceFlagReceiver())
}

func TestHungryHelmsman_CreateNamespaceFlagSender(t *testing.T) {
	assert.NoError(t, h.CreateNamespaceFlagSender())
}

func TestHungryHelmsman_CreateServiceAccount(t *testing.T) {
	assert.NoError(t, h.CreateServiceAccount())
}

func TestHungryHelmsman_CreateFlag(t *testing.T) {
	assert.NoError(t, h.CreateFlag())
}

func TestHungryHelmsman_CreateDeploymentFlagSender(t *testing.T) {
	assert.NoError(t, h.CreateDeploymentFlagSender())
}

func TestHungryHelmsman_CreateNetworkPolicy(t *testing.T) {
	assert.NoError(t, h.CreateNetworkPolicy())
}

func TestHungryHelmsman_SetupRbacAllowCtfPlayerListResources(t *testing.T) {
	assert.NoError(t, h.SetupRbacAllowCtfPlayerListResources())
}

func TestHungryHelmsman_SetupRbacAllowCtfPlayerCreatePodsServices(t *testing.T) {
	assert.NoError(t, h.SetupRbacAllowCtfPlayerCreatePodsServices())
}

func TestHungryHelmsman_SetupRbacAllowCtfPlayerGetPodLog(t *testing.T) {
	assert.NoError(t, h.SetupRbacAllowCtfPlayerGetPodLog())
}

func TestHungryHelmsman_PrintPlayerConfig(t *testing.T) {
	assert.NoError(t, h.PrintPlayerConfig())
}
