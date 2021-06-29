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

func TestNetworkBandwidth(t *testing.T) {
	// travis rans out of memory when we get a high nodes count.. so we'll skip it for now.
	if testing.Short() {
		t.Skip()
	}

	relayCounts := 8
	nodeCounts := []int{5, 10, 15, 20, 40}

	deadlock.Opts.Disable = true
	rnd := rand.New(rand.NewSource(0))
	k := 4 // outgoing connections
	if k > relayCounts {
		k = relayCounts
	}
	statConf := &TrafficStatisticsFilterConfig{
		OutputFormat: 2,
	}
	for i := range nodeCounts {
		nodeCount := nodeCounts[i]
		t.Run(fmt.Sprintf("TestNetworkBandwidth-%d", nodeCount),
			func(t *testing.T) {
				nodes := nodeCount
				topologyConfig := TopologyFilterConfig{
					NodesConnection: make(map[int][]int),
				}
				for j := 0; j < relayCounts; j++ {
					r := rnd.Perm(relayCounts)
					topologyConfig.NodesConnection[j] = append(topologyConfig.NodesConnection[j], r[:k]...)
					for _, d := range r[:k] {
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
				for i := 0; i < relayCounts+nodes; i++ {
					onlineNodes[i] = (i >= relayCounts)
				}

				netConfig := &FuzzerConfig{
					FuzzerName:  fmt.Sprintf("networkBandwidth-%d", nodes),
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
			})
	}
}
