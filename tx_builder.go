package cardano

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/tclairet/cardano-go/crypto"
	"golang.org/x/crypto/blake2b"
)

const maxUint64 uint64 = 1<<64 - 1

type TXBuilderInput struct {
	input  transactionInput
	amount uint64
}

type TXBuilderOutput struct {
	address Address
	amount  uint64
}

type TXBuilder struct {
	tx       Transaction
	protocol ProtocolParams
	inputs   []TXBuilderInput
	outputs  []TXBuilderOutput
	ttl      uint64
	fee      uint64
	vkeys    map[string]crypto.ExtendedVerificationKey
	pkeys    map[string]crypto.ExtendedSigningKey
}

func NewTxBuilder(protocol ProtocolParams) *TXBuilder {
	return &TXBuilder{
		protocol: protocol,
		vkeys:    map[string]crypto.ExtendedVerificationKey{},
		pkeys:    map[string]crypto.ExtendedSigningKey{},
	}
}

func (builder *TXBuilder) AddInput(xvk crypto.ExtendedVerificationKey, txId transactionID, index, amount uint64) {
	input := TXBuilderInput{input: transactionInput{ID: txId.Bytes(), Index: index}, amount: amount}
	builder.inputs = append(builder.inputs, input)

	vkeyHashBytes := blake2b.Sum256(xvk)
	vkeyHashString := hex.EncodeToString(vkeyHashBytes[:])
	builder.vkeys[vkeyHashString] = xvk
}

func (builder *TXBuilder) AddInputWithoutSig(txId transactionID, index, amount uint64) {
	input := TXBuilderInput{input: transactionInput{ID: txId.Bytes(), Index: index}, amount: amount}
	builder.inputs = append(builder.inputs, input)
}

func (builder *TXBuilder) AddOutput(address Address, amount uint64) {
	output := TXBuilderOutput{address: address, amount: amount}
	builder.outputs = append(builder.outputs, output)
}

func (builder *TXBuilder) SetTtl(ttl uint64) {
	builder.ttl = ttl
}

func (builder *TXBuilder) SetFee(fee uint64) {
	builder.fee = fee
}

// This assumes that the builder inputs and outputs are defined
func (builder *TXBuilder) AddFee(address Address) error {
	// Set a temporary fee in order to serialize a valid transaction
	builder.SetFee(maxUint64)
	minFee := builder.calculateMinFee()
	inputAmount := uint64(0)
	for _, txIn := range builder.inputs {
		inputAmount += txIn.amount
	}

	outputAmount := uint64(0)
	for _, txOut := range builder.outputs {
		outputAmount += txOut.amount
	}
	outputWithFeeAmount := outputAmount + minFee

	if inputAmount > outputWithFeeAmount {
		minAda := builder.protocol.MinimumUtxoValue
		change := inputAmount - outputWithFeeAmount
		if change > minAda {
			feeChange := builder.feeForOuput(address, change)
			newFee := minFee + feeChange
			change = inputAmount - (outputAmount + newFee)

			if change > minAda {
				builder.AddOutput(address, change)
				builder.SetFee(newFee)
				//logger.Infow("Adding change output")
			} else {
				builder.SetFee(minFee + change)
				//logger.Infow("Burning remaining change")
			}
		} else {
			builder.SetFee(minFee + change)
			//logger.Infow("Burning remaining change")
		}

	} else if inputAmount == outputWithFeeAmount {
		builder.SetFee(minFee)
		//log.Infow("No remaining change")
	} else {
		return fmt.Errorf("Insuficient input in transaction")
	}

	return nil
}

