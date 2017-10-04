package main

type Settings struct {
	Admin        string `json:"admin"`
}

type Asset struct {
	History    	[]string `json:"history"`
	Value   	uint64 `json:"value"`
	Info  		InfoEntry `json:"info"`
}

type User struct {
	Role    	string `json:"role"`
	Name        string `json:"name"`
	Balance   	uint64 `json:"userBalance"`
	BalanceHistory []HistoryEntry `json:"balanceHistory"`
}

type Transfer struct {
	Receiver    string `json:"receiver"`
	Value 		uint64 `json:"value"`
}

type BankObligation struct {
	Bank    string `json:"bank"`
	Value 	uint64 `json:"value"`
}

type TransferEvent struct {
	Sender		string `json:"sender"`
	Receiver    string `json:"receiver"`
	Value 		uint64 `json:"value"`
	Info  		InfoEntry `json:"info"`
}

type Allowance struct {
	Buyer string `json:"buyer"`
	Value uint64 `json:"value"`
}

type AllowanceEvent struct {
	Buyer string `json:"buyer"`
	Value uint64 `json:"value"`
	Shop  string `json:"shop"`
	Info  InfoEntry `json:"info"`
}

type HistoryEntry struct {
	Value 		string `json:"value"`
	TxId 		string `json:"txId"`
	Timestamp 	int64 `json:"timeStamp"`
}

type InfoEntry struct {
	TxId 		string `json:"txId"`
	Timestamp 	int64 `json:"timeStamp"`
}

type ValueType string

const (
	String  = ValueType("string")
	UInt64 	= ValueType("uInt64")
)
