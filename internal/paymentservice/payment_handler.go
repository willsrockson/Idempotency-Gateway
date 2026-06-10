package paymentservice

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

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
	// Grab the key from headers
	idempKey := ctx.Request().Header.Get("Idempotency-Key")
	if idempKey == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Missing Idempotency-Key header"})
	}

	// Read the raw body to hash it - we have to put it back so ctx.Bind works later
	bodyBytes, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	ctx.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	currentHash := hashPayload(bodyBytes)

	// lock the store to safely read and write
	store.Lock()
	record, exists := store.m[idempKey]

	if exists {
		// User story 3: Same key, different body fraud/error check
		if record.PayloadHash != currentHash {
			store.Unlock()
			return ctx.JSON(http.StatusConflict, map[string]string{
				"error": "Idempotency key already used for a different request body.",
			})
		}

		// Bonus story: The first request is still running so let's wait for it.
		if record.State == Processing {
			store.Unlock() // let other requests through while we wait
			<-record.Ready // blocks here until the channel is closed

			// once unblocked, lock again to safely read the finished data
			store.RLock()
			defer store.RUnlock()
			ctx.Response().Header().Set("X-Cache-Hit", "true")
			return ctx.JSON(record.StatusCode, record.Response)
		}

		// User story 2: It's already completed so return cached response right away.
		store.Unlock()
		ctx.Response().Header().Set("X-Cache-Hit", "true")
		return ctx.JSON(record.StatusCode, record.Response)
	}

	// First time seeing this key. Create a new record.
	record = &IdempotencyRecord{
		State:       Processing,
		PayloadHash: currentHash,
		Ready:       make(chan struct{}),
	}
	store.m[idempKey] = record
	store.Unlock()

	//Payment processing simulation

	// User story 1: Simulate the 2 seconds bank delay
	time.Sleep(2 * time.Second)

	wallet := new(Wallet)
	if err := ctx.Bind(wallet); err != nil {
		// mark as failed so they can retry
		store.Lock()
		delete(store.m, idempKey)
		store.Unlock()
		return err
	}

	successMsg := fmt.Sprintf("Charged %.2f %s", wallet.Amount, wallet.Currency)

	// Update our store with the final result
	store.Lock()
	record.State = Completed
	record.StatusCode = http.StatusOK
	record.Response = successMsg
	close(record.Ready) //broadcast to anyone waiting that we are done
	store.Unlock()

	return ctx.JSON(http.StatusOK, successMsg)
}
