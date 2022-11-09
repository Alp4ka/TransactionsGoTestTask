package models

import (
	"fmt"
	"time"
)

type TransactionStatus string

const (
	Created    TransactionStatus = "CREATED"
	InProgress                   = "IN_PROGRESS"
	Cancelled                    = "CANCELLED"
	Failed                       = "FAILED"
	Success                      = "SUCCESS"
)

type Transaction struct {
	Id          int               `db:"id"`
	Timestamp   time.Time         `db:"timestamp"`
	FromBalance int               `db:"from_balance"`
	ToBalance   int               `db:"to_balance"`
	Value       float64           `db:"value"`
	Currency    string            `db:"currency"`
	Status      TransactionStatus `db:"status"`
}

type TransactionRequest struct {
	FromBalance int
	ToBalance   int
	Value       float64
	Currency    string
}

func (Transaction) FromTransactionRequest(request TransactionRequest) Transaction {
	fromBalance := request.FromBalance
	toBalance := request.ToBalance
	value := request.Value
	currency := request.Currency
	timestamp := time.Now()
	status := Created
	return Transaction{
		FromBalance: fromBalance,
		ToBalance:   toBalance,
		Value:       value,
		Currency:    currency,
		Timestamp:   timestamp,
		Status:      status,
	}
}

func (tr TransactionRequest) String() string {
	return fmt.Sprintf(
		"[TransactionRequest]: FromBalance: %d, ToBalance: %d, Value: %f, Currency: %s",
		tr.FromBalance,
		tr.ToBalance,
		tr.Value,
		tr.Currency,
	)
}

func (t Transaction) String() string {
	return fmt.Sprintf(
		"[Transaction]: Id: %d, Timestamp: %s, FromBalance: %d, ToBalance: %d, Value: %f, Currency: %s",
		t.Id,
		t.Timestamp,
		t.FromBalance,
		t.ToBalance,
		t.Value,
		t.Currency,
	)
}
