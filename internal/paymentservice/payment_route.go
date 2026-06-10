package paymentservice

import "github.com/labstack/echo/v5"

func Route(router *echo.Echo) {

	handler := handler{}

	router.POST("/api/v1/process-payment", handler.ProcessPayment)
}