func (builder *TXBuilder) calculateMinFee() uint64 {
	fakeXSigningKey := crypto.NewExtendedSigningKey([]byte{
		0x0c, 0xcb, 0x74, 0xf3, 0x6b, 0x7d, 0xa1, 0x64, 0x9a, 0x81, 0x44, 0x67, 0x55, 0x22, 0xd4, 0xd8, 0x09, 0x7c, 0x64, 0x12,
	}, "")

	body := builder.buildBody()

	witnessSet := transactionWitnessSet{}
	for _, vkey := range builder.vkeys {
		witness := vkeyWitness{VKey: fakeXSigningKey.ExtendedVerificationKey()[:32], Signature: fakeXSigningKey.Sign(vkey)}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	tx := Transaction{
		Body:       body,
		WitnessSet: witnessSet,
		Metadata:   nil,
	}

	return builder.calculateFee(&tx)
}

func (builder *TXBuilder) feeForOuput(address Address, amount uint64) uint64 {
	builderCpy := *builder

	// We don't care about the fee impact on the tx size since we are
	// calculating the fee difference
	builderCpy.SetFee(0)

	feeBefore := builderCpy.calculateMinFee()
	builderCpy.AddOutput(address, amount)
	feeAfter := builderCpy.calculateMinFee()

	return feeAfter - feeBefore
}

func (builder *TXBuilder) calculateFee(tx *Transaction) uint64 {
	txBytes := tx.Bytes()
	txLength := uint64(len(txBytes))
	return builder.protocol.MinFeeA*txLength + builder.protocol.MinFeeB
}

func (builder *TXBuilder) Sign(xsk crypto.ExtendedSigningKey) {
	pkeyHashBytes := blake2b.Sum256(xsk)
	pkeyHashString := hex.EncodeToString(pkeyHashBytes[:])
	builder.pkeys[pkeyHashString] = xsk
}

func (builder *TXBuilder) Build() Transaction {
	if len(builder.pkeys) != len(builder.vkeys) {
		panic("missing signatures")
	}

	body := builder.buildBody()
	witnessSet := transactionWitnessSet{}
	for _, pkey := range builder.pkeys {
		txHash := blake2b.Sum256(body.Bytes())
		publicKey := pkey.ExtendedVerificationKey()[:32]
		signature := pkey.Sign(txHash[:])
		witness := vkeyWitness{VKey: publicKey, Signature: signature}

		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	return Transaction{Body: body, WitnessSet: witnessSet, Metadata: nil}
}

func (builder *TXBuilder) BuildRawTransaction(receiver Address, pickedUtxos []Utxo, amount uint64) (string, []byte, error) {
	if builder.fee == 0 {
		return "", nil, fmt.Errorf("fee is not setted")
	}
	if builder.ttl == 0 {
		return "", nil, fmt.Errorf("ttl is not setted")
	}

	for _, utxo := range pickedUtxos {
		builder.AddInputWithoutSig(utxo.TxId, utxo.Index, utxo.Amount)
	}
	builder.AddOutput(receiver, amount)

/*
	changeAddress := pickedUtxos[0].Address
	err := builder.AddFee(changeAddress)
	if err != nil {
		return "", nil, err
	}
*/
	tx := builder.buildBody()
	hash:= builder.hash()
	return hex.EncodeToString(hash[:]), tx.Bytes(), nil
}

func (builder *TXBuilder) hash() [32]byte {
	body := builder.buildBody()
	return blake2b.Sum256(body.Bytes())
}

func (builder *TXBuilder) buildBody() transactionBody {
	inputs := make([]transactionInput, len(builder.inputs))
	for i, txInput := range builder.inputs {
		inputs[i] = transactionInput{
			ID:    txInput.input.ID,
			Index: txInput.input.Index,
		}
	}

	outputs := make([]transactionOutput, len(builder.outputs))
	for i, txOutput := range builder.outputs {
		outputs[i] = transactionOutput{
			Address: txOutput.address.Bytes(),
			Amount:  txOutput.amount,
		}
	}
	return transactionBody{
		Inputs:  inputs,
		Outputs: outputs,
		Fee:     builder.fee,
		Ttl:     builder.ttl,
	}
}

func pretty(v interface{}) string {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes)
}
