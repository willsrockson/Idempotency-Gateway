package paymentservice

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/labstack/echo/v5"
)

const (
	Processing = "PROCESSING"
	Completed  = "COMPLETED"
)

// IdempotencyRecord holds the state of a single transaction intent
type IdempotencyRecord struct {
	State       string
	PayloadHash string
	StatusCode  int
	Response    interface{}
	Ready       chan struct{} //channel used to block duplicate in flight requests
}

// Thread safe map to store our records
var store = struct {
	sync.RWMutex
	m map[string]*IdempotencyRecord
}{m: make(map[string]*IdempotencyRecord)}

type handler struct {
}

// generate SHA-256 hash
func hashPayload(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (h *handler) ProcessPayment(ctx *echo.Context) error {
	wallet := new(Wallet)
	if err := ctx.Bind(wallet); err != nil {
		return err
	}
	fmt.Println(wallet)

	return ctx.JSON(200, fmt.Sprintf("Charged %f %s", wallet.Amount, wallet.Currency))
}
