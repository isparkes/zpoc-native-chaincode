package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
)


func (t *LoyaltyChaincode) removeAllowance(stub shim.ChaincodeStubInterface, shop string, spender string) error {
	key, _ := stub.CreateCompositeKey(IndexShopAllowances, []string{shop, spender})
	return stub.DelState(key)
}

func (t *LoyaltyChaincode) createAllowance(stub shim.ChaincodeStubInterface, shop string, spender string, value uint64) (*Allowance, error) {
	key, _ := stub.CreateCompositeKey(IndexShopAllowances, []string{shop, spender})

	allowance := Allowance{
		Spender:spender,
		Value:value,
	}

	data, err := json.Marshal(allowance)
	if err != nil {
		return nil, errors.New("Error creating allowance: " + err.Error())
	}

	err = stub.PutState(key, []byte(data))
	if err != nil {
		return nil, errors.New("Error creating allowance: " + err.Error())
	}

	err = t.updateUserBalance(stub, IndexCustomer, spender, value, true)
	if err != nil {
		return nil, errors.New("Error creating allowance: " + err.Error())
	}

	return &allowance, nil
}