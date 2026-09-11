// Package herdrdeploy holds the pure parsing and judgment logic behind
// verify-herdr-deploy. Shelling out to git/nix/ps/lsof stays in main.go;
// this package only interprets their output.
package herdrdeploy

import (
	"encoding/json"
	"fmt"
)

type flakeLockNode struct {
	Inputs map[string]json.RawMessage `json:"inputs"`
	Locked struct {
		Rev string `json:"rev"`
	} `json:"locked"`
}

type flakeLock struct {
	Root  string                   `json:"root"`
	Nodes map[string]flakeLockNode `json:"nodes"`
}

// LockedRev resolves flake.lock's root node's inputs[<inputName>] to a node
// ID and returns that node's locked.rev. It doesn't assume the node ID
// equals the top-level input name: flake.lock can dedupe a shared input
// under a different node ID (this repo's own flake.lock does exactly that
// for "nixpkgs", which resolves to node "nixpkgs_3"), so "herdr" could
// likewise end up under e.g. "herdr_2" after a future dependency change.
func LockedRev(data []byte, inputName string) (string, error) {
	var lock flakeLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return "", fmt.Errorf("parse flake.lock: %w", err)
	}
	rootName := lock.Root
	if rootName == "" {
		rootName = "root"
	}
	rootNode, ok := lock.Nodes[rootName]
	if !ok {
		return "", fmt.Errorf("flake.lock has no root node %q", rootName)
	}
	rawRef, ok := rootNode.Inputs[inputName]
	if !ok {
		return "", fmt.Errorf("flake.lock root node has no input %q", inputName)
	}
	var nodeID string
	if err := json.Unmarshal(rawRef, &nodeID); err != nil {
		return "", fmt.Errorf("flake.lock root input %q isn't a plain node reference (follows chain?): %w", inputName, err)
	}
	node, ok := lock.Nodes[nodeID]
	if !ok {
		return "", fmt.Errorf("flake.lock has no node %q (referenced by root input %q)", nodeID, inputName)
	}
	if node.Locked.Rev == "" {
		return "", fmt.Errorf("flake.lock node %q has no locked.rev", nodeID)
	}
	return node.Locked.Rev, nil
}
