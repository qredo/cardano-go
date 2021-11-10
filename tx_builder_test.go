package cardano

import (
	"testing"
)

func TestBuildTxHash(t *testing.T) {
	type args struct {
		receiver    Address
		pickedUtxos []Utxo
		amount      uint64
		ttl uint64
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
				amount:      20982393645,
				ttl: 39851191,
			},
			hash: "71e5121b92c53834937730f6f0a6cf692496714dc3a426bba302868edc76a72a",
		},
	}

	for _, tt := range tests {
		builder := NewTxBuilder(ProtocolParams{
			MinimumUtxoValue: 1000000,
			MinFeeA:          44,
			MinFeeB:          155381,
		})
		builder.SetTtl(tt.args.ttl)

		rawTx, err := builder.BuildTransactionBody(tt.args.receiver, tt.args.pickedUtxos, tt.args.amount, tt.args.pickedUtxos[0].Address)
		if (err != nil) != tt.wantErr {
			t.Errorf("BuildTxHash() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		hash := rawTx.ID()
		if got, want := string(hash), tt.hash; got != want {
			t.Errorf("BuildTxHash() got = %v, want %v", got, want)
		}
	}
}
