package main

import (
	"github.com/google/uuid"
)

type TransfereRequest struct {
    ToAccount string `json:"toAccount"`
    Ammount int64 `json:"ammount"`
}

type CreateAccountRequest struct{
    FirstName string `json:"firstName"`
    LastName string `json:"lastName"`
}
type Account struct {
    ID int64 `json:"id"`
    UUID  uuid.UUID `json:"uuid"`
    FirstName string `json:"firstName"`
    LastName string `json:"lastName"`
    Balance int64 `json:"balance"`
}

func NewAccount(FirstName,LastName string) *Account {
    return &Account{
        UUID: uuid.New(),
        FirstName: FirstName,
        LastName: LastName,
    }
}
