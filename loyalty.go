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
const IndexBankAsset = "cn~bank~asset"
const IndexBanksCustomers = "cn~bank~customer"
const IndexShop = "cn~shop"
const IndexShopAsset = "cn~shop~asset"
const IndexShopAllowances = "cn~shop~allowances"

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
	case "customerBalance":
		return t.getUserBalance(stub, args, "customer")
	case "bankBalance":
		return t.getUserBalance(stub, args, "bank")
	case "shopBalance":
		return t.getUserBalance(stub, args, "shop")
	case "customerBalanceInfo":
		return t.customerBalanceInfo(stub, args)
	case "createActors":
		return t.createActors(stub, args)
	case "getCustomersNames":
		return t.getAllCostumerNames(stub, args)
	case "getShopClaims":
		return t.getShopClaims(stub, args)
	case "getBankObligations":
		return t.getBankObligations(stub, args)
	case "provideAsset":
		return t.provideAsset(stub, args)
	case "getMyCustomerList":
		return t.getMyCustomerList(stub, args)
	case "buy":
		return t.buy(stub, args)
	case "withdraw":
		return t.withdraw(stub, args)
	default:
		return shim.Error("Incorrect function name: " + function)
	}
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
		err := t.createUser(stub, users[i].Name, users[i].Role)

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

func (t *LoyaltyChaincode) getUserBalance(stub shim.ChaincodeStubInterface, args []string, role string) pb.Response {

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}

	prefix := IndexCustomer
	switch role{
	case "shop":
		prefix = IndexShop
	case "bank":
		prefix = IndexBank
	}

	balance, err := t.userBalance(stub, prefix, caller)
	if err != nil {
		return shim.Error("Error getting userBalance: " + err.Error())
	}

	balanceJson := User{
		Name:  caller,
		Balance: balance,
	}

	result, _ := json.Marshal(balanceJson)
	return shim.Success(result)
}

func (t *LoyaltyChaincode) customerBalanceInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}

	if !t.userExists(stub, caller, "customer") {
		return shim.Error("I don't know you, " + caller + "!")
	}

	iterator, err := stub.GetStateByPartialCompositeKey(IndexCustomerAsset, []string{caller})
	if err != nil {
		return shim.Error("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	var result []*TransferEvent = []*TransferEvent{}
	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		_, parts, err := stub.SplitCompositeKey(kv.Key)
		spender := parts[1]

		asset := Asset {}
		err = json.Unmarshal([]byte(kv.Value), &asset)
		if err != nil {
			return shim.Error("asset parsing error: " + err.Error())
		}

		transfer := TransferEvent {
			Receiver: caller,
			Sender: spender,
			Value: asset.Value,
		}

		result = append(result, &transfer)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return shim.Error("Could not marshal json: " + err.Error())
	}

	return shim.Success(resultJson)
}

func (t *LoyaltyChaincode) getShopClaims(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	bank, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}

	if !t.userExists(stub, bank, "bank") {
		return shim.Error("I don't know you, " + bank + "!")
	}

	iterator, err := stub.GetStateByPartialCompositeKey(IndexBankAsset, []string{bank})
	if err != nil {
		return shim.Error("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	var result []*Asset = []*Asset{}
	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		asset := Asset {}
		err = json.Unmarshal([]byte(kv.Value), &asset)
		if err != nil {
			return shim.Error("asset parsing error: " + err.Error())
		}

		result = append(result, &asset)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return shim.Error("Could not marshal json: " + err.Error())
	}

	return shim.Success(resultJson)
}

func (t *LoyaltyChaincode) getBankObligations(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	shop, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}

	if !t.userExists(stub, shop, "shop") {
		return shim.Error("I don't know you, " + shop + "!")
	}

	iterator, err := stub.GetStateByPartialCompositeKey(IndexShopAsset, []string{shop})
	if err != nil {
		return shim.Error("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	var result []*BankObligation = []*BankObligation{}
	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		asset := Asset {}
		err = json.Unmarshal([]byte(kv.Value), &asset)
		if err != nil {
			return shim.Error("asset parsing error: " + err.Error())
		}

		bankObligation := BankObligation {
			Bank: asset.History[0],
			Value: asset.Value,
		}

		result = append(result, &bankObligation)
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

	if !t.userExists(stub, params.Receiver, "customer") {
		return shim.Error("Bad request: receiver doesn't exist")
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
		return shim.Error("Transfer to yourself is not allowed")
	}

	if !t.userExists(stub, transfer.Receiver, "customer") {
		return shim.Error("Bad request: receiver doesn't exist")
	}


	err = t.userToUserTransfer(stub, from, transfer.Receiver, transfer.Value)
	if err != nil {
		return shim.Error(err.Error())
	}

	// send event
	transferEvent := TransferEvent{}
	transferEvent.Sender = from
	transferEvent.Receiver = transfer.Receiver
	transferEvent.Value = transfer.Value
	evtData, _ := json.Marshal(transferEvent)
	stub.SetEvent("Transfer", evtData)

	return shim.Success(evtData)
}

func main() {
	err := shim.Start(&LoyaltyChaincode{})
	if err != nil {
		fmt.Errorf("Error starting Token chaincode: %s", err)
	}
}

func (t *LoyaltyChaincode) buy(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	buyer, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}
	if len(args) != 1 {
		return shim.Error("Buy expected 1 argument")
	}

	transfer := Transfer{}
	err = json.Unmarshal([]byte(args[0]), &transfer)
	if err != nil {
		return shim.Error("Error parsing arguments")
	}

	if !t.userExists(stub, transfer.Receiver, "shop") {
		return shim.Error("Bad request: shop doesn't exist")
	}

	allowance, err := t.createAllowance(stub, transfer.Receiver, buyer, transfer.Value)
	if err != nil {
		return shim.Error(err.Error())
	}

	// send event
	allowanceEvent := AllowanceEvent{}
	allowanceEvent.Buyer = allowance.Buyer
	allowanceEvent.Shop = transfer.Receiver
	allowanceEvent.Value = allowance.Value
	evtData, _ := json.Marshal(allowanceEvent)
	stub.SetEvent("Buy", evtData)

	return shim.Success(nil)
}

func (t *LoyaltyChaincode) withdraw(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	shopCn, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error extracting user identity")
	}
	if len(args) != 1 {
		return shim.Error("Buy expected 1 argument")
	}

	allowance := Allowance{}
	err = json.Unmarshal([]byte(args[0]), &allowance)
	if err != nil {
		return shim.Error("Error parsing arguments")
	}

	if !t.userExists(stub, allowance.Buyer, "customer") {
		return shim.Error("Bad request: customer doesn't exist")
	}

	err = t.withdrawUserAssets(stub, allowance.Buyer, shopCn, allowance.Value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}