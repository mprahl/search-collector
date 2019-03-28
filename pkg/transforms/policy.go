package transforms

import (
	mcm "github.ibm.com/IBMPrivateCloud/hcm-compliance/pkg/apis/policy/v1alpha1"
)

// Takes a *mcm.Policy and yields a Node
func transformPolicy(resource *mcm.Policy) Node {

	policy := transformCommon(resource) // Start off with the common properties

	// Extract the properties specific to this type
	policy.Properties["kind"] = "Policy"
	policy.Properties["remediationAction"] = string(resource.Spec.RemediationAction)
	policy.Properties["compliant"] = string(resource.Status.ComplianceState)
	policy.Properties["valid"] = resource.Status.Valid

	policy.Properties["numRules"] = 0
	if resource.Spec.RoleTemplates != nil {
		rules := 0
		for _, role := range resource.Spec.RoleTemplates {
			if role != nil {
				rules += len(role.Rules)
			}
		}
		policy.Properties["numRules"] = rules
	}

	return policy
}