/*
IBM Confidential
OCO Source Materials
5737-E67
(C) Copyright IBM Corporation 2019 All Rights Reserved
The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
*/

package reconciler

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	lru "github.com/golang/groupcache/lru"
	tr "github.ibm.com/IBMPrivateCloud/search-collector/pkg/transforms"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/helm/pkg/proto/hapi/release"
)

type NodeEdge struct {
	BuildNode  []tr.Node
	BuildEdges []func(tr.NodeStore) []tr.Edge
}

func initTestReconciler() *Reconciler {
	return &Reconciler{
		currentNodes:       make(map[string]tr.Node),
		previousNodes:      make(map[string]tr.Node),
		diffNodes:          make(map[string]tr.NodeEvent),
		k8sEventNodes:      make(map[string]tr.NodeEvent),
		previousEventEdges: make(map[string]tr.Edge),
		edgeFuncs:          make(map[string]func(ns tr.NodeStore) []tr.Edge),

		Input:       make(chan tr.NodeEvent),
		purgedNodes: lru.New(CACHE_SIZE),
	}
}

// This function will help to easily verify if any edge, especially if newly added, is part of the Complete payload, rather than just checking the number of edges. Pass in all the built edges, the source and destination UIDs and the edge type.
func verifyEdge(edges []tr.Edge, src string, dest string, edgeType string) bool {
	for _, edge := range edges {
		if edge.SourceUID == src && edge.DestUID == dest && string(edge.EdgeType) == edgeType {
			return true
		}
	}
	return false
}
func createNodeEvents() []tr.NodeEvent {
	events := NodeEdge{}
	nodeEvents := []tr.NodeEvent{}
	//First Node
	unstructuredInput := unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "testowner",
			"metadata": map[string]interface{}{
				"uid": "1234",
			},
		},
	}
	unstructuredNode := tr.UnstructuredResource{Unstructured: &unstructuredInput}.BuildNode()
	bEdges := tr.UnstructuredResource{Unstructured: &unstructuredInput}.BuildEdges
	events.BuildNode = append(events.BuildNode, unstructuredNode)
	events.BuildEdges = append(events.BuildEdges, bEdges)

	// Second Node
	p := v1.Pod{}
	p.APIVersion = "v1"
	p.Name = "testpod"
	p.Namespace = "default"
	p.SelfLink = "/api/v1/namespaces/default/pods/testpod"
	p.UID = "5678"
	podNode := tr.PodResource{Pod: &p}.BuildNode()
	podNode.Metadata["OwnerUID"] = "local-cluster/1234"
	podEdges := tr.PodResource{Pod: &p}.BuildEdges

	events.BuildNode = append(events.BuildNode, podNode)
	events.BuildEdges = append(events.BuildEdges, podEdges)

	//Convert events to node events
	for i := range events.BuildNode {
		ne := tr.NodeEvent{
			Time:         time.Now().Unix(),
			Operation:    tr.Create,
			Node:         events.BuildNode[i],
			ComputeEdges: events.BuildEdges[i],
		}
		nodeEvents = append(nodeEvents, ne)
	}
	return nodeEvents
}

func TestReconcilerOutOfOrderDelete(t *testing.T) {
	s := initTestReconciler()
	ts := time.Now().Unix()

	go func() {
		s.Input <- tr.NodeEvent{
			Time:      ts,
			Operation: tr.Delete,
			Node: tr.Node{
				UID: "test-event",
			},
		}

		s.Input <- tr.NodeEvent{
			Time:      ts - 1000, // insert out of order based off of time
			Operation: tr.Create,
			Node: tr.Node{
				UID: "test-event",
			},
		}
	}()

	// need two calls to drain the queue
	s.reconcileNode()
	s.reconcileNode()

	if _, found := s.currentNodes["test-event"]; found {
		t.Fatal("failed to ignore add event received out of order")
	}

	if _, found := s.purgedNodes.Get("test-event"); !found {
		t.Fatal("failed to added deleted NodeEvent to purgedNodes cache")
	}
}

func TestReconcilerOutOfOrderAdd(t *testing.T) {
	s := initTestReconciler()
	ts := time.Now().Unix()

	go func() {
		s.Input <- tr.NodeEvent{
			Time:      ts,
			Operation: tr.Create,
			Node: tr.Node{
				UID: "test-event",
			},
		}

		s.Input <- tr.NodeEvent{
			Time:      ts - 1000, // insert out of order based off of time
			Operation: tr.Create,
			Node: tr.Node{
				UID: "test-event",
				Properties: map[string]interface{}{
					"staleData": true,
				},
			},
		}
	}()

	// need two calls to drain the queue
	s.reconcileNode()
	s.reconcileNode()

	testNode, ok := s.currentNodes["test-event"]
	if !ok {
		t.Fatal("failed to add test node to current state")
	}

	if _, ok := testNode.Properties["staleData"]; ok {
		t.Fatal("inserted nodes out of order: found stale data")
	}
}

func TestReconcilerAddDelete(t *testing.T) {
	s := initTestReconciler()

	go func() {
		s.Input <- tr.NodeEvent{
			Time:      time.Now().Unix(),
			Operation: tr.Create,
			Node: tr.Node{
				UID: "test-event",
			},
		}
	}()

	s.reconcileNode()

	if _, ok := s.currentNodes["test-event"]; !ok {
		t.Fatal("failed to add test event to current state")
	}
	if _, ok := s.diffNodes["test-event"]; !ok {
		t.Fatal("failed to add test event to diff state")
	}

	go func() {
		s.Input <- tr.NodeEvent{
			Time:      time.Now().Unix(),
			Operation: tr.Delete,
			Node: tr.Node{
				UID: "test-event",
			},
		}
	}()

	s.reconcileNode()

	if _, ok := s.currentNodes["test-event"]; ok {
		t.Fatal("failed to remove test event from current state")
	}
	if _, ok := s.diffNodes["test-event"]; ok {
		t.Fatal("failed to remove test event from diff state")
	}
}

func TestReconcilerRedundant(t *testing.T) {
	s := initTestReconciler()
	s.previousNodes["test-event"] = tr.Node{
		UID: "test-event",
		Properties: map[string]interface{}{
			"very": "important",
		},
	}

	go func() {
		s.Input <- tr.NodeEvent{
			Time:      time.Now().Unix(),
			Operation: tr.Create,
			Node: tr.Node{
				UID: "test-event",
				Properties: map[string]interface{}{
					"very": "important",
				},
			},
		}
	}()

	s.reconcileNode()

	if _, ok := s.diffNodes["test-event"]; ok {
		t.Fatal("failed to ignore redundant add event")
	}
}

func TestReconcilerAddEdges(t *testing.T) {
	testReconciler := initTestReconciler()
	//Add events
	events := createNodeEvents()

	//Input node events to reconciler
	go func() {
		for _, ne := range events {
			testReconciler.Input <- ne
		}
	}()

	for range events {
		testReconciler.reconcileNode()
	}
	//Build edges
	edgeMap1 := testReconciler.allEdges()

	//Expected edge
	edgeMap2 := make(map[string]map[string]tr.Edge, 1)
	edge := tr.Edge{EdgeType: "ownedBy", SourceUID: "local-cluster/5678", DestUID: "local-cluster/1234"}
	edgeMap2["local-cluster/5678"] = map[string]tr.Edge{}
	edgeMap2["local-cluster/5678"]["local-cluster/1234"] = edge

	//Check if the actual and expected edges are the same
	if !reflect.DeepEqual(edgeMap1, edgeMap2) {
		t.Fatal("Expected edges not found")
	} else {
		t.Log("Expected edges found")
	}
}

func TestReconcilerDiff(t *testing.T) {
	testReconciler := initTestReconciler()
	//Add a node to reconciler previous nodes
	testReconciler.previousNodes["local-cluster/1234"] = tr.Node{
		UID: "local-cluster/1234",
		Properties: map[string]interface{}{
			"very": "important",
		},
	}
	//Add events
	events := createNodeEvents()

	//Input node events to reconciler
	go func() {
		for _, ne := range events {
			testReconciler.Input <- ne
		}
	}()

	for range events {
		testReconciler.reconcileNode()
	}
	//Compute reconciler diff - this time there should be 1 node and edge to add, 1 node to update
	diff := testReconciler.Diff()
	//Compute reconciler diff again - this time there shouldn't be any new edges or nodes to add/update
	nextDiff := testReconciler.Diff()

	if (len(diff.AddNodes) != 1 || len(diff.UpdateNodes) != 1 || len(diff.AddEdges) != 1) ||
		(len(nextDiff.AddNodes) != 0 || len(nextDiff.UpdateNodes) != 0 || len(nextDiff.AddEdges) != 0) {
		t.Fatal("Error: Reconciler Diff() not working as expected")
	} else {
		t.Log("Reconciler Diff() working as expected")
	}
}

func TestReconcilerComplete(t *testing.T) {
	input := make(chan *tr.Event)
	output := make(chan tr.NodeEvent)
	ts := time.Now().Unix()
	//Read all files in test-data
	dir := "../../test-data"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	events := make([]tr.Event, 0)
	var appInput unstructured.Unstructured

	//Variables to keep track of helm release object
	var c v1.ConfigMap
	var rls release.Release
	rlsFileCount := 0
	var rlsEvnt = &tr.Event{}

	//Convert to events
	for _, file := range files {
		filePath := dir + "/" + file.Name()
		if file.Name() != "helmrelease_release.json" {
			tr.UnmarshalFile(filePath, &appInput, t)
			appInputLocal := appInput
			var in = &tr.Event{
				Time:      ts,
				Operation: tr.Create,
				Resource:  &appInputLocal,
			}
			// This will process one of the helmrelease files - the helmrelease configmap and store the results
			if file.Name() == "helmrelease_configmap.json" {
				tr.UnmarshalFile(filePath, &c, t)
				rlsFileCount++
				rlsEvnt = in
				continue
			}
			events = append(events, *in)
		} else if file.Name() == "helmrelease_release.json" {
			tr.UnmarshalFile(filePath, &rls, t)
			rlsFileCount++
			continue
		}
	}
	testReconciler := initTestReconciler()
	go tr.TransformRoutine(input, output)

	//Convert events to Node events
	go func() {
		for _, ev := range events {
			localEv := &ev
			input <- localEv
			actual := <-output
			testReconciler.Input <- actual
		}
	}()

	for range events {
		testReconciler.reconcileNode()
	}
	// The rlsFileCount will ensure that both the release configmap and the helm release files are read - so that the release event can be added to reconciler
	if rlsFileCount == 2 {
		releaseTrans := tr.HelmReleaseResource{ConfigMap: &c, Release: &rls}
		go func() {
			testReconciler.Input <- tr.NewNodeEvent(rlsEvnt, releaseTrans, "releases")
		}()
		testReconciler.reconcileNode()
	}
	_, pass := testReconciler.k8sEventNodes["local-cluster/a1140d22-f04b-11e9-ba0f-0016ac10172d"]
	if !pass {
		t.Log(len(testReconciler.k8sEventNodes))
		for k := range testReconciler.k8sEventNodes {
			t.Logf("Event UIDs present : %s", k)
		}
		t.Fatal("Error: Reconciler Missing EventNode")

	}

	// Compute reconciler Complete() state
	com := testReconciler.Complete()
	// Check if edge from AppHelmCR to HelmRelease exists
	if verifyEdge(com.Edges, "local-cluster/fg265feg-d932-22g2-82c2-22345g131h34", "local-cluster/Release/helmrelease-ex", "attachedTo") {
		t.Log("Reconciler Complete() working as expected - expected edge found")
	} else {
		t.Fatal("Error: Reconciler Complete() not working as expected - expected edge local-cluster/fg265feg-d932-22g2-82c2-22345g131h34->'attachedTo'->local-cluster/Release/helmrelease-ex not found")
	}
	// Currently we have 28 nodes and 30 edges. If we change the transform test json's to add more, update the testcase accordingly. This will also help us in testing when we add more nodes/edges
	// We dont create Nodes for kind = Event
	if len(com.Edges) != 30 || com.TotalEdges != 30 || len(com.Nodes) != 28 || com.TotalNodes != 28 {
		t.Fatal("Error: Reconciler Complete() not working as expected")
	} else {
		t.Log("Reconciler Complete() working as expected")
	}
}
