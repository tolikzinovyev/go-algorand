package server

import (
	"context"
	"fmt"
	"log"

	"github.com/algorand/go-algorand/daemon/algod/api/server/grpc/encoding"
	"github.com/algorand/go-algorand/daemon/algod/api/server/grpc/proto"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/node"
)

type GrpcServer struct {
	proto.UnimplementedDefaultServer

	node *node.AlgorandFullNode
}

func MakeGrpcServer(node *node.AlgorandFullNode) GrpcServer {
	return GrpcServer{
		node: node,
	}
}

func (s *GrpcServer) AccountInformation(ctx context.Context, in *proto.AccountRequest) (*proto.AccountResponse, error) {
	var address basics.Address
	copy(address[:], in.Address)

	log.Printf("Received: %s", address)

	accountData, round, amountWithoutPendingRewards, err :=
		s.node.Ledger().LookupLatest(address)
	if err != nil {
		return nil, fmt.Errorf("AccountInformation() err: %w", err)
	}

	cparams, err := s.node.Ledger().ConsensusParams(round)
	if err != nil {
		return nil, fmt.Errorf("AccountInformation() err: %w", err)
	}

	res := &proto.AccountResponse{
		AccountData: encoding.ConvertAccountData(&accountData),
		Round: uint64(round),
		AmountWithoutPendingRewards: amountWithoutPendingRewards.Raw,
		MinBalance: accountData.MinBalance(&cparams).Raw,
	}
	return res, nil
}
