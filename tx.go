package cardano

import (
	"encoding/hex"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

type ProtocolParams struct {
	MinimumUtxoValue uint64
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          uint64
	MinFeeB          uint64
}

type TransactionID string

func (id TransactionID) Bytes() []byte {
	bytes, err := hex.DecodeString(string(id))
	if err != nil {
		panic(err)
	}

	return bytes
}

type Transaction struct {
	_          struct{} `cbor:",toarray"`
	Body       TransactionBody
	WitnessSet transactionWitnessSet
	Metadata   *transactionMetadata // or null
}

func (tx *Transaction) Bytes() []byte {
	bytes, err := cbor.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (tx *Transaction) CborHex() string {
	return hex.EncodeToString(tx.Bytes())
}

func (tx *Transaction) ID() TransactionID {
	txHash := blake2b.Sum256(tx.Body.Bytes())
	return TransactionID(hex.EncodeToString(txHash[:]))
}

type transactionWitnessSet struct {
	VKeyWitnessSet []vkeyWitness `cbor:"0,keyasint,omitempty"`
	// TODO: add optional fields 1-4
}

type vkeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte   // ed25519 public key
	Signature []byte   // ed25519 signature
}

// Cbor map
type transactionMetadata map[uint64]transactionMetadatum

// This could be cbor map, array, int, bytes or a text
type transactionMetadatum struct{}

type TransactionBody struct {
	Inputs       []transactionInput  `cbor:"0,keyasint"`
	Outputs      []transactionOutput `cbor:"1,keyasint"`
	Fee          uint64              `cbor:"2,keyasint"`
	Ttl          uint64              `cbor:"3,keyasint"`
	Certificates []certificate       `cbor:"4,keyasint,omitempty"` // Omit for now
	Withdrawals  *uint               `cbor:"5,keyasint,omitempty"` // Omit for now
	Update       *uint               `cbor:"6,keyasint,omitempty"` // Omit for now
	MetadataHash *uint               `cbor:"7,keyasint,omitempty"` // Omit for now
}

func (body *TransactionBody) Bytes() []byte {
	bytes, err := cbor.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes
}

type RawTransaction struct {
	Inputs       []transactionInput  `cbor:"0,keyasint"`
	Outputs      []transactionOutput `cbor:"1,keyasint"`
	Fee          uint64              `cbor:"2,keyasint"`
	Ttl          uint64              `cbor:"3,keyasint"`
}

func (tx *RawTransaction) Hash() string {
	hash := blake2b.Sum256(tx.Bytes())
	return hex.EncodeToString(hash[:])
}

func (tx *RawTransaction) Bytes() []byte {
	bytes, err := cbor.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (tx *RawTransaction) AddSignature(publicKeys [][]byte, signatures [][]byte) (*Transaction, error) {
	if len(publicKeys) != len(signatures) {
		return nil, fmt.Errorf("missmatch length of publicKeys and signatures")
	}
	if len(signatures) != len(tx.Inputs) {
		return nil, fmt.Errorf("missmatch length of signatures and inputs")
	}
	witnessSet := transactionWitnessSet{}

	for i := 0; i <len(publicKeys); i++  {
		witness := vkeyWitness{VKey: publicKeys[i], Signature: signatures[i]}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	return &Transaction{
		Body: tx.toBody(),
		WitnessSet: witnessSet,
		Metadata: nil,
	}, nil
}

func (tx *RawTransaction) toBody() TransactionBody {
	return TransactionBody{
		Inputs:       tx.Inputs,
		Outputs:      tx.Outputs,
		Fee:          tx.Fee,
		Ttl:          tx.Ttl,
	}
}

type transactionInput struct {
	_     struct{} `cbor:",toarray"`
	ID    []byte   // HashKey 32 bytes
	Index uint64
}

type transactionOutput struct {
	_       struct{} `cbor:",toarray"`
	Address []byte
	Amount  uint64
}

// TODO: This should a cbor array with one element:
//  stake_registration
//	stake_deregistration
//	stake_delegation
//	pool_registration
//	pool_retirement
//	genesis_key_delegation
//	move_instantaneous_rewards_cert
type certificate struct{}
