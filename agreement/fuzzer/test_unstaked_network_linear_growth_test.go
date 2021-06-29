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
	"time"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-deadlock"
)

func TestUnstakedNetworkLinearGrowth(t *testing.T) {
	// travis rans out of memory when we get a high nodes count.. so we'll skip it for now.
	if testing.Short() {
		t.Skip()
	}

	relayCount := 8
	stakedNodeCount := 4
	deadlock.Opts.Disable = true
	nodeCount := []int{stakedNodeCount, relayCount + stakedNodeCount*2, 3*relayCount + stakedNodeCount*4}

	relayMaxBandwidth := []int{}

	k := 4 // outgoing connections
	if k > relayCount {
		k = relayCount
	}
	statConf := &TrafficStatisticsFilterConfig{
		OutputFormat: 0,
	}
	for i := range nodeCount {
		rnd := rand.New(rand.NewSource(0))
		nodeCount := nodeCount[i]
		nodes := nodeCount
		topologyConfig := TopologyFilterConfig{
			NodesConnection: make(map[int][]int),
		}
		for j := 0; j < relayCount; j++ {
			r := rnd.Perm(relayCount)
			topologyConfig.NodesConnection[j] = append(topologyConfig.NodesConnection[j], r[:k]...)
			for _, d := range r[:k] {
				topologyConfig.NodesConnection[d] = append(topologyConfig.NodesConnection[d], j)
			}
		}

		for i := relayCount; i < relayCount+nodes; i++ {
			r := rnd.Perm(relayCount)
			topologyConfig.NodesConnection[i] = append(topologyConfig.NodesConnection[i], r[:k]...)
			for _, d := range r[:k] {
				topologyConfig.NodesConnection[d] = append(topologyConfig.NodesConnection[d], i)
			}
		}
		// set the last stakedNodeCount accounts to have stake
		onlineNodes := make([]bool, relayCount+nodes)
		for i := 0; i < relayCount+nodes; i++ {
			onlineNodes[i] = false
		}
		for i := relayCount; i < relayCount+stakedNodeCount; i++ {
			onlineNodes[i] = true
		}

		statFilterFactory := MakeTrafficStatisticsFilterFactory(statConf)

		netConfig := &FuzzerConfig{
			FuzzerName:  fmt.Sprintf("networkUnstakedLinearGrowth-%d", nodes),
			NodesCount:  nodes + relayCount,
			OnlineNodes: onlineNodes,
			Filters: []NetworkFilterFactory{
				&DuplicateMessageFilter{},
				MakeTopologyFilter(topologyConfig),
				statFilterFactory,
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

		// we want to figure out what is the highest outgoing data rate on any of the relays.
		highestBytesRate := 0
		for relayIdx := 0; relayIdx < relayCount; relayIdx++ {
			statFilter := statFilterFactory.nodes[relayIdx]
			for _, metrics := range statFilter.tickOutgoingTraffic {
				if metrics.Bytes > highestBytesRate {
					highestBytesRate = metrics.Bytes
				}
			}
		}
		// convert from ticks to seconds unit.
		highestBytesRate = int(time.Duration(highestBytesRate) * time.Second / statFilterFactory.nodes[0].fuzzer.tickGranularity)
		relayMaxBandwidth = append(relayMaxBandwidth, highestBytesRate)
		/*fmt.Printf("Outgoing bytes/sec for %d nodes network (%d non-staked relays, %d non-staked nodes, %d staked nodes) rate %s\n",
		nodes+relayCount,
		relayCount,
		nodes-stakedNodeCount,
		stakedNodeCount,
		ByteCountBinary(highestBytesRate))*/
	}

	// make sure that the max bandwidth is linear at most :
	for i := 1; i < len(relayMaxBandwidth); i++ {
		// the nodes ratio should be 2, but instead of assuming it, calculate it.
		nodesRatio := float32((relayCount + nodeCount[i])) / float32(relayCount+nodeCount[i-1])

		require.Truef(t, int(float32(relayMaxBandwidth[i])/nodesRatio) < relayMaxBandwidth[i-1],
			"Network load with %d nodes was %s. Network load with %d nodes was %s. That's more than a %f ratio.",
			nodeCount[i-1], relayMaxBandwidth[i-1],
			nodeCount[i], relayMaxBandwidth[i],
			nodesRatio)
	}
}
