package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
)

func (t *LoyaltyChaincode) getAllowance(stub shim.ChaincodeStubInterface, prefix string, cn1 string, cn2 string) (*Allowance, error) {

	key, _ := stub.CreateCompositeKey(prefix, []string{cn1, cn2})
	data, err := stub.GetState(key)
	if err != nil {
		return nil, errors.New("Error fetching user allowance:" + err.Error())
	} else if data == nil {
		return nil, errors.New("No allowance for '" + cn1 + "<->" + cn2 + "' found")
	}

	allowance := Allowance{}
	err = json.Unmarshal(data, &allowance)
	if err != nil {
		return nil, errors.New("Error parsing allowance:" + err.Error())
	}

	return &allowance, nil
}

func (t *LoyaltyChaincode) updateAllowance(stub shim.ChaincodeStubInterface, prefix string, cn1 string, cn2 string, delta uint64, negSign bool) (*Allowance, error) {

	allowance, _ := t.getAllowance(stub, prefix, cn1, cn2)

	if allowance != nil {
		if negSign {
			if allowance.Value - delta > allowance.Value {
				return nil, errors.New("value of allowance is to small to proceed transaction")
			}
			allowance.Value = allowance.Value - delta
		} else {
			allowance.Value += delta
		}
	} else {
		if negSign {
			return nil, errors.New("value of allowance is to small to proceed transaction")
		} else {
			allowance = &Allowance{
				Buyer: cn2,
				Value: delta,
			}
		}
	}

	key, _ := stub.CreateCompositeKey(prefix, []string{cn1, cn2})

	data, err := json.Marshal(allowance)
	if err != nil {
		return nil, errors.New("Error creating allowance: " + err.Error())
	}

	err = stub.PutState(key, []byte(data))
	if err != nil {
		return nil, errors.New("Error creating allowance: " + err.Error())
	}

	return allowance, nil
}