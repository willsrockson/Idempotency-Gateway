package paymentservice

import (
	"fmt"

	"github.com/labstack/echo/v5"
)

type handler struct {
}

func (h *handler) ProcessPayment(ctx *echo.Context) error {
	wallet := new(Wallet)
	if err := ctx.Bind(wallet); err != nil {
		return err
	}
	fmt.Println(wallet)

	return ctx.JSON(200, fmt.Sprintf("Charged %f %s", wallet.Amount, wallet.Currency))
}
