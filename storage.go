package main

import (
	"encoding/binary"
	"errors"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func (t *LoyaltyChaincode) setInitUserBalance(stub shim.ChaincodeStubInterface, prefix string, cn string, balance uint64) error {

	if balance < 0 {
		return errors.New("balance can't be negative")
	}

	key, _ := stub.CreateCompositeKey(prefix, []string{cn})
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, balance)
	return stub.PutState(key, data)
}

func (t *LoyaltyChaincode) udpateUserBalance(stub shim.ChaincodeStubInterface, prefix string, cn string, delta uint64, negSign bool) error {

	key, _ := stub.CreateCompositeKey(prefix, []string{cn})
	data, err := stub.GetState(key)
	if err != nil {
		return err
	} else if data == nil {
		return errors.New("User '" + cn + "' doesn't exist")
	}

	newBalance := binary.LittleEndian.Uint64(data)

	if negSign {
		if newBalance - delta < 0 {
			return errors.New("balance can't be negative")
		}
		newBalance -= delta
	} else {
		newBalance += delta
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, newBalance)
	return stub.PutState(key, data)
}

func (t *LoyaltyChaincode) userBalance(stub shim.ChaincodeStubInterface, cn string) (uint64, error) {
	key, _ := stub.CreateCompositeKey(IndexCustomer, []string{cn})
	data, err := stub.GetState(key)
	if err != nil {
		return 0, err
	}

	// if the user cn is not in the state, then the userBalance is 0
	if data == nil {
		return 0, errors.New("User '" + cn + "' doesn't exist!")
	}

	return binary.LittleEndian.Uint64(data), nil
}

func (t *LoyaltyChaincode) makeGiftToTheUserAsBank(stub shim.ChaincodeStubInterface, bankCn string, userCn string, balance uint64) error {

	if balance < 0 {
		return errors.New("gift to the user can't be negative")
	}

	_, err := t.createAsset(stub, IndexCustomerAsset, userCn, bankCn, []string{bankCn}, balance)
	if err != nil {
		return errors.New("Could not create Asset for '" + userCn + "':" + err.Error())
	}

	key, _ := stub.CreateCompositeKey(IndexBanksCustomers, []string{bankCn, userCn})
	newBallance := balance
	data, err := stub.GetState(key)
	if err != nil {
		return err
	}

	if data != nil {
		newBallance += binary.LittleEndian.Uint64(data)
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, newBallance)

	err = stub.PutState(key, data)
	if err != nil {
		return err
	}


	return t.udpateUserBalance(stub, IndexCustomer, userCn, balance, false)
}


func (t *LoyaltyChaincode) getBanksCustomers(stub shim.ChaincodeStubInterface, bankCn string) ([]*User, error) {

	iterator, err := stub.GetStateByPartialCompositeKey(IndexBanksCustomers, []string{bankCn})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var result []*User = []*User{}
	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return nil, err
		}

		_, parts, err := stub.SplitCompositeKey(kv.Key)
		cName := parts[1]
		cBalance := kv.Value

		customer := User{
			Name: cName,
			Balance:  binary.LittleEndian.Uint64(cBalance),
		}

		result = append(result, &customer)
	}

	return result, nil
}

func (t *LoyaltyChaincode) userToUserTransfer(stub shim.ChaincodeStubInterface, fromCn string, toCn string, trValue uint64) error {

	if trValue < 0 {
		return errors.New("transfer can't be negative")
	}

	// get the balances from state
	fromBalance, err := t.userBalance(stub, fromCn)
	if err != nil {
		return errors.New("Error getting to or from userBalance:" + err.Error())
	}

	if fromBalance < trValue {
		return errors.New(fromCn + " does not have enough userBalance")
	}

	iterator, err := stub.GetStateByPartialCompositeKey(IndexCustomerAsset, []string{fromCn})
	if err != nil {
		return errors.New("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	restSum := trValue

	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return errors.New(err.Error())
		}

		_, parts, err := stub.SplitCompositeKey(kv.Key)
		sourceCn := parts[1]
		id := parts[2]

		asset:= Asset{}
		err = json.Unmarshal(kv.Value, &asset)
		if err != nil {
			return err
		}

		if asset.Value <= restSum {
			asset.History = append(asset.History, fromCn)
			_, err = t.createAsset(stub, IndexCustomerAsset, toCn, fromCn, asset.History, asset.Value)
			if err != nil {
				return errors.New("Error creating Asset for '" + toCn + "':" + err.Error())
			}
			err = t.removeAsset(stub, IndexCustomerAsset, fromCn, sourceCn, id)
			if err != nil {
				return errors.New("Error removing Asset '" + fromCn + "-" + sourceCn + "-" + id+ "':" + err.Error())
			}
			restSum -= asset.Value
		} else {
			_, err = t.storeAsset(stub, IndexCustomerAsset, fromCn, sourceCn, id, asset.History, asset.Value - restSum)
			if err != nil {
				return errors.New("Error updating Asset '" + fromCn + "-" + sourceCn + "-" + id+ "':" + err.Error())
			}
			asset.History = append(asset.History, fromCn)
			_, err = t.createAsset(stub, IndexCustomerAsset, toCn, fromCn, asset.History, restSum)
			if err != nil {
				return errors.New("Error creating Asset for '" + toCn + "':" + err.Error())
			}
			restSum = 0
		}

		if restSum == 0 {
			break
		}
	}

	if restSum != 0 {
		return errors.New("User Balance and the sum of his assets have different amount of tokens")
	}

	err = t.udpateUserBalance(stub, IndexCustomer, fromCn, trValue, true)
	err = t.udpateUserBalance(stub, IndexCustomer, toCn, trValue, false)
	if err != nil {
		return errors.New("Error setting to or from userBalance: " + err.Error())
	}


	return nil
}