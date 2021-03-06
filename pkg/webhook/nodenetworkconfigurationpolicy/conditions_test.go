package nodenetworkconfigurationpolicy

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	nmstatev1alpha1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1alpha1"
)

func expectConditionsUnknown(policy nmstatev1alpha1.NodeNetworkConfigurationPolicy) {
	numberOfConditionTypes := len(nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionTypes)
	ExpectWithOffset(1, policy.Status.Conditions).To(HaveLen(numberOfConditionTypes))
	for _, conditionType := range nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionTypes {
		condition := policy.Status.Conditions.Find(conditionType)
		ExpectWithOffset(1, condition).ToNot(BeNil())
		ExpectWithOffset(1, condition.Status).To(Equal(corev1.ConditionUnknown))
		ExpectWithOffset(1, condition.Reason).To(Equal(nmstatev1alpha1.ConditionReason("")))
		ExpectWithOffset(1, condition.Message).To(Equal(""))
		ExpectWithOffset(1, condition.LastTransitionTime.Time).To(BeTemporally(">", time.Unix(0, 0)))
		ExpectWithOffset(1, condition.LastHeartbeatTime.Time).To(BeTemporally(">", time.Unix(0, 0)))
	}
}

func callHook(hook *webhook.Admission, request webhook.AdmissionRequest) webhook.AdmissionResponse {

	response := hook.Handle(context.TODO(), request)
	for _, patch := range response.Patches {
		_, err := patch.MarshalJSON()
		ExpectWithOffset(2, err).ToNot(HaveOccurred(), "The patches should contain valid JSON")
	}
	ExpectWithOffset(2, response.Allowed).To(BeTrue(), "Mutation of the request should be allowed")
	return response
}

func callDeleteConditions(policy nmstatev1alpha1.NodeNetworkConfigurationPolicy) webhook.AdmissionResponse {
	request := requestForPolicy(policy)
	return callHook(deleteConditionsHook(), request)
}

func callSetConditionsUnknown(policy nmstatev1alpha1.NodeNetworkConfigurationPolicy) webhook.AdmissionResponse {
	request := requestForPolicy(policy)
	return callHook(setConditionsUnknownHook(), request)
}

var _ = Describe("NNCP Conditions Mutating Admission Webhook", func() {
	var (
		obtainedResponse webhook.AdmissionResponse
		policy           = nmstatev1alpha1.NodeNetworkConfigurationPolicy{}
	)
	Context("when setConditionsUnknown is called with nil conditions", func() {
		BeforeEach(func() {
			policy.Status.Conditions = nil
			obtainedResponse = callSetConditionsUnknown(policy)
		})
		It("should have all policy conditions with Unknown state", func() {
			patchedPolicy := patchPolicy(policy, obtainedResponse)
			expectConditionsUnknown(patchedPolicy)
		})

	})
	Context("when setConditionsUnknown is called with empty conditions", func() {
		BeforeEach(func() {
			policy.Status.Conditions = nmstatev1alpha1.ConditionList{}
			obtainedResponse = callSetConditionsUnknown(policy)
		})
		It("should have all policy conditions with Unknown state", func() {
			patchedPolicy := patchPolicy(policy, obtainedResponse)
			expectConditionsUnknown(patchedPolicy)
		})
	})
	Context("when setConditionsUnknown is called with empty conditions", func() {
		BeforeEach(func() {
			policy.Status.Conditions = nmstatev1alpha1.ConditionList{}
			obtainedResponse = callSetConditionsUnknown(policy)
		})
		It("should have all policy conditions with Unknown state", func() {
			patchedPolicy := patchPolicy(policy, obtainedResponse)
			expectConditionsUnknown(patchedPolicy)
		})
	})
	Context("when setConditionsUnknown is called with Some conditions", func() {
		BeforeEach(func() {
			conditions := nmstatev1alpha1.ConditionList{}
			conditions.Set(
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionDegraded,
				corev1.ConditionFalse,
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionSuccessfullyConfigured,
				"",
			)
			conditions.Set(
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionAvailable,
				corev1.ConditionTrue,
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionSuccessfullyConfigured,
				"Foo message",
			)
			policy.Status.Conditions = conditions
			obtainedResponse = callSetConditionsUnknown(policy)
		})
		It("should not change the conditions", func() {
			Expect(obtainedResponse.Patches).To(BeEmpty())
		})

	})
	Context("when deleteConditions is called with empty conditions", func() {
		BeforeEach(func() {
			policy.Status.Conditions = nmstatev1alpha1.ConditionList{}
			obtainedResponse = callDeleteConditions(policy)
		})
		It("should do nothing", func() {
			Expect(obtainedResponse.Patches).To(BeEmpty())
		})
	})
	Context("when deleteConditions is called with some conditions", func() {
		BeforeEach(func() {
			conditions := nmstatev1alpha1.ConditionList{}
			conditions.Set(
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionDegraded,
				corev1.ConditionFalse,
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionSuccessfullyConfigured,
				"",
			)
			conditions.Set(
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionAvailable,
				corev1.ConditionTrue,
				nmstatev1alpha1.NodeNetworkConfigurationPolicyConditionSuccessfullyConfigured,
				"Foo message",
			)
			policy.Status.Conditions = conditions
			obtainedResponse = callDeleteConditions(policy)
		})
		It("should remove all the conditions", func() {
			By(fmt.Sprintf("obtainedResponse: %+v", obtainedResponse))
			patchedPolicy := patchPolicy(policy, obtainedResponse)
			Expect(patchedPolicy.Status.Conditions).To(BeEmpty())
		})
	})

})
