package dto

import "time"

type CardDTO struct {
	Id      int64     `json:"id"`
	Number  string    `json:"number"`
	Balance int64     `json:"balance"`
	Issuer  string    `json:"issuer"`
	Holder  string    `json:"holder"`
	OwnerId int64     `json:"owner_id"`
	Status  string    `json:"status"`
	Created time.Time `json:"created"`
}

type TransactionDTO struct {
	Id             int64     `json:"id"`
	CardId         int64     `json:"card_id"`
	Amount         int64     `json:"amount"`
	Created        time.Time `json:"created"`
	Status         string    `json:"status"`
	MccId          int64     `json:"mcc_id"`
	Description    string    `json:"description"`
	SupplierIconId int64     `json:"supplier_icon_id"`
}

type AnalyticCategories struct {
	MccId       int64  `json:"mcc_id"`
	Col         int64  `json:"col"`
	Description string `json:"description"`
}

type AnalyticSum struct {
	MccId       int64  `json:"mcc_id"`
	SumAmount   int64  `json:"sum_amount"`
	Description string `json:"description"`
}

type ErrDTO struct {
	Err string `json:"error"`
}
