package sui

type BlockchainClient interface{}

type BlockchainClient struct {
}

func (this *BlockchainClient) DecodePayload(bytes []byte) (interface{}, error) {

}

func (this *BlockchainClient) TriggerInteraction(iact Interaction) error {

}
