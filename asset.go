package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
)


func (t *LoyaltyChaincode) removeAsset(stub shim.ChaincodeStubInterface, prefix string, owner string, spender string, id string) error {
	key, _ := stub.CreateCompositeKey(prefix, []string{owner, spender, id})
	return stub.DelState(key)
}

func (t *LoyaltyChaincode) createAsset(stub shim.ChaincodeStubInterface, prefix string, owner string, spender string, history []string, value uint64) (*Asset, error) {
	var id uint64 = uint64(0)

	// check if key exists already
	for {
		id = uint64Random()
		key, _ := stub.CreateCompositeKey(prefix, []string{owner, spender, uintToString(id)})
		res, err := stub.GetState(key)
		if err != nil {
			return nil, errors.New("Error trying to find an unused key: " + err.Error())
		} else if res == nil {
			break
		}
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
