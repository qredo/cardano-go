package cardano

import (
	"testing"
)

func TestBuildTxHash(t *testing.T) {
	type args struct {
		receiver    Address
		pickedUtxos []Utxo
		amount      uint64
		ttl         uint64
	}
	tests := []struct {
		args           args
		hasBurnedInput bool
		hash           string
		wantErr        bool
	}{
		{
			args: args{
				receiver: "addr_test1vrct863kj4p2tpyzjzmjweyttquttk4z2kw7h42alj4p7gqma8ms5",
				pickedUtxos: []Utxo{
					{
						Address: "addr_test1vrct863kj4p2tpyzjzmjweyttquttk4z2kw7h42alj4p7gqma8ms5",
						TxId:    "2432fc624eb86075fcf035ca198cd89eff491ee38c0ada3434eb70c3af797acc",
						Amount:  20982558002,
						Index:   0,
					},
				},
				amount: 20982393645,
				ttl:    39851191,
			},
			hasBurnedInput: false,
			hash:           "71e5121b92c53834937730f6f0a6cf692496714dc3a426bba302868edc76a72a",
		}, {
			args: args{
				receiver: "addr_test1vrct863kj4p2tpyzjzmjweyttquttk4z2kw7h42alj4p7gqma8ms5",
				pickedUtxos: []Utxo{
					{
						Address: "addr_test1vrk8cshw3k95vvvrkwktylv6avap3unpcq407w5et6sccdsz9n78e",
						TxId:    "d08ac114cc35cfab1cbfbebd892246d825c9f65322b21b1be7ae8ebdbea4533f",
						Amount:  20982558002,
						Index:   0,
					},
					{
						Address: "addr_test1vrk8cshw3k95vvvrkwktylv6avap3unpcq407w5et6sccdsz9n78e",
						TxId:    "71fc933c748c0184bb45d93e6095d79e70196505908c39414587314deb19d81e",
						Amount:  20982558002,
						Index:   0,
					},
					{
						Address: "addr_test1vrk8cshw3k95vvvrkwktylv6avap3unpcq407w5et6sccdsz9n78e",
						TxId:    "d08ac114cc35cfab1cbfbebd892246d825c9f65322b21b1be7ae8ebdbea4533f",
						Amount:  20982558002,
						Index:   0,
					},
					{
						Address: "addr_test1vrk8cshw3k95vvvrkwktylv6avap3unpcq407w5et6sccdsz9n78e",
						TxId:    "f811d93ee348d73208bc6e12a60dc88e9f7207607675416d8b37cb38b1de70c2",
						Amount:  20982558002,
						Index:   0,
					},
				},
				amount: 20982393645,
				ttl:    39851191,
			},
			hasBurnedInput: true,
			hash:           "2e1ec7c140297d6648ad982ce2ee89d28673ca93e50288c2b4806cdd71d2ff11",
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

		rawTx2, err := NewTransactionBodyWithTTL(tt.args.receiver, tt.args.pickedUtxos, tt.args.amount, tt.args.pickedUtxos[0].Address, tt.args.ttl)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewTransactionBodyWithTTL() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		hash2 := rawTx2.ID()
		if got, want := string(hash2), tt.hash; got != want {
			t.Errorf("NewTransactionBodyWithTTL() got = %v, want %v", got, want)
		}

		if got, want := rawTx.Fee, rawTx2.Fee; got != want  {
			t.Errorf("Different fee from two function got = %v, want %v", got, want)
		}
	}
}
