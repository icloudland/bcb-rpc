package bcb

import (
	"fmt"
	"github.com/icloudland/bcb-rpc/bcbjson"
)

func (c *JSONRPCClient) BlockHeight() (blockHeight int64, err error) {

	result := new(bcbjson.BlockHeightResult)
	_, err = c.Call("bcb_blockHeight", map[string]interface{}{}, result)
	if err != nil {
		err = fmt.Errorf("Cannot get block height, error=%s \n", err.Error())
		return
	}

	blockHeight = result.LastBlock

	return
}

func (c *JSONRPCClient) Block(height int64) (*bcbjson.BlockResult, error) {

	result := new(bcbjson.BlockResult)
	_, err := c.Call("bcb_block", map[string]interface{}{"height": height}, result)
	if err != nil {
		err = fmt.Errorf("Cannot get block data, height=%d, error=%s \n", height, err.Error())
	}

	return result, nil
}

func (c *JSONRPCClient) Transaction(txHash string) (*bcbjson.TxResult, error) {

	result := new(bcbjson.TxResult)
	_, err := c.Call("bcb_transaction", map[string]interface{}{"txHash": txHash}, result)
	if err != nil {
		err = fmt.Errorf("Cannot get transaction, txHash=%s, error=%s \n", txHash, err.Error())
		return nil, err
	}

	return result, nil
}

func (c *JSONRPCClient) BalanceOfToken(address string, tokenAddress string) (*bcbjson.BalanceResult, error) {

	result := new(bcbjson.BalanceResult)

	_, err := c.Call("bcb_balanceOfToken", map[string]interface{}{"address": address, "tokenAddress": tokenAddress,}, result)
	if err != nil {
		err = fmt.Errorf("Cannot get balance of token, address=%s, tokenAddress=%s, error=%s \n", address, tokenAddress, err.Error())
		return nil, err
	}

	return result, nil
}

func (c *JSONRPCClient) Balance(address string) (*bcbjson.BalanceResult, error) {

	result := new(bcbjson.BalanceResult)

	_, err := c.Call("bcb_balance", map[string]interface{}{"address": address}, result)
	if err != nil {
		err = fmt.Errorf("Cannot get balance, address=%s, error=%s \n", address, err.Error())
		return nil, err
	}

	return result, nil
}

func (c *JSONRPCClient) Transfer(name, accessKey, smcAddress, gasLimit, note, to, value string) (
	*bcbjson.TransferResult, error) {

	result := new(bcbjson.TransferResult)
	transferParam := bcbjson.TransferParam{
		SmcAddress: smcAddress,
		GasLimit:   gasLimit,
		Note:       note,
		To:         to,
		Value:      value,
	}

	_, err := c.Call("bcb_transfer",
		map[string]interface{}{"name": name, "accessKey": accessKey, "walletParams": transferParam}, result)
	if err != nil {
		err = fmt.Errorf("Cannot transfer, name=%s, walletParam=%v,\n error=%s \n", name, transferParam, err.Error())
		return nil, err
	}

	return result, nil
}
