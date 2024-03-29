package service

import (
	"log"
	"strconv"

	"github.com/block-well/decux-keeper-client/btc"
	"github.com/block-well/decux-keeper-client/config"
	"github.com/block-well/decux-keeper-client/coordinator"
	"github.com/block-well/decux-keeper-client/eth"
	"github.com/block-well/decux-keeper-client/eth/contract"
	"github.com/block-well/decux-keeper-client/service/helper"
	"github.com/block-well/decux-keeper-proto/golang/message"
)

type DepositService struct {
	taskManager *helper.TaskManager
}

func NewDepositService() *DepositService {
	return &DepositService{
		taskManager: helper.NewTaskManager(),
	}
}

func (s *DepositService) signDeposit(receipt *contract.Receipt, utxo *btc.Utxo) error {
	signature, err := eth.SignDeposit(receipt.ReceiptId,
		"0x"+utxo.Txid, strconv.FormatInt(int64(utxo.Status.Block_Height), 10))
	if err != nil {
		return err
	}

	op := &message.Operation{
		Operation: &message.Operation_DepositSignature{
			DepositSignature: &message.DepositSignature{
				ReceiptId:   receipt.ReceiptId,
				TxId:        utxo.Txid,
				BlockHeight: uint64(utxo.Status.Block_Height),
				Signature:   signature,
			},
		},
	}
	return coordinator.SendOperation(op, nil)
}

func (s *DepositService) ProcessDeposit(receipt *contract.Receipt) error {
	if !s.taskManager.Available(receipt.ReceiptId) {
		return nil
	}

	utxo, err := helper.UtxoByReceipt(receipt)
	if err != nil {
		return err
	}
	if utxo == nil {
		return nil
	}

	height, err := btc.QueryHeight()
	if err != nil {
		return err
	}
	confirmations := height - utxo.Status.Block_Height + 1
	if confirmations < config.C.Btc.Confirmations {
		return nil
	}

	if err := s.signDeposit(receipt, utxo); err != nil {
		return err
	}
	s.taskManager.CheckLater(receipt.ReceiptId)
	log.Print("deposit signed: ", receipt.ReceiptId, " txid: ", utxo.Txid,
		" confirmations: ", confirmations)

	return nil
}
