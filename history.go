package main

import (
	"encoding/binary"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	b64 "encoding/base64"
	"errors"
)

func (t *LoyaltyChaincode) getHistory(stub shim.ChaincodeStubInterface, key string, valueType ValueType) ([]HistoryEntry, error) {

	historyIer, err := stub.GetHistoryForKey(key)

	if err != nil {
		return nil, err
	}

	var history []HistoryEntry = []HistoryEntry{}
	for i := 0; historyIer.HasNext(); i++ {
		modification, err := historyIer.Next()
		if err != nil {
			return nil, err
		}

		var value string;

		switch valueType {
		case UInt64:
			uintValue := binary.LittleEndian.Uint64(modification.Value)
			value = uintToString(uintValue)
		case String:
			value = string(modification.Value)
		default:
			value = b64.StdEncoding.EncodeToString(modification.Value)
		}

		historyEntry := &HistoryEntry{
			TxId: modification.TxId,
			Timestamp: modification.Timestamp.Seconds,
			Value:   value,
		}

		history = append(history, *historyEntry)
	}

	return history, nil
}

func (t *LoyaltyChaincode) getEntryInfo(stub shim.ChaincodeStubInterface, key string) (*InfoEntry, error) {

	history, err := t.getHistory(stub, key, String)
	if err != nil {
		return nil, err
	}

	n := len(history) - 1
	if n < 0 {
		return nil, errors.New("No entry history found")
	}


	return &InfoEntry{
		TxId: history[n].TxId,
		Timestamp: history[n].Timestamp,
	},
	nil
}
