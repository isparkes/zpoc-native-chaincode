package main

type Settings struct {
	Admin        string `json:"admin"`
}

type Asset struct {
	History    	[]string `json:"history"`
	Value   	uint64 `json:"value"`
}

type User struct {
	Role    	string `json:"role"`
	Name        string `json:"name"`
	Balance   	uint64 `json:"userBalance"`
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
}

type Allowance struct {
	Buyer string `json:"buyer"`
	Value uint64 `json:"value"`
}

type AllowanceEvent struct {
	Buyer string `json:"buyer"`
	Value uint64 `json:"value"`
	Shop  string `json:"shop"`
}
