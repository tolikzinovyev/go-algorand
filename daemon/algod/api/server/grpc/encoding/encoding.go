package encoding

import (
	"github.com/algorand/go-algorand/daemon/algod/api/server/grpc/proto"
	"github.com/algorand/go-algorand/data/basics"
)

func convertStatus(status basics.Status) proto.Status {
	switch status {
	case basics.Offline:
		return proto.Status_OFFLINE
	case basics.Online:
		return proto.Status_ONLINE
	case basics.NotParticipating:
		return proto.Status_NOT_PARTICIPATING
	}
	return -1
}

func convertAssetParams(assetParams *basics.AssetParams) *proto.AssetParams {
	return &proto.AssetParams{
		Total: assetParams.Total,
		Decimals: assetParams.Decimals,
		DefaultFrozen: assetParams.DefaultFrozen,
		UnitName: []byte(assetParams.UnitName),
		Url: []byte(assetParams.URL),
		MetadataHash: assetParams.MetadataHash[:],
		Manager: assetParams.Manager[:],
		Reserve: assetParams.Reserve[:],
		Freeze: assetParams.Freeze[:],
		Clawback: assetParams.Clawback[:],
	}
}

func convertAssetParamsMap(m map[basics.AssetIndex]basics.AssetParams) map[uint64]*proto.AssetParams {
	res := make(map[uint64]*proto.AssetParams)
	for k, v := range m {
		res[uint64(k)] = convertAssetParams(&v)
	}
	return res
}

func convertAssetHolding(assetHolding *basics.AssetHolding) *proto.AssetHolding {
	return &proto.AssetHolding{
		Amount: assetHolding.Amount,
		Frozen: assetHolding.Frozen,
	}
}

func convertAssetHoldingMap(m map[basics.AssetIndex]basics.AssetHolding) map[uint64]*proto.AssetHolding {
	res := make(map[uint64]*proto.AssetHolding)
	for k, v := range m {
		res[uint64(k)] = convertAssetHolding(&v)
	}
	return res
}

func convertStateSchema(stateSchema basics.StateSchema) *proto.StateSchema {
	return &proto.StateSchema{
		NumUint: stateSchema.NumUint,
		NumByteSlice: stateSchema.NumByteSlice,
	}
}

func convertTealKeyValue(tkv basics.TealKeyValue) []*proto.TealKeyValue {
	res := make([]*proto.TealKeyValue, 0, len(tkv))

	for k, v := range tkv {
		e := new(proto.TealKeyValue)
		*e = proto.TealKeyValue{
			Key: []byte(k),
		}
		switch v.Type {
		case basics.TealUintType:
			value := new(proto.TealKeyValue_Uint)
			value.Uint = v.Uint
			e.Value = value
		case basics.TealBytesType:
			value := new(proto.TealKeyValue_Bytes)
			value.Bytes = []byte(v.Bytes)
			e.Value = value
		}
	}

	return res
}

func convertAppLocalState(state *basics.AppLocalState) *proto.AppLocalState {
	return &proto.AppLocalState{
		Schema: convertStateSchema(state.Schema),
		KeyValue: convertTealKeyValue(state.KeyValue),
	}
}

func convertAppLocalStateMap(m map[basics.AppIndex]basics.AppLocalState) map[uint64]*proto.AppLocalState {
	res := make(map[uint64]*proto.AppLocalState)
	for k, v := range m {
		res[uint64(k)] = convertAppLocalState(&v)
	}
	return res
}

func convertAppParams(params *basics.AppParams) *proto.AppParams {
	return &proto.AppParams{
		ApprovalProgram: params.ApprovalProgram,
		ClearStateProgram: params.ClearStateProgram,
		GlobalState: convertTealKeyValue(params.GlobalState),
		LocalStateSchema: convertStateSchema(params.StateSchemas.LocalStateSchema),
		GlobalStateSchema: convertStateSchema(params.StateSchemas.GlobalStateSchema),
		ExtraProgramPages: params.ExtraProgramPages,
	}
}

func convertAppParamsMap(m map[basics.AppIndex]basics.AppParams) map[uint64]*proto.AppParams {
	res := make(map[uint64]*proto.AppParams)
	for k, v := range m {
		res[uint64(k)] = convertAppParams(&v)
	}
	return res
}

func ConvertAccountData(accountData *basics.AccountData) *proto.AccountData {
	return &proto.AccountData{
		Status: convertStatus(accountData.Status),
		Microalgos: accountData.MicroAlgos.Raw,
		RewardsBase: accountData.RewardsBase,
		RewardedMicroalgos: accountData.RewardedMicroAlgos.Raw,
		VoteId: accountData.VoteID[:],
		SelectionId: accountData.SelectionID[:],
		StateProofId: accountData.StateProofID[:],
		VoteFirstValid: uint64(accountData.VoteFirstValid),
		VoteLastValid: uint64(accountData.VoteLastValid),
		VoteKeyDilution: accountData.VoteKeyDilution,
		AssetParams: convertAssetParamsMap(accountData.AssetParams),
		AssetHoldings: convertAssetHoldingMap(accountData.Assets),
		AuthAddr: accountData.AuthAddr[:],
		AppLocalStates: convertAppLocalStateMap(accountData.AppLocalStates),
		AppParams: convertAppParamsMap(accountData.AppParams),
		TotalAppSchema: convertStateSchema(accountData.TotalAppSchema),
		TotalExtraAppPages: accountData.TotalExtraAppPages,
	}
}
