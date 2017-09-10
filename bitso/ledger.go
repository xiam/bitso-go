package bitso

import (
	"encoding/json"
	"errors"
)

type Operation uint8

const (
	OperationNone Operation = iota

	OperationFunding
	OperationWithdrawal
	OperationTrade
	OperationFee
)

var operationNames = map[Operation]string{
	OperationFunding:    "funding",
	OperationWithdrawal: "withdrawal",
	OperationTrade:      "trade",
	OperationFee:        "fee",
}

func (o Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Operation) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	for k, v := range operationNames {
		if v == z {
			*o = k
			return nil
		}
	}
	return errors.New("unsupported operation")
}

func (o Operation) String() string {
	if z, ok := operationNames[o]; ok {
		return z
	}
	panic("unsupported operation")
}

type BalanceUpdate struct {
	Currency Currency `json:"currency"`
	Amount   Monetary `json:"amount"`
}

type Transaction struct {
	EID            string                 `json:"eid"`
	Operation      Operation              `json:"operation"`
	CreatedAt      Time                   `json:"created_at"`
	BalanceUpdates []BalanceUpdate        `json:"balance_updates"`
	Details        map[string]interface{} `json:"details"`
}
