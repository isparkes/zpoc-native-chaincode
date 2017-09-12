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

func (t *LoyaltyChaincode) updateUserBalance(stub shim.ChaincodeStubInterface, prefix string, cn string, delta uint64, negSign bool) error {

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
			return errors.New("balance of user '" + cn + "' is to small to proceed transaction")
		}
		newBalance -= delta
	} else {
		newBalance += delta
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, newBalance)
	return stub.PutState(key, data)
}

func (t *LoyaltyChaincode) userBalance(stub shim.ChaincodeStubInterface, prefix string, cn string) (uint64, error) {
	key, _ := stub.CreateCompositeKey(prefix, []string{cn})
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


	return t.updateUserBalance(stub, IndexCustomer, userCn, balance, false)
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
	fromBalance, err := t.userBalance(stub, IndexCustomer, fromCn)
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
		if err != nil {
			return errors.New("Error splitting composite key" + err.Error())
		}

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

	err = t.updateUserBalance(stub, IndexCustomer, fromCn, trValue, true)
	err = t.updateUserBalance(stub, IndexCustomer, toCn, trValue, false)
	if err != nil {
		return errors.New("Error setting to or from userBalance: " + err.Error())
	}


	return nil
}

func (t *LoyaltyChaincode) withdrawUserAssets(stub shim.ChaincodeStubInterface, userCn string, shopCn string, claim uint64) error {

	allowance, err := t.getAllowance(stub, shopCn, userCn)
	if err != nil {
		return err
	}

	if claim != allowance.Value {
		return errors.New("Shop claim and user allowance are not equal: " + err.Error())
	}

	iterator, err := stub.GetStateByPartialCompositeKey(IndexCustomerAsset, []string{userCn})
	if err != nil {
		return errors.New("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	restSum := allowance.Value

	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return errors.New(err.Error())
		}

		_, parts, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			return errors.New("Error splitting composite key" + err.Error())
		}

		sourceCn := parts[1]
		id := parts[2]

		asset:= Asset{}
		err = json.Unmarshal(kv.Value, &asset)
		if err != nil {
			return err
		}

		if asset.Value <= restSum {
			asset.History = append(asset.History, userCn)

			// move asset to shop
			_, err = t.createAsset(stub, IndexShopAsset, shopCn, userCn, asset.History, asset.Value)
			if err != nil {
				return errors.New("Error creating Asset for '" + shopCn + "':" + err.Error())
			}

			// move asset to bank since it shops claim
			asset.History = append(asset.History, shopCn)
			_, err = t.createAsset(stub, IndexBankAsset, asset.History[0], shopCn, asset.History, asset.Value)
			if err != nil {
				return errors.New("Error creating Asset for '" + asset.History[0] + "':" + err.Error())
			}

			// commit claim balance to the bank
			err = t.updateUserBalance(stub, IndexBank, asset.History[0],  asset.Value, false)
			if err != nil {
				return errors.New("Error updating bank balance: " + err.Error())
			}

			err = t.removeAsset(stub, IndexCustomerAsset, userCn, sourceCn, id)
			if err != nil {
				return errors.New("Error removing Asset '" + userCn + "-" + sourceCn + "-" + id+ "':" + err.Error())
			}
			restSum -= asset.Value
		} else {
			_, err = t.storeAsset(stub, IndexCustomerAsset, userCn, sourceCn, id, asset.History, asset.Value - restSum)
			if err != nil {
				return errors.New("Error updating Asset '" + userCn + "-" + sourceCn + "-" + id+ "':" + err.Error())
			}
			// move asset to shop
			asset.History = append(asset.History, userCn)
			_, err = t.createAsset(stub, IndexShopAsset, shopCn, userCn, asset.History, restSum)
			if err != nil {
				return errors.New("Error creating Asset for '" + shopCn + "':" + err.Error())
			}

			// move asset to bank since it shops claim
			asset.History = append(asset.History, shopCn)
			_, err = t.createAsset(stub, IndexBankAsset, asset.History[0], shopCn, asset.History, restSum)
			if err != nil {
				return errors.New("Error creating Asset for '" + asset.History[0] + "':" + err.Error())
			}

			// commit claim balance to the bank
			err = t.updateUserBalance(stub, IndexBank, asset.History[0], restSum, false)
			if err != nil {
				return errors.New("Error updating bank balance: " + err.Error())
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


	// update shop balance
	err = t.updateUserBalance(stub, IndexShop, shopCn, allowance.Value, false)
	if err != nil {
		return errors.New("Error setting to or from userBalance: " + err.Error())
	}

	return t.removeAllowance(stub, shopCn, userCn)
}