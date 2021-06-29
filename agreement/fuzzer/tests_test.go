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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	//ossignal "os/signal"
	//"runtime/pprof"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-deadlock"
	//"github.com/algorand/go-algorand/protocol"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		log.Println("Skipping fuzzer tests in short mode.")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

/*func setupBreakDump() {
	c := make(chan os.Signal, 1)
	ossignal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		os.Exit(1)
	}()
}

func printResults(t *testing.T, r *RunResult) {
	if r.NetworkStalled {
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		require.Failf(t, "No network activitry detected", "Network has stalled.")
		os.Exit(1)
		return
	}
	require.Truef(t, (r.PostRecoveryHighRound-r.PostRecoveryLowRound <= 1),
		"Initial Rounds %d-%d\nPre Recovery Rounds %d-%d\nPost Recovery Rounds %d-%d",
		r.StartLowRound, r.StartHighRound,
		r.PreRecoveryLowRound, r.PreRecoveryHighRound,
		r.PostRecoveryLowRound, r.PostRecoveryHighRound,
	)
	require.NotEqualf(t, int(r.PreRecoveryHighRound), int(r.PostRecoveryHighRound),
		"Initial Rounds %d-%d\nPre Recovery Rounds %d-%d\nPost Recovery Rounds %d-%d",
		r.StartLowRound, r.StartHighRound,
		r.PreRecoveryLowRound, r.PreRecoveryHighRound,
		r.PostRecoveryLowRound, r.PostRecoveryHighRound,
	)
	if r.PreRecoveryHighRound != r.PreRecoveryLowRound {
		// network got disputed by the filters.
		fmt.Printf("%v partitioned the network ( %d - %d ), but recovered correctly reaching round %d\n", t.Name(), r.PreRecoveryLowRound, r.PreRecoveryHighRound, r.PostRecoveryHighRound)
	} else {
		if r.PreRecoveryHighRound == r.StartLowRound {
			fmt.Printf("%v stalled the network, and the network reached round %d\n", t.Name(), r.PostRecoveryHighRound)
		} else {
			fmt.Printf("%v did not partition the network, and the network reached round %d\n", t.Name(), r.PostRecoveryHighRound)
		}
	}
}*/

/*
func testConfig(t *testing.T, config NetworkConfig) (network *Network) {
	network = MakeNetwork(config)
	require.NotNil(t, network)
	network.Start()
	_, runRes := network.Run(time.Millisecond*5000, time.Millisecond*3000, time.Second)
	if !runRes.NetworkStalled {
		network.Shutdown()
	}
	printResults(t, runRes)
	return
}
*/

/*
func TestManyBandwidthFilter(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("TestBandwidthFilter", TestBandwidthFilter)
	}
}

func TestBandwidthFilter(t *testing.T) {
	setupBreakDump()

	upstreamBandwidth := map[int]int{0: 1000, 1: 500, 2: 1000, 3: 4000, 4: 4000}
	downstreamBandwidth := map[int]int{0: 2000, 1: 2000, 2: 2000, 3: 2000, 4: 2000}

	config := NetworkConfig{
		NetworkName: "LimitBandwidth",
		NodesCount:  5,
		Filters:     []NetworkFilterFactory{NetworkFilterFactory(MakeBandwidthFilter(upstreamBandwidth, downstreamBandwidth))},
		LogLevel:    logging.Debug,
	}
	testConfig(t, config)
}

func TestMessageReordering(t *testing.T) {
	nodeCounts := []int{5, 8, 13}
	sendShuffles := []int{0, 3, 5}
	receiveShuffles := []int{0, 3, 5}

	for n := range nodeCounts {
		nodeCount := nodeCounts[n]
		for s := range sendShuffles {
			sendShuffle := sendShuffles[s]
			for r := range receiveShuffles {
				receiveShuffle := receiveShuffles[r]
				t.Run(fmt.Sprintf("TestMessagerReordering-%v-%v-%v", nodeCount, sendShuffle, receiveShuffle),
					func(t *testing.T) {
						nodes := nodeCount
						send := sendShuffle
						receive := receiveShuffle

						reorderingConfig := MessageReorderingFilterConfig{
							NodesShuffleConfig: make(map[int]NodeShuffleConfig),
						}
						reorderingConfig.NodesShuffleConfig[0] = NodeShuffleConfig{
							SendShuffleSize:    send,
							ReceiveShuffleSize: receive,
							MaxRetension:       time.Second,
						}

						config := NetworkConfig{
							NetworkName: fmt.Sprintf("messageReordering--%v-%v-%v", nodes, send, receive),
							NodesCount:  nodes,
							Filters:     []NetworkFilterFactory{NetworkFilterFactory(MakeMessageReorderingFilter(reorderingConfig))},
							LogLevel:    logging.Error,
						}
						testConfig(t, config)
					})
			}
		}

	}
}

func TestManyMessageDelayFilter(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("TestMessageDelayFilter", TestMessageDelayFilter)
	}
}

func TestMessageDelayFilter(t *testing.T) {
	setupBreakDump()

	bDelay := map[int]map[string]int{0: {"*": 12}, 1: {"*": 2}, 2: {"*": 9}, 3: {"*": 6}, 4: {"*": 5}}
	aDelay := map[int]map[string]int{0: {"AV": 30, "VV": 40, "*": 4}, 1: {"*": 4}, 2: {"*": 9}, 3: {"*": 6}, 4: {"*": 5}}

	config := NetworkConfig{
		NetworkName: "MessageDelay",
		NodesCount:  5,
		Filters:     []NetworkFilterFactory{NetworkFilterFactory(MakeMessageDelayFilter(aDelay, bDelay)), NetworkFilterFactory(MakeMessageDelayFilter(bDelay, aDelay))},
		LogLevel:    logging.Debug,
	}
	testConfig(t, config)
}

func TestMessageDecoderFilter(t *testing.T) {
	setupBreakDump()

	msgDecoder := &MessageDecoderFilter{}
	config := NetworkConfig{
		NetworkName: "messageDecoder",
		NodesCount:  5,
		Filters:     []NetworkFilterFactory{NetworkFilterFactory(msgDecoder)},
		LogLevel:    logging.Debug,
	}
	testConfig(t, config)

	require.Truef(t, msgDecoder.getDecodedMessageCounts(protocol.AgreementVoteTag) != 0,
		"Decoded vote messages %v", msgDecoder.getDecodedMessageCounts(protocol.AgreementVoteTag))
	require.Truef(t, msgDecoder.getDecodedMessageCounts(protocol.ProposalPayloadTag) != 0,
		"Decoded proposals messages %v", msgDecoder.getDecodedMessageCounts(protocol.ProposalPayloadTag))
	require.Truef(t, msgDecoder.getDecodedMessageCounts(protocol.VoteBundleTag) == 0,
		"Decoded bundles messages %v", msgDecoder.getDecodedMessageCounts(protocol.VoteBundleTag))
}

func TestManyMessageDuplicationFilter(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("TestMessageDuplicationFilter", TestMessageDuplicationFilter)
	}
}
func TestMessageDuplicationFilter(t *testing.T) {
	setupBreakDump()

	aDuplication := map[int]map[string]int{0: {"*": 4}, 1: {"*": 4}, 2: {"*": 10}, 3: {"AV": 4}, 4: {"VV": 4}}
	bDuplication := map[int]map[string]int{0: {"*": 12}, 1: {"*": 2}, 2: {"*": 9}, 3: {"*": 6}, 4: {"*": 5}}
	cDuplication := map[int]map[string]int{0: {"AV": 4, "VV": 2, "*": 1}, 1: {"*": 4}, 2: {"*": 9}, 3: {"*": 6}, 4: {"*": 5}}

	config := NetworkConfig{
		NetworkName: "MessageDelay",
		NodesCount:  5,
		Filters:     []NetworkFilterFactory{NetworkFilterFactory(MakeMessageDuplicationFilter(aDuplication, bDuplication)), NetworkFilterFactory(MakeMessageDuplicationFilter(bDuplication, cDuplication))},
		LogLevel:    logging.Debug,
	}
	testConfig(t, config)
}

func TestManyMessageReflectionFilter(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("TestMessageReflectionFilter", TestMessageReflectionFilter)
	}
}

func TestMessageReflectionFilter(t *testing.T) {
	setupBreakDump()

	aReflection := map[int]map[string]int{0: {"VV": 4}, 1: {"VV": 4}, 2: {"VV": 10}, 3: {"VV": 40}, 4: {"VV": 4}}
	bReflection := map[int]map[string]int{0: {"PP": 12}, 1: {"PP": 2}, 2: {"PP": 9}, 3: {"PP": 6}, 4: {"PP": 5}}
	cReflection := map[int]map[string]int{0: {"AV": 1}, 1: {"AV": 4}, 2: {"AV": 9}, 3: {"AV": 6}, 4: {"AV": 5}}
	dReflection := map[int]map[string]int{0: {"VB": 1}, 1: {"VB": 40}, 2: {"VB": 9}, 3: {"VB": 6}, 4: {"VB": 5}}

	config := NetworkConfig{
		NetworkName: "MessageReflection",
		NodesCount:  5,
		Filters:     []NetworkFilterFactory{NetworkFilterFactory(MakeMessageReflectionFilter(aReflection, bReflection)), NetworkFilterFactory(MakeMessageReflectionFilter(cReflection, dReflection))},
		LogLevel:    logging.Debug,
	}
	testConfig(t, config)
}

func TestCertVoteDrops(t *testing.T) {
	topologyConfig := TopologyFilterConfig{
		NodesConnection: map[int][]int{
			0: []int{},
			1: []int{2, 3, 4},
			2: []int{1, 3, 4},
			3: []int{1, 2, 4},
			4: []int{1, 2, 3},
		},
	}
	voteFilterConfig := &VoteFilterConfig{
		IncludeMasks: []VoteFilterMask{
			VoteFilterMask{
				StartRound:  300,
				EndRound:    350,
				StartPeriod: 0,
				EndPeriod:   5000,
				StartStep:   propose,
				EndStep:     255,
			},
		},
		ExcludeMasks: []VoteFilterMask{
			VoteFilterMask{
				StartRound:  300,
				EndRound:    350,
				StartPeriod: 0,
				EndPeriod:   5000,
				StartStep:   cert,
				EndStep:     cert,
			},
		},
	}
	//type voteFilterConfigJSON struct {
	//	Name string
	//	VoteFilterConfig  VoteFilterConfig
	//}
	//tempVoteConfig := & voteFilterConfigJSON  {
	//	Name : "VoteFilterJSON",
	//	VoteFilterConfig : *voteFilterConfig,
	//}
	//tempFilterBytes, _ := json.Marshal(tempVoteConfig  )
	//fmt.Printf("json \n %s \n", string(tempFilterBytes[:]) )

	catchupFilterConfig := &CatchupFilterConfig{
		Nodes: []int{0},
		Count: 1,
	}
	schedConfig1 := &SchedulerFilterConfig{
		Filters: []NetworkFilterFactory{
			MakeTopologyFilter(topologyConfig),
		},
		Schedule: []SchedulerFilterSchedule{
			SchedulerFilterSchedule{
				FirstDuration: 500 * time.Millisecond,
				Operation:     Before,
				Nodes:         []int{0, 1, 2, 3, 4},
			},
		},
		ScheduleName: "Disconnect",
	}
	schedConfig2 := &SchedulerFilterConfig{
		Filters: []NetworkFilterFactory{
			MakeTopologyFilter(topologyConfig),
			MakeVoteFilter(voteFilterConfig),
		},
		Schedule: []SchedulerFilterSchedule{
			SchedulerFilterSchedule{
				FirstDuration:  500 * time.Millisecond,
				SecondDuration: 10000 * time.Millisecond,
				Operation:      Between,
				Nodes:          []int{0, 1, 2, 3, 4},
			},
		},
		ScheduleName: "Disconnect+Filter Certs",
	}
	schedConfig3 := &SchedulerFilterConfig{
		Filters: []NetworkFilterFactory{
			MakeTopologyFilter(topologyConfig),
			MakeCatchupFilterFactory(catchupFilterConfig),
		},
		Schedule: []SchedulerFilterSchedule{
			SchedulerFilterSchedule{
				FirstDuration:  10000 * time.Millisecond,
				SecondDuration: 11000 * time.Millisecond,
				Operation:      Between,
				Nodes:          []int{0},
			},
		},
		ScheduleName: "Catchup Node 0",
	}
	schedConfig4 := &SchedulerFilterConfig{
		Filters: []NetworkFilterFactory{
			MakeVoteFilter(voteFilterConfig),
		},
		Schedule: []SchedulerFilterSchedule{
			SchedulerFilterSchedule{
				FirstDuration:  11000 * time.Millisecond,
				SecondDuration: 12500 * time.Millisecond,
				Operation:      Between,
				Nodes:          []int{0, 1, 2, 3, 4},
			},
		},
		ScheduleName: "Filter Certs",
	}
	msgDecoder := &MessageDecoderFilter{}
	config := NetworkConfig{
		NetworkName: "schedulerTest",
		NodesCount:  5,
		Filters: []NetworkFilterFactory{
			NetworkFilterFactory(msgDecoder),
			NetworkFilterFactory(MakeScheduleFilterFactory(schedConfig1)),
			NetworkFilterFactory(MakeScheduleFilterFactory(schedConfig2)),
			NetworkFilterFactory(MakeScheduleFilterFactory(schedConfig3)),
			NetworkFilterFactory(MakeScheduleFilterFactory(schedConfig4))},
		LogLevel: logging.Debug,
	}

	network := MakeNetwork(config)
	require.NotNil(t, network)
	network.Start()
	_, runRes := network.Run(time.Millisecond*2000, time.Millisecond*1000, time.Second)
	if !runRes.NetworkStalled {
		network.Shutdown()
	}
	printResults(t, runRes)
}*/

func BenchmarkNetworkPerformance(b *testing.B) {
	relayCount := 8
	stakedNodeCount := 40
	nodeCount := []int{stakedNodeCount}
	relayMaxBandwidth := []int{}

	// report memory allocations
	b.ReportAllocs()

	// disable deadlock checking code
	deadlock.Opts.Disable = true

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

		validator := MakeValidator(validatorCfg, b)

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

		require.Truef(b, int(float32(relayMaxBandwidth[i])/nodesRatio) < relayMaxBandwidth[i-1],
			"Network load with %d nodes was %s. Network load with %d nodes was %s. That's more than a %f ratio.",
			nodeCount[i-1], relayMaxBandwidth[i-1],
			nodeCount[i], relayMaxBandwidth[i],
			nodesRatio)
	}
}
