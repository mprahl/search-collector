// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package transforms

import (
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestTransformPolicyReport(t *testing.T) {
	var pr PolicyReport
	UnmarshalFile("policyreport.json", &pr, t)
	node := PolicyReportResourceBuilder(&pr).BuildNode()

	// Test unique fields that exist in policy report and are shown in UI - the common test will test the other bits
	AssertDeepEqual("category Length", len(node.Properties["category"].([]string)), 5, t)
	AssertDeepEqual("rules", node.Properties["rules"], []string{"policyreport testing risk 1 policy", "policyreport testing risk 2 policy"}, t)
	AssertDeepEqual("policies", node.Properties["policies"], []string{"policyreport testing risk 1 policy", "policyreport testing risk 2 policy"}, t)
	AssertDeepEqual("numRuleViolations", node.Properties["numRuleViolations"], 2, t)
	AssertDeepEqual("critical", node.Properties["critical"], 0, t)
	AssertDeepEqual("important", node.Properties["important"], 0, t)
	AssertDeepEqual("moderate", node.Properties["moderate"], 1, t)
	AssertDeepEqual("low", node.Properties["low"], 1, t)

	AssertDeepEqual("scope", node.Properties["scope"], "test-cluster", t)
}

func TestTransformKyvernoClusterPolicyReport(t *testing.T) {
	var pr PolicyReport
	UnmarshalFile("kyverno-clusterpolicyreport.json", &pr, t)
	node := PolicyReportResourceBuilder(&pr).BuildNode()

	AssertDeepEqual("category", node.Properties["category"].([]string), []string{"Kubecost"}, t)
	AssertDeepEqual("policies", node.Properties["policies"], []string{"require-kubecost-labels"}, t)
	// 1 failure and 1 error
	AssertDeepEqual("numRuleViolations", node.Properties["numRuleViolations"], 2, t)
}

func TestTransformKyvernoPolicyReport(t *testing.T) {
	var pr PolicyReport
	UnmarshalFile("kyverno-policyreport.json", &pr, t)
	node := PolicyReportResourceBuilder(&pr).BuildNode()

	AssertDeepEqual("category", node.Properties["category"].([]string), []string{"Kubecost"}, t)
	AssertDeepEqual(
		"policies",
		node.Properties["policies"],
		[]string{"open-cluster-management-agent-addon/require-kubecost-labels", "require-kubecost-labels"},
		t,
	)
	AssertDeepEqual("numRuleViolations", node.Properties["numRuleViolations"], 2, t)
}

func TestKyvernoClusterPolicyReportBuildEdges(t *testing.T) {
	p := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "ClusterPolicy",
			"metadata": map[string]interface{}{
				"name": "require-kubecost-labels",
				"annotations": map[string]interface{}{
					"policies.kyverno.io/severity": "critical",
				},
				"uid": "132ec5b8-892b-40da-8b92-af141c377dfe",
			},
			"spec": map[string]interface{}{
				"validationFailureAction": "deny",
				"random":                  "value",
			},
		},
	}

	nodes := []Node{KyvernoPolicyResourceBuilder(&p).node}
	nodeStore := BuildFakeNodeStore(nodes)

	var pr PolicyReport
	UnmarshalFile("kyverno-clusterpolicyreport.json", &pr, t)

	edges := PolicyReportResourceBuilder(&pr).BuildEdges(nodeStore)

	if len(edges) != 1 {
		t.Fatalf("Expected one edge but got %d", len(edges))
	}

	edge := edges[0]
	expectedEdge := Edge{
		EdgeType:   "reportedBy",
		SourceUID:  "local-cluster/509de4c9-ed73-4309-9764-c88334781eae",
		DestUID:    "local-cluster/132ec5b8-892b-40da-8b92-af141c377dfe",
		SourceKind: "ClusterPolicyReport",
		DestKind:   "ClusterPolicy",
	}
	AssertDeepEqual("edge", edge, expectedEdge, t)
}

func TestKyvernoPolicyReportBuildEdges(t *testing.T) {
	policy := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "Policy",
			"metadata": map[string]interface{}{
				"name":      "require-kubecost-labels",
				"namespace": "open-cluster-management-agent-addon",
				"annotations": map[string]interface{}{
					"policies.kyverno.io/severity": "medium",
				},
				"uid": "132ec5b8-892b-40da-8b92-af141c377dfe",
			},
			"spec": map[string]interface{}{
				"validationFailureAction": "deny",
				"random":                  "value",
			},
		},
	}

	clusterPolicy := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "ClusterPolicy",
			"metadata": map[string]interface{}{
				"name": "require-kubecost-labels",
				"annotations": map[string]interface{}{
					"policies.kyverno.io/severity": "medium",
				},
				"uid": "162ec5b8-892b-40da-8b92-af141c377ddd",
			},
			"spec": map[string]interface{}{
				"validationFailureAction": "deny",
				"random":                  "value",
			},
		},
	}

	nodes := []Node{KyvernoPolicyResourceBuilder(&policy).node, KyvernoPolicyResourceBuilder(&clusterPolicy).node}
	nodeStore := BuildFakeNodeStore(nodes)

	var pr PolicyReport
	UnmarshalFile("kyverno-policyreport.json", &pr, t)

	// Test a PolicyReport generated by a Policy kind
	edges := PolicyReportResourceBuilder(&pr).BuildEdges(nodeStore)

	if len(edges) != 2 {
		t.Fatalf("Expected one edge but got %d", len(edges))
	}

	sort.Slice(edges, func(i, j int) bool { return edges[i].DestUID < edges[j].DestUID })

	edge1 := edges[0]
	expectedEdge := Edge{
		EdgeType:   "reportedBy",
		SourceUID:  "local-cluster/53cd0e2e-34e0-454b-a0c4-e4dbf9306470",
		DestUID:    "local-cluster/132ec5b8-892b-40da-8b92-af141c377dfe",
		SourceKind: "PolicyReport",
		DestKind:   "Policy",
	}
	AssertDeepEqual("edge", edge1, expectedEdge, t)

	edge2 := edges[1]
	expectedEdge2 := Edge{
		EdgeType:   "reportedBy",
		SourceUID:  "local-cluster/53cd0e2e-34e0-454b-a0c4-e4dbf9306470",
		DestUID:    "local-cluster/162ec5b8-892b-40da-8b92-af141c377ddd",
		SourceKind: "PolicyReport",
		DestKind:   "ClusterPolicy",
	}
	AssertDeepEqual("edge", edge2, expectedEdge2, t)
}

func TestPolicyReportBuildEdges(t *testing.T) {
	// Build a fake NodeStore with nodes needed to generate edges.
	nodes := make([]Node, 0)
	nodeStore := BuildFakeNodeStore(nodes)

	// Build edges from mock resource policyreport.json
	var pr PolicyReport
	UnmarshalFile("policyreport.json", &pr, t)
	edges := PolicyReportResourceBuilder(&pr).BuildEdges(nodeStore)

	// Validate results
	AssertEqual("PolicyReport has no edges:", len(edges), 0, t)
}
