package paymentservice

import "fmt"

type Wallet struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required,oneof=GHS"`
}

func (wallet Wallet) String() string {
	return fmt.Sprintf("Amount: %f, Currency: %s", wallet.Amount, wallet.Currency)
}
