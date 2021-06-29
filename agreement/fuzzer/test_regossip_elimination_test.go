// Copyright (C) 2019-2021 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package fuzzer

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-deadlock"
)

func pickNodes(totalNodesCount, currentNode, desiredSelectionCount int, rnd *rand.Rand) []int {
	if desiredSelectionCount > totalNodesCount-1 {
		desiredSelectionCount = totalNodesCount - 1
	}

	r := rnd.Perm(totalNodesCount - 1)
	for i := 0; i < desiredSelectionCount; i++ {
		if r[i] >= currentNode {
			r[i]++
		}
	}
	return r[:desiredSelectionCount]
}

func TestRegossipinngElimination(t *testing.T) {
	// travis rans out of memory when we get a high nodes count.. so we'll skip it for now.
	if testing.Short() {
		t.Skip()
	}

	relayCounts := 8
	nodeCount := 20
	deadlock.Opts.Disable = true
	rnd := rand.New(rand.NewSource(0))
	k := 4 // outgoing connections
	if k > relayCounts {
		k = relayCounts
	}
	statConf := &TrafficStatisticsFilterConfig{
		OutputFormat: 2,
	}

	nodes := nodeCount
	topologyConfig := TopologyFilterConfig{
		NodesConnection: make(map[int][]int),
	}
	for j := 0; j < relayCounts; j++ {
		r := pickNodes(relayCounts, j, k, rnd)
		topologyConfig.NodesConnection[j] = append(topologyConfig.NodesConnection[j], r...)
		for _, d := range r {
			topologyConfig.NodesConnection[d] = append(topologyConfig.NodesConnection[d], j)
		}
	}

	for i := relayCounts; i < relayCounts+nodes; i++ {
		r := rnd.Perm(relayCounts)
		topologyConfig.NodesConnection[i] = append(topologyConfig.NodesConnection[i], r[:k]...)
		for _, d := range r[:k] {
			topologyConfig.NodesConnection[d] = append(topologyConfig.NodesConnection[d], i)
		}
	}
	onlineNodes := make([]bool, relayCounts+nodes)
	nodesIndices := make([]int, nodes)
	for i := 0; i < relayCounts+nodes; i++ {
		onlineNodes[i] = (i >= relayCounts)
	}
	for i := 0; i < nodes; i++ {
		nodesIndices[i] = relayCounts + i
	}

	netConfig := &FuzzerConfig{
		FuzzerName:  fmt.Sprintf("networkRegossiping-baseline-%d", nodes),
		NodesCount:  nodes + relayCounts,
		OnlineNodes: onlineNodes,
		Filters: []NetworkFilterFactory{
			MakeTopologyFilter(topologyConfig),
			&DuplicateMessageFilter{},
			MakeTrafficStatisticsFilterFactory(statConf),
		},
		LogLevel:      logging.Error,
		DisableTraces: true,
	}
	validatorCfg := &ValidatorConfig{
		NetworkRunTicks:     150,
		NetworkRecoverTicks: 0,
	}
	validator := MakeValidator(validatorCfg, t)
	validator.Go(netConfig)

	netConfig2 := &FuzzerConfig{
		FuzzerName:  fmt.Sprintf("networkRegossiping-eliminated-regossip-%d", nodes),
		NodesCount:  nodes + relayCounts,
		OnlineNodes: onlineNodes,
		Filters: []NetworkFilterFactory{
			MakeScheduleFilterFactory(&SchedulerFilterConfig{
				Filters: []NetworkFilterFactory{
					&MessageRegossipFilter{},
				},
				Schedule: []SchedulerFilterSchedule{
					{
						Operation: 4, // not before
						FirstTick: 0, // zero
						Nodes:     nodesIndices,
					},
				},
			}),
			MakeTopologyFilter(topologyConfig),
			&DuplicateMessageFilter{},
			MakeTrafficStatisticsFilterFactory(statConf),
		},
		LogLevel:      logging.Error,
		DisableTraces: true,
	}

	validator = MakeValidator(validatorCfg, t)
	validator.Go(netConfig2)
}
