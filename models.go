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
	Value uint64 `json:"value"`
}

type TransferEvent struct {
	Sender    string `json:"sender"`
	Receiver    string `json:"receiver"`
	Value uint64 `json:"value"`
}

type Allowance struct {
	Spender string `json:"spender"`
	Value   uint64 `json:"value"`
}

type AllowanceEvent struct {
	Spender string `json:"spender"`
	Value   uint64 `json:"value"`
	Shop string `json:"shop"`
}