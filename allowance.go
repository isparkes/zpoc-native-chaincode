package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
)

func (t *LoyaltyChaincode) getAllowance(stub shim.ChaincodeStubInterface, shop string, spender string) (*Allowance, error) {
	key, _ := stub.CreateCompositeKey(IndexShopAllowances, []string{shop, spender})
	data, err := stub.GetState(key)
	if err != nil {
		return nil, errors.New("Error fetching user allowance:" + err.Error())
	} else if data == nil {
		return nil, errors.New("No withdraw allowance for user'" + spender + "' found")
	}

	allowance := Allowance{}
	err = json.Unmarshal(data, &allowance)
	if err != nil {
		return nil, errors.New("Error parsing allowance:" + err.Error())
	}

	return &allowance, nil
}

func (t *LoyaltyChaincode) removeAllowance(stub shim.ChaincodeStubInterface, shop string, spender string) error {
	key, _ := stub.CreateCompositeKey(IndexShopAllowances, []string{shop, spender})
	return stub.DelState(key)
}

func (t *LoyaltyChaincode) createAllowance(stub shim.ChaincodeStubInterface, shop string, spender string, value uint64) (*Allowance, error) {

	userBalance, err := t.userBalance(stub, IndexCustomer, spender)
	if err != nil {
		return nil, err
	} else if userBalance < value {
		return nil, errors.New("User has not enough balance to proceed transaction")
	}

	key, _ := stub.CreateCompositeKey(IndexShopAllowances, []string{shop, spender})

	existingAllowance, err := t.getAllowance(stub, shop, spender)
	if existingAllowance != nil {
		if err != nil {
			return nil, errors.New("Last allowance is still not resolved by the shop: " + err.Error())
		}
	}

	allowance := Allowance{
		Buyer: spender,
		Value: value,
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