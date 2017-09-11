package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)


func (t *LoyaltyChaincode) removeAsset(stub shim.ChaincodeStubInterface, prefix string, owner string, spender string, id string) error {
	key, _ := stub.CreateCompositeKey(prefix, []string{owner, spender, id})
	return stub.DelState(key)
}

func (t *LoyaltyChaincode) createAsset(stub shim.ChaincodeStubInterface, prefix string, owner string, spender string, history []string, value uint64) (*Asset, error) {
	id, err := t.countExistingAssetsfromSameSource(stub, prefix, owner, spender)
	if err != nil {
		return nil, err
	}

	return t.storeAsset(stub, prefix, owner, spender, uintToString(id), history, value)
}

func (t *LoyaltyChaincode) storeAsset(stub shim.ChaincodeStubInterface, prefix string, owner string, spender string, id string, history []string, value uint64) (*Asset, error) {


	asset := Asset{
		Value: value,
		History: history,
	}

	result, err := json.Marshal(asset)
	if err != nil {
		return nil, err
	}

	key, _ := stub.CreateCompositeKey(prefix, []string{owner, spender, id})
	err = stub.PutState(key, []byte(result))
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (t *LoyaltyChaincode) countExistingAssetsfromSameSource(stub shim.ChaincodeStubInterface, prefix string, owner string, spender string, ) (uint64, error)  {


	iterator, err := stub.GetStateByPartialCompositeKey(prefix, []string{owner, spender})
	if err != nil {
		return 0, err
	}
	defer iterator.Close()
	var num = uint64(0)

	for i := 0; iterator.HasNext(); i++ {
		_ , err = iterator.Next()
		num++
	}

	if err != nil {
		return 0, err
	}

	return num, nil
}