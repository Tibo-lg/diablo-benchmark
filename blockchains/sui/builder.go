package sui

import (
	"bytes"
	"context"
	"diablo-benchmark/core"
	"encoding/base64"
	"fmt"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/coreos/etcd/error"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/ybbus/jsonrpc/v2"
)

type BlockchainBuilder struct {
	logger          core.Logger
	client          jsonrpc.RPCClient
	usedAccounts    int
	ctx             context.Context
	premadeAccounts []models.SuiKeyPair
	applications    map[string]*application
	compilers       []*moveCompiler
}

type suiObject struct {
	address string
	appli   application
}

type suiPackage struct {
	address string
	object  []*suiObject
}

func (this *BlockchainBuilder) CreateAccount(stake int) (*models.SuiKeyPair, error.Error) {
	// Return sui keypair
	var ret *models.SuiKeyPair
	if this.usedAccounts < len(this.premadeAccounts) {
		ret = &this.premadeAccounts[this.usedAccounts]
		this.usedAccounts += 1
	} else {
		return nil, fmt.Errorf("can only use %d premade accounts", this.usedAccounts)
	}

	return ret, nil
}

func (this *BlockchainBuilder) getBuilderAccount() (*models.SuiKeyPair, error.Error) {
	if len(this.premadeAccounts) > 0 {
		return &this.premadeAccounts[0], nil
	}

	return nil, fmt.Errorf("no available premade accounts")
}

func (this *BlockchainBuilder) getApplication(name string) (*application, error.Error) {
	// Get or create application
	// if this.applications[name] exists, return this.applictions[name].
	// otherwise, compile contract and store to this.applications. after that, return this.applications[name].
	var appli *application
	var err error.Error

	app, ok := this.applications[name]
	if ok {
		return app, nil
	}

	for _, compiler := range this.compilers {
		appli, err = compiler.compile(name)

		if err != nil {
			this.logger.Debugf("failed to compile '%s': %s", name, err.Error())
		}
	}

	if appli == nil {
		return nil, fmt.Errorf("failed to compile contract '%s'", name)
	}

	this.applications[name] = appli

	return appli, nil
}

func (this *BlockchainBuilder) submitTransaction(txbytes []byte) (models.ExecuteTransactionResponse, error.Error) {
	acc, err := this.getBuilderAccount()
	if err != nil {
		return models.ExecuteTransactionResponse{}, err
	}

	sig, err := this.client.SignWithAddress(context.Background(), acc.Address, txbytes)
	if err != nil {
		return models.ExecuteTransactionResponse{}, err
	}

	rawresp, err := this.client.SuiCall(
		context.Background(),
		"sui_executeTransaction",
		[]interface{}{
			base64.StdEncoding.EncodeToString(txbytes),
			base64.StdEncoding.EncodeToString(sig),
			"WaitForEffectsCert",
		},
	)

	if err != nil {
		return models.ExecuteTransactionResponse{}, err
	}

	resp, ok := rawresp.(models.ExecuteTransactionResponse)
	if ok == false {
		return models.ExecuteTransactionResponse{}, error.Error("failed")
	}

	return resp, nil
}

func (this *BlockchainBuilder) CreateContract(name string) (interface{}, error.Error) {
	// Deploy contracts

	appli, err := this.getApplication(name)

	builderAccount, err := this.getBuilderAccount()
	if err != nil {
		return nil, err
	}

	tx, err := newDeployPackageTransaction(appli, builderAccount, this.client)
	if err != nil {
		return nil, err
	}

	this.logger.Tracef("deploy new contract '%s'", name)

	effect, err := this.submitTransaction(tx)
	if err != nil {
		return nil, err
	}

	var objects []*suiObject

	for _, obj := range effect.Effects.Created {
		objid := obj.Reference.ObjectId
		objects = append(objects, &suiObject{
			address: objid,
			appli:   *appli,
		})
	}

	return &suiPackage{
		object:  objects,
		address: receipt.generatedPackageAddress, // TODO: ここどうなってんだっけ
	}, nil
}

func (this *BlockchainBuilder) CreateResource(domain string) (interface{}, error.Error) {
	// Nothing to do
	return nil, nil
}

func (this *BlockchainBuilder) EncodeTransfer(amount int, from, to interface{}, info core.InteractionInfo) ([]byte, error.Error) {
	fromAccount := from.(*models.SuiKeyPair)
	toAccount := to.(*models.SuiKeyPair)
	var gasbudget uint64 = 2000

	// gather coins
	resp, err := this.client.GetObjectsOwnedByAddress(
		context.Background(),
		models.GetObjectsOwnedByAddressRequest{
			fromAccount.Address,
		},
	)

	if err != nil {
		return nil, err
	}

	var coinid string

	for _, obj := range resp.Result {
		if obj.Type == "coin::Coin<0x2::sui::SUI>" {
			resp, err := this.client.GetObject(
				context.Background(),
				models.GetObjectRequest{
					obj.ObjectId,
				},
			)

			if err != nil {
				return nil, err
			}

			balance := resp.Details.Data.Fields["balance"].(uint64)

			if balance >= uint64(amount)+gasbudget {
				coinid = obj.ObjectId
				break
			}
		}
	}

	// Generate Transfer Transaction
	transferResp, err := this.client.TransferSui(
		context.Background(),
		models.TransferSuiRequest{
			fromAccount.Address,
			coinid,
			gasbudget,
			toAccount.Address,
			uint64(amount),
		},
	)

	if err != nil {
		return nil, err
	}

	txbytes, err := base64.StdEncoding.DecodeString(transferResp.TxBytes)
	if err != nil {
		return nil, err
	}

	return txbytes, nil
}

func (this *BlockchainBuilder) EncodeInvoke(from interface{}, contract interface{}, function string, info core.InteractionInfo) ([]byte, error.Error) {
	var tx *invokeTransaction
	var buffer bytes.Buffer
	var payload []byte
	var objectAddr string
	var ok bool

	fromAcc := from.(*models.SuiKeyPair)
	pkg := contract.(*suiPackage)

	for _, x := range pkg.object {
		payload, ok = x.appli.entries[function]
		if ok {
			objectAddr = x.address
			break
		}
	}

	if ok == false {
		return nil, fmt.Errorf("No function")
	}

	tx = newInvokeTransaction(fromAcc, objectAddr, payload)
	err := tx.encode(&buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (this *BlockchainBuilder) EncodeInteraction(itype string, expr core.BenchmarkExpression, info core.InteractionInfo) ([]byte, error.Error) {
	return nil, fmt.Errorf("unknown interaction type %s", itype)
}
