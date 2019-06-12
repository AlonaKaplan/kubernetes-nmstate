package e2e

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	yaml "sigs.k8s.io/yaml"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"

	nmstatev1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1"
)

func writePodsLogs(namespace string, writer io.Writer) error {
	if framework.Global.LocalOperator {
		return nil
	}
	podLogOpts := corev1.PodLogOptions{}
	podList := &corev1.PodList{}
	err := framework.Global.Client.List(context.TODO(), &dynclient.ListOptions{}, podList)
	Expect(err).ToNot(HaveOccurred())
	podsClientset := framework.Global.KubeClient.CoreV1().Pods(namespace)

	for _, pod := range podList.Items {
		req := podsClientset.GetLogs(pod.Name, &podLogOpts)
		podLogs, err := req.Stream()
		if err != nil {
			io.WriteString(writer, fmt.Sprintf("error in opening stream: %v\n", err))
			continue
		}
		defer podLogs.Close()
		_, err = io.Copy(writer, podLogs)
		if err != nil {
			io.WriteString(writer, fmt.Sprintf("error in copy information from podLogs to buf: %v\n", err))
			continue
		}

	}
	return nil
}

func prepare(t *testing.T) (*framework.TestCtx, string) {
	By("Initialize cluster resources")
	cleanupRetryInterval := time.Second * 1
	cleanupTimeout := time.Second * 5
	ctx := framework.NewTestCtx(t)
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	Expect(err).ToNot(HaveOccurred())

	// get namespace
	namespace, err := ctx.GetNamespace()
	Expect(err).ToNot(HaveOccurred())

	err = e2eutil.WaitForOperatorDeployment(t, framework.Global.KubeClient, namespace, "nmstate-manager", 1, time.Second*5, time.Second*30)
	Expect(err).ToNot(HaveOccurred())
	return ctx, namespace
}

func readStateFromFile(file string, namespace string) nmstatev1.NodeNetworkState {
	manifest, err := ioutil.ReadFile(file)
	Expect(err).ToNot(HaveOccurred())

	state := nmstatev1.NodeNetworkState{}
	err = yaml.Unmarshal(manifest, &state)
	Expect(err).ToNot(HaveOccurred())
	state.ObjectMeta.Namespace = namespace
	return state
}

func createStateFromFile(file string, namespace string, cleanupOptions *framework.CleanupOptions) {
	state := readStateFromFile(file, namespace)
	err := framework.Global.Client.Create(context.TODO(), &state, cleanupOptions)
	Expect(err).ToNot(HaveOccurred())
}

func updateStateSpecFromFile(file string, key types.NamespacedName) {
	state := nmstatev1.NodeNetworkState{}
	stateFromManifest := readStateFromFile(file, key.Namespace)
	err := framework.Global.Client.Get(context.TODO(), key, &state)
	Expect(err).ToNot(HaveOccurred())
	state.Spec = stateFromManifest.Spec
	err = framework.Global.Client.Update(context.TODO(), &state)
	Expect(err).ToNot(HaveOccurred())
}

func nodeNetworkState(key types.NamespacedName) nmstatev1.NodeNetworkState {
	state := nmstatev1.NodeNetworkState{}
	timeout := 5 * time.Second
	interval := 1 * time.Second
	Eventually(func() error {
		return framework.Global.Client.Get(context.TODO(), key, &state)
	}, timeout, interval).ShouldNot(HaveOccurred())
	return state
}