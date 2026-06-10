package main

import (
	"Idempotency-Gateway/internal/paymentservice"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	app := echo.New()
	app.Use(middleware.RequestLogger())

	paymentservice.Route(app)

	if err := app.Start(":8080"); err != nil {
		app.Logger.Error("Failed to start server. ", "error", err)
	}
}
