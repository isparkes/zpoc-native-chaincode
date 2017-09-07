package main

import (
	"encoding/binary"
	"encoding/json"
	//"errors"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/loyalty/chaincode/mock"
	"github.com/loyalty/chaincode/testdata"
	"testing"
	"fmt"
	"strconv"
)

var settings = Settings{
	Admin:        "testUser",
}

func initToken(t *testing.T) *mock.FullMockStub {
	loyalty := &LoyaltyChaincode{}

	stub := mock.NewFullMockStub("loyalty", loyalty)
	stub.MockCreator("default", testdata.TestUser1Cert)

	tokenBytes, _ := json.Marshal(settings)
	res := stub.MockInit("1", util.ToChaincodeArgs("init", string(tokenBytes)))
	if res.Status != shim.OK {
		t.Error("Loyalty cc init failed: " + res.Message)
	}

	st := Settings{Admin: "testUser"}
	stBytes, _ := json.Marshal(st)
	infoRes := stub.MockInvoke("1", util.ToChaincodeArgs("info", string(stBytes)))
	settings := Settings{}
	err := json.Unmarshal(infoRes.Payload, &settings)

	if (err != nil) {
		t.Error("Could not get info")
	}

	if settings.Admin != st.Admin {
		t.Error("Chaincode admin name is wrong")
	}

	return stub
}

func createActors(t *testing.T, stub *mock.FullMockStub, body string)  {
	res := stub.MockInvoke("1", util.ToChaincodeArgs("createActors", body))

	if res.Status != shim.OK {
		t.Errorf("Failed to createUser: %s", res.Message)
		t.FailNow()
	}

}

func provideAsset(t *testing.T, stub *mock.FullMockStub, body string)  {
	res := stub.MockInvoke("1", util.ToChaincodeArgs("provideAsset", body))

	if res.Status != shim.OK {
		t.Errorf("Failed to provideAsset: %s", res.Message)
		t.FailNow()
	}

}

func transferUserToUser(t *testing.T, stub *mock.FullMockStub, receiver string, value int)  {
	res := stub.MockInvoke("1", util.ToChaincodeArgs("transfer", `{"receiver":"` + receiver + `", "value":` + strconv.Itoa(value ) + `}`))

	if res.Status != shim.OK {
		t.Errorf("Failed to transfer: %s", res.Message)
		t.FailNow()
	}

}


func getMyCustomerList(t *testing.T, stub *mock.FullMockStub) []User {
	res := stub.MockInvoke("1", util.ToChaincodeArgs("getMyCustomerList"))

	if res.Status != shim.OK {
		t.Errorf("Failed to getBanksCustomerList: %s", res.Message)
		t.FailNow()
	}

	var users = []User{}
	err := json.Unmarshal(res.Payload, &users)
	if err != nil {
		t.Errorf("Failed to parse customer list: %s", err.Error())
		t.FailNow()
	}

	return users
}

func getUserBalance(t *testing.T, stub *mock.FullMockStub) User {
	res := stub.MockInvoke("1", util.ToChaincodeArgs("userBalance"))

	if res.Status != shim.OK {
		t.Errorf("Failed to getUserBalance: %s", res.Message)
		t.FailNow()
	}

	var user = User{}
	err := json.Unmarshal(res.Payload, &user)
	if err != nil {
		t.Errorf("Failed to parse UserBalance: %s", err.Error())
		t.FailNow()
	}

	return user
}

func getCustomerBalanceInfo(t *testing.T, stub *mock.FullMockStub) []TransferEvent {
	res := stub.MockInvoke("1", util.ToChaincodeArgs("customerBalanceInfo"))

	if res.Status != shim.OK {
		t.Errorf("Failed to getCustomerBalanceInfo: %s", res.Message)
		t.FailNow()
	}

	var transfers = []TransferEvent{}
	err := json.Unmarshal(res.Payload, &transfers)
	if err != nil {
		t.Errorf("Failed to parse ballanceInfo: %s", err.Error())
		t.FailNow()
	}

	return transfers
}

// ---------------------------------------------------------------------------------------------------------------------
// TESTS
// ---------------------------------------------------------------------------------------------------------------------

func TestCreateActors(t *testing.T) {
	stub := initToken(t)
	stub.MockCreator("default", testdata.TestUser1Cert)
	createActors(t, stub, `[{"role": "customer", "name": "user1"}, {"role": "bank", "name": "bank1"}, {"role": "shop", "name": "shop1"}]`)

	key, _ := stub.CreateCompositeKey(IndexCustomer, []string{"user1"})
	res, err := stub.GetState(key)
	if err != nil || res == nil {
		t.Errorf("Failed to create Customer: %s", "user1")
		t.FailNow()
	}

	key, _ = stub.CreateCompositeKey(IndexBank, []string{"bank1"})
	res, err = stub.GetState(key)
	if err != nil || res == nil {
		t.Errorf("Failed to create Bank: %s", "bank1")
		t.FailNow()
	}

	key, _ = stub.CreateCompositeKey(IndexShop, []string{"shop1"})
	res, err = stub.GetState(key)
	if err != nil || res == nil {
		t.Errorf("Failed to create Shop: %s", "shop1")
		t.FailNow()
	}
}

func TestProvideAsset(t *testing.T) {
	stub := initToken(t)
	stub.MockCreator("default", testdata.TestUser1Cert)
	createActors(t, stub, `[{"role": "customer", "name": "user1"}, {"role": "bank", "name": "testUser"}]`)
	provideAsset(t, stub, `{"receiver": "user1", "value": 1000}`)

	key, _ := stub.CreateCompositeKey(IndexCustomer, []string{"user1"})
	res, err := stub.GetState(key)
	if err != nil || res == nil || binary.LittleEndian.Uint64(res) != 1000 {
		t.Errorf("Failed to retrieve %s balance", "user1")
		t.FailNow()
	}

	key, _ = stub.CreateCompositeKey(IndexBanksCustomers, []string{"testUser", "user1"})
	res, err = stub.GetState(key)
	if err != nil || res == nil || binary.LittleEndian.Uint64(res) != 1000 {
		t.Errorf("Failed to get balance of %s for Bank %s", "user1", "testUser")
		t.FailNow()
	}

	key, _ = stub.CreateCompositeKey(IndexCustomerAsset, []string{"user1", "testUser", "0"})
	res, err = stub.GetState(key)
	if err != nil || res == nil {
		t.Errorf("Failed to retrieve Asset for user %s", "user1")
		t.FailNow()
	}

	users := getMyCustomerList(t, stub)
	if len(users) != 1 {
		t.Errorf("Expected 1 but got %d bank customers", len(users))
		t.FailNow()
	}
	if users[0].Name != "user1" {
		t.Errorf("Expected name for bank customer user1 but received %s bank customers", users[0].Name)
		t.FailNow()
	}
	if users[0].Balance != 1000 {
		t.Errorf("Expected customer balance 1000 but got %d", users[0].Balance)
		t.FailNow()
	}
}

func TestTransferUserToUserAsset(t *testing.T) {
	fmt.Printf("------------------------------------------------------\n")
	stub := initToken(t)
	stub.MockCreator("default", testdata.TestUser1Cert)
	createActors(t, stub, `[{"role": "bank", "name": "testUser"}, {"role": "customer", "name": "testUser"}, {"role": "customer", "name": "testUser2"}]`)
	provideAsset(t, stub, `{"receiver": "testUser", "value": 1000}`)
	provideAsset(t, stub, `{"receiver": "testUser2", "value": 111}`)
	transferUserToUser(t, stub, "testUser2", 500)

	stub.MockCreator("default", testdata.TestUser2Cert)
	transfers := getCustomerBalanceInfo(t, stub)
	if len(transfers) != 2 {
		t.Errorf("Expected 2 but got %d transfers", len(transfers))
		t.FailNow()
	}

	stub.MockCreator("default", testdata.TestUser1Cert)
	transferUserToUser(t, stub, "testUser2", 500)

	stub.MockCreator("default", testdata.TestUser2Cert)
	transfers = getCustomerBalanceInfo(t, stub)
	if len(transfers) != 3 {
		t.Errorf("Expected 3 but got %d transfers", len(transfers))
		t.FailNow()
	}

	for key, val := range stub.State {
		fmt.Printf("%s<->%s\n", key, val)
	}

	sum := uint64(0)
	for i:=0; i < len(transfers); i++ {
		sum += transfers[i].Value
	}

	userInfo := getUserBalance(t, stub)

	if userInfo.Balance != sum {
		t.Errorf("user balance(%d) and transactionSum(%d) do not coincide" , userInfo.Balance, sum)
		t.FailNow()
	}

	stub.MockCreator("default", testdata.TestUser1Cert)
	transfers = getCustomerBalanceInfo(t, stub)
	userInfo = getUserBalance(t, stub)

	sum = uint64(0)
	for i:=0; i < len(transfers); i++ {
		sum += transfers[i].Value
	}

	if userInfo.Balance != sum && sum != 0{
		t.Errorf("user balance(%d) and transactionSum(%d) do not coincide" , userInfo.Balance, sum)
		t.FailNow()
	}


}

func TestInitToken(t *testing.T) {
	initToken(t)
}

