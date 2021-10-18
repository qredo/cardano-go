package cardano

import (
	"reflect"
	"testing"
)

func TestBuildTxHash(t *testing.T) {
	type args struct {
		receiver    Address
		pickedUtxos []Utxo
		amount      uint64
		ttl uint64
		fee uint64
	}
	tests := []struct {
		args    args
		hash    string
		wantErr bool
	}{
		{
			args: args{
				receiver:    "addr_test1vrct863kj4p2tpyzjzmjweyttquttk4z2kw7h42alj4p7gqma8ms5",
				pickedUtxos: []Utxo{
					{
						Address: "addr_test1vrct863kj4p2tpyzjzmjweyttquttk4z2kw7h42alj4p7gqma8ms5",
						TxId:    "2432fc624eb86075fcf035ca198cd89eff491ee38c0ada3434eb70c3af797acc",
						Amount:  20982558002,
						Index:   0,
					},
				},
				amount:      20982393997,
				ttl: 39851191,
				fee: 164005,
			},
			hash: "349bd9133fa19c3abb93b18cad4859e36280fbeca832c28fe7e78c9c961fcd3a",
		},
	}

	for _, tt := range tests {
		builder := NewTxBuilder(ProtocolParams{})
		builder.SetTtl(tt.args.ttl)
		builder.SetFee(tt.args.fee)

		hash, _, err := builder.BuildRawTransaction(tt.args.receiver, tt.args.pickedUtxos, tt.args.amount)
		if (err != nil) != tt.wantErr {
			t.Errorf("BuildTxHash() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(hash, tt.hash) {
			t.Errorf("BuildTxHash() got = %v, want %v", hash, tt.hash)
		}
	}
}
