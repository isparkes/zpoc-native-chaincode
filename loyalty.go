package main

import (
	"encoding/json"
	"fmt"
//	"strings"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type LoyaltyChaincode struct {
}

const KeySettings = "__settings"
const IndexCustomer = "cn~customer"
const IndexCustomerAsset = "cn~customer~asset"
const IndexBank = "cn~bank"
const IndexBanksCustomers = "cn~bank~customer"
const IndexShop = "cn~shop"

func (t *LoyaltyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()


	if function != "init" {
		return shim.Error("Expected 'init' function.")
	}

	if len(args) != 1 {
		return shim.Error("Expected 1 argument, but got " + strconv.Itoa(len(args)))
	}

	// get token data from JSON
	settings := Settings{}
	err := json.Unmarshal([]byte(args[0]), &settings)

	if err != nil {
		return shim.Error("Error parsing settings json")
	}

	err = stub.PutState(KeySettings, []byte(args[0]))
	if err != nil {
		return shim.Error("Error saving token data")
	}

	return shim.Success(nil)
}

func (t *LoyaltyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// call routing
	switch function {
	case "info":
		info, _ := stub.GetState(KeySettings)
		return shim.Success(info)
	case "transfer":
		return t.transfer(stub, args)
	case "userBalance":
		return t.userBalanceAsJson(stub, args)
	case "createActors":
		return t.createActors(stub, args)
	case "getCustomersNames":
		return t.getAllCostumerNames(stub, args)
	case "provideAsset":
		return t.provideAsset(stub, args)
	case "getMyCustomerList":
		return t.getMyCustomerList(stub, args)
	}

	return shim.Error("Incorrect function name: " + function)
}

func (t *LoyaltyChaincode) getSettings(stub shim.ChaincodeStubInterface) (Settings, error) {
	settingsByteArr, err := stub.GetState(KeySettings)
	if err != nil {
		return Settings{}, err
	}

	settings := Settings{}
	err = json.Unmarshal(settingsByteArr, &settings)
	if err != nil {
		return Settings{}, err
	}

	return settings, nil
}

func (t *LoyaltyChaincode) createActors(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("createActors expected 1 argument")
	}

	settings, err := t.getSettings(stub)
	if err != nil {
		return shim.Error("Error getting settings")
	}

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}

	// only admin is able to create another users
	if caller != settings.Admin {
		return shim.Error("I don't know you, " + caller + "!")
	}

	users := []User{}
	err = json.Unmarshal([]byte(args[0]), &users)
	if err != nil {
		return shim.Error("Error parsing users[] json")
	}

	for i := 0; i < len(users); i++ {
		switch users[i].Role {
		case "customer":
			err = t.setInitUserBalance(stub, IndexCustomer, users[i].Name, 0)
		case "bank":
			err = t.setInitUserBalance(stub, IndexBank, users[i].Name, 0)
		case "shop":
			err = t.setInitUserBalance(stub, IndexShop, users[i].Name, 0)
		}

		if err != nil {
			return shim.Error("Error creating user '" + users[i].Name + "'")
		}
	}

	b, err := json.Marshal(users)
	if err != nil {
	}

	return shim.Success(b)
}

func (t *LoyaltyChaincode) getAllCostumerNames(stub shim.ChaincodeStubInterface, args []string) pb.Response {


	iterator, err := stub.GetStateByPartialCompositeKey(IndexCustomer, []string{})
	if err != nil {
		return shim.Error("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	var result []*string = []*string{}
	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		_, parts, err := stub.SplitCompositeKey(kv.Key)
		cName := parts[0]

		result = append(result, &cName)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return shim.Error("Could not marshal json: " + err.Error())
	}

	return shim.Success(resultJson)
}

func (t *LoyaltyChaincode) provideAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}
	key, _ := stub.CreateCompositeKey(IndexBank, []string{caller})
	bankRes, err := stub.GetState(key)
	if err != nil || bankRes == nil {
		return shim.Error("I don't know you, " + caller + "!")
	}

	if len(args) != 1 {
		return shim.Error("provideAsset expected 1 argument")
	}

	params := Transfer{}
	err = json.Unmarshal([]byte(args[0]), &params)

	if params.Receiver == "" || params.Value <= 0 {
		return shim.Error("Bad request: wrong params!")
	}

	err = t.makeGiftToTheUserAsBank(stub, caller, params.Receiver, params.Value);
	if err != nil {
		return shim.Error("Could not commit gift to the user: " + err.Error())
	}

	return shim.Success([]byte("Git is committed"))
}

func (t *LoyaltyChaincode) getMyCustomerList(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}
	key, _ := stub.CreateCompositeKey(IndexBank, []string{caller})
	bankRes, err := stub.GetState(key)
	if err != nil || bankRes == nil {
		return shim.Error("I don't know you, " + caller + "!")
	}

	customers, err := t.getBanksCustomers(stub, caller)

	if err != nil {
		return shim.Error("Error getting banks customers: " + err.Error())
	}

	resultJson, err := json.Marshal(customers)
	if err != nil {
		return shim.Error("Could not marshal json: " + err.Error())
	}

	return shim.Success(resultJson)
}


func (t *LoyaltyChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	from, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}

	if len(args) != 1 {
		return shim.Error("Transfer expected 1 argument")
	}

	transfer := Transfer{}
	err = json.Unmarshal([]byte(args[0]), &transfer)
	if err != nil {
		return shim.Error("Error parsing transfer json")
	}

	// to prevent "generating" tokens because of
	// committed state reading
	if from == transfer.Receiver {
		return shim.Success(nil)
	}

	t.userToUserTransfer(stub, from, transfer.Receiver, transfer.Value)

	// send event
	transferEvent := TransferEvent{}
	transferEvent.Sender = from
	transferEvent.Receiver = transfer.Receiver
	transferEvent.Value = transfer.Value
	evtData, _ := json.Marshal(transferEvent)
	stub.SetEvent("Transfer", evtData)

	return shim.Success(nil)
}

func main() {
	err := shim.Start(&LoyaltyChaincode{})
	if err != nil {
		fmt.Errorf("Error starting Token chaincode: %s", err)
	}
}
