package sui

import (
	"context"
	"encoding/base64"
	"io"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
)

type invokeTransaction struct {
	from    *models.SuiKeyPair
	addr    string
	payload []byte
}

func newDeployPackageTransaction(appli *application, from *models.SuiKeyPair, client sui.ISuiAPI) ([]byte, error) {
	compiledModules := base64.StdEncoding.EncodeToString(appli.text)
	req := models.PublishRequest{
		Sender:          from.Address,
		CompiledModules: []string{compiledModules},
		Gas:             "",
		GasBudget:       3000, // TODO: ちゃんと計算する
	}

	resp, err := client.Publish(context.Background(), req)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := base64.StdEncoding.DecodeString(resp.TxBytes)
	if err != nil {
		return nil, err
	}

	return decodedBytes, nil
}

func newInvokeTransaction(key *models.SuiKeyPair, address string, payload []byte) *invokeTransaction {
	return &invokeTransaction{
		from:    key,
		addr:    address,
		payload: payload,
	}
}

func (this *invokeTransaction) encode(dest io.Writer) error {
	// encode transaction

	return nil
}
