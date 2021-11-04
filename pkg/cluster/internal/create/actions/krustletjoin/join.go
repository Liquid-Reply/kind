/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package kubeadmjoin implements the kubeadm join action
package krustletjoin

import (
	"strings"
	"time"

	"sigs.k8s.io/kind/pkg/cluster/constants"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/errors"
	"sigs.k8s.io/kind/pkg/log"

	"sigs.k8s.io/kind/pkg/cluster/nodeutils"

	"sigs.k8s.io/kind/pkg/cluster/internal/create/actions"
	"sigs.k8s.io/kind/pkg/cluster/internal/providers"
	"sigs.k8s.io/kind/pkg/exec"
)

// Action implements action for creating the kubeadm join
// and deploying it on the bootstrap control-plane node.
type Action struct{}

// NewAction returns a new action for creating the kubeadm jion
func NewAction() actions.Action {
	return &Action{}
}

// Execute runs the action
func (a *Action) Execute(ctx *actions.ActionContext) error {
	allNodes, err := ctx.Nodes()
	if err != nil {
		return err
	}

	// then join worker nodes if any
	workers, err := nodeutils.SelectNodesByRole(allNodes, constants.KrustletNodeRoleValue)
	if err != nil {
		return err
	}
	// fetch a controllplane node to approve the csr after krustlet has started
	cpNodes, err := nodeutils.SelectNodesByRole(allNodes, constants.ControlPlaneNodeRoleValue)
	if err != nil {
		return err
	}
	if len(workers) > 0 {
		if err := joinWorkers(ctx, workers, cpNodes[0]); err != nil {
			return err
		}
	}

	return nil
}

func joinWorkers(
	ctx *actions.ActionContext,
	workers []nodes.Node,
	cpNode nodes.Node,
) error {
	ctx.Status.Start("Joining krustlet nodes 🦀")
	defer ctx.Status.End(false)

	// create the workers concurrently
	fns := []func() error{}
	for _, node := range workers {
		node := node // capture loop variable
		fns = append(fns, func() error {
			return runKrustletJoin(ctx.Logger, node, ctx.Provider, ctx.Config.Name, cpNode)
		})
	}
	if err := errors.UntilErrorConcurrent(fns); err != nil {
		return err
	}

	ctx.Status.End(true)
	return nil
}

// runKrustletJoin starts the krustlet and approves the csr
func runKrustletJoin(logger log.Logger, node nodes.Node, provider providers.Provider, name string, cpNode nodes.Node) error {

	cmd := cpNode.Command(
		"curl", "-sSL", "https://raw.githubusercontent.com/krustlet/krustlet/main/scripts/bootstrap.sh",
	)
	script, err := exec.CombinedOutputLines(cmd)
	logger.V(3).Info(strings.Join(script, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to download krustlet bootstrap token script")
	}

	cmd = cpNode.Command(
		"bash", "-c", strings.Join(script, "\n"),
	)
	lines, err := exec.CombinedOutputLines(cmd)
	logger.V(3).Info(strings.Join(lines, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to execute krustlet bootstrap token script")
	}

	cmd = cpNode.Command(
		"cat", "/root/.krustlet/config/bootstrap.conf",
	)
	bootconf, err := exec.CombinedOutputLines(cmd)
	logger.V(3).Info(strings.Join(bootconf, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to fetch bootstrap.conf")
	}

	err = nodeutils.WriteFile(node, "/etc/kubernetes/bootstrap-kubelet.conf", strings.Join(bootconf, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to write kubeconfig")
	}

	cmd = node.Command(
		"systemctl", "enable", "krustlet",
	)
	lines, err = exec.CombinedOutputLines(cmd)
	logger.V(3).Info(strings.Join(lines, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to enable krustlet sevice")
	}

	cmd = node.Command(
		"systemctl", "start", "krustlet",
	)
	lines, err = exec.CombinedOutputLines(cmd)
	logger.V(3).Info(strings.Join(lines, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to run `systemctl start krustlet`")
	}

	for i := 0; i <= 30; i++ {
		err = cpNode.Command(
			"kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf", "get", "csr", node.String()+"-tls",
		).Run()
		if err == nil {
			break
		} else {
			logger.V(2).Info(err.Error())
		}
		time.Sleep(time.Second)
	}

	cmd = cpNode.Command(
		"kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf", "certificate", "approve", node.String()+"-tls",
	)
	lines, err = exec.CombinedOutputLines(cmd)
	logger.V(3).Info(strings.Join(lines, "\n"))
	if err != nil {
		return errors.Wrap(err, "failed to run `systemctl start krustlet`")
	}

	return nil
}
