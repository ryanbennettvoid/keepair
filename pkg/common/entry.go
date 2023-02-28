package common

type Entry struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

type EntryOperation struct {
	Action EntryOperationAction `json:"action"`
	Entry  Entry                `json:"entry"`
}

type EntryOperationAction string

var SetEntry = EntryOperationAction("set")
var DeleteEntry = EntryOperationAction("delete")
