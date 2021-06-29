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
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-deadlock"
)

func calcQuadricCoefficients(x []float32, y []float32) (a, b, c float32) {
	a = y[0]/((x[0]-x[1])*(x[0]-x[2])) + y[1]/((x[1]-x[0])*(x[1]-x[2])) + y[2]/((x[2]-x[0])*(x[2]-x[1]))
	b = -y[0]*(x[1]+x[2])/((x[0]-x[1])*(x[0]-x[2])) - y[1]*(x[0]+x[2])/((x[1]-x[0])*(x[1]-x[2])) - y[2]*(x[0]+x[1])/((x[2]-x[0])*(x[2]-x[1]))
	c = y[0]*x[1]*x[2]/((x[0]-x[1])*(x[0]-x[2])) + y[1]*x[0]*x[2]/((x[1]-x[0])*(x[1]-x[2])) + y[2]*x[0]*x[1]/((x[2]-x[0])*(x[2]-x[1]))
	return
}

func TestStakedNetworkQuadricGrowth(t *testing.T) {
	// travis rans out of memory when we get a high nodes count.. so we'll skip it for now.
	if testing.Short() {
		t.Skip()
	}

	relayCount := 1
	nodeCount := []int{4, 5, 6, 7, 8, 9, 10}
	totalRelayedMessages := []int{}
	deadlock.Opts.Disable = true

	k := 2 // outgoing connections
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

		onlineNodes := make([]bool, relayCount+nodes)
		for i := 0; i < relayCount; i++ {
			onlineNodes[i] = false
		}
		for i := relayCount; i < relayCount+nodeCount; i++ {
			onlineNodes[i] = true
		}

		statFilterFactory := MakeTrafficStatisticsFilterFactory(statConf)

		netConfig := &FuzzerConfig{
			FuzzerName:  fmt.Sprintf("stakedNetworkQuadricGrowth-%d", nodes),
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

		relayStatFilter := statFilterFactory.nodes[0]
		totalRelayedMessages = append(totalRelayedMessages, relayStatFilter.totalSentMessage.Count)

		/*fmt.Printf("total relayed messaged for %d nodes is %d\n",
		nodes,
		totalRelayedMessages[len(totalRelayedMessages)-1])*/
	}

	// try to predict the outcome and calculate the diff
	for i := 3; i < len(totalRelayedMessages); i++ {
		// calculate the quadric formula y=ax^2+bx+c coefficients :
		ys := []float32{float32(totalRelayedMessages[0]), float32(totalRelayedMessages[1]), float32(totalRelayedMessages[i-1])}
		xs := []float32{float32(nodeCount[0]), float32(nodeCount[1]), float32(nodeCount[i-1])}
		a, b, c := calcQuadricCoefficients(xs, ys)

		x := float32(nodeCount[i])
		y := float64(a*x*x + b*x + c)
		//fmt.Printf("actual value for %d nodes is %v; expected %v\n", int(x), totalRelayedMessages[i], int(y))
		diff := math.Abs(y - float64(totalRelayedMessages[i]))
		require.Truef(t, diff < float64(totalRelayedMessages[i]),
			"Non quadric growth of network messages found. nodes count = %d, expected message count = %d, actual message count = %d",
			nodeCount[i],
			int(y),
			totalRelayedMessages[i])
	}
}
