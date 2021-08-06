package contract

import (
	"encoding/hex"

	"github.com/decus-io/decus-keeper-client/eth/abi"
)

const (
	Available = iota
	DepositRequested
	DepositReceived
	WithdrawRequested
)

type Receipt struct {
	ReceiptId string
	abi.IDeCusSystemReceipt
}

func ReceiptIdToString(receiptId [32]byte) string {
	return "0x" + hex.EncodeToString(receiptId[:])
}
