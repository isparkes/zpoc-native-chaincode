package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
)


func (t *LoyaltyChaincode) userExists(stub shim.ChaincodeStubInterface, cn string, role string) bool {
	var prefix string
	switch role {
	case "customer":
		prefix = IndexCustomer
	case "bank":
		prefix = IndexBank
	case "shop":
		prefix = IndexShop
	default:
		return false
	}

	key, _ := stub.CreateCompositeKey(prefix, []string{cn})

	data, err := stub.GetState(key)
	if err != nil {
		return false
	} else if data == nil {
		return false
	}

	return true
}

func (t *LoyaltyChaincode) createUser(stub shim.ChaincodeStubInterface, cn string, role string) error {
	var err error
	switch role {
	case "customer":
		err = t.setInitUserBalance(stub, IndexCustomer, cn, 0)
	case "bank":
		err = t.setInitUserBalance(stub, IndexBank, cn, 0)
	case "shop":
		err = t.setInitUserBalance(stub, IndexShop, cn, 0)
	}

	if err != nil {
		return errors.New("Error creating user '" + cn + "' with the role '" + role + "': " + err.Error())
	}

	return nil
}