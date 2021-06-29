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
	"testing"

	"github.com/algorand/go-algorand/logging"
)

func TestCircularNetworkTopology(t *testing.T) {
	var nodeCounts []int
	if testing.Short() {
		nodeCounts = []int{4, 6}
	} else {
		nodeCounts = []int{9, 14}
	}
	for i := range nodeCounts {
		nodeCount := nodeCounts[i]
		t.Run(fmt.Sprintf("TestCircularNetworkTopology-%d", nodeCount),
			func(t *testing.T) {
				nodes := nodeCount
				topologyConfig := TopologyFilterConfig{
					NodesConnection: make(map[int][]int),
				}
				for i := 0; i < nodes; i++ {
					topologyConfig.NodesConnection[i] = []int{(i + nodes - 1) % nodes, (i + 1) % nodes}
				}
				netConfig := &FuzzerConfig{
					FuzzerName: fmt.Sprintf("circularNetworkTopology-%d", nodes),
					NodesCount: nodes,
					Filters:    []NetworkFilterFactory{NetworkFilterFactory(MakeTopologyFilter(topologyConfig))},
					LogLevel:   logging.Info,
				}
				validatorCfg := &ValidatorConfig{
					NetworkRunTicks:     50,
					NetworkRecoverTicks: 20,
				}
				validator := MakeValidator(validatorCfg, t)
				validator.Go(netConfig)
			})
	}
}
