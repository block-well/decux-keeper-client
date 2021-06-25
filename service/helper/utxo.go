package helper

import (
	"encoding/hex"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/decus-io/decus-keeper-client/btc"
	"github.com/decus-io/decus-keeper-client/eth/abi"
	"github.com/decus-io/decus-keeper-client/eth/contract"
)

func txidEqual(txidStr string, txidBytes chainhash.Hash) bool {
	return strings.EqualFold(txidStr, hex.EncodeToString(txidBytes[:]))
}

func UtxoByReceipt(receipt *contract.Receipt) (*btc.Utxo, error) {
	utxo, err := btc.QueryUtxo(receipt.GroupBtcAddress)
	if err != nil {
		return nil, err
	}
	// reduce the chance that different keepers select different utxo
	// (normally there won't be multiple utxo)
	sort.Slice(utxo, func(i, j int) bool {
		if utxo[i].Status.Block_Height == utxo[j].Status.Block_Height {
			return utxo[i].Txid < utxo[j].Txid
		} else {
			return utxo[i].Status.Block_Height < utxo[j].Status.Block_Height
		}
	})

	for _, v := range utxo {
		if v.Status.Confirmed && v.Value == receipt.AmountInSatoshi.Uint64() {
			if receipt.Status == contract.DepositRequested {
				if v.Status.Block_Time > receipt.UpdateTimestamp {
					return &v, nil
				}
			} else {
				if txidEqual(v.Txid, receipt.TxId) && receipt.Height == v.Status.Block_Height {
					return &v, nil
				}
			}
		}
	}

	return nil, nil
}

func UtxoByRefundData(refundData *abi.IDeCusSystemBtcRefundData) (*btc.Utxo, error) {
	utxo, err := btc.QueryUtxo(refundData.GroupBtcAddress)
	if err != nil {
		return nil, err
	}
	for _, v := range utxo {
		if txidEqual(v.Txid, refundData.TxId) {
			return &v, nil
		}
	}

	return nil, nil
}
