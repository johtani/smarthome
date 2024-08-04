package kagome

import (
	"testing"
)

func TestKagomeAction_Run(t *testing.T) {
	type fields struct {
		name       string
		dictionary Dict
	}
	type args struct {
		args string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "ipa", fields: fields{name: "ipa", dictionary: IPA}, args: args{args: "ipaだよ"}, want: "```\nipa\t名詞,一般,*,*,*,*,*\nだ\t助動詞,*,*,*,特殊・ダ,基本形,だ,ダ,ダ\nよ\t助詞,終助詞,*,*,*,*,よ,ヨ,ヨ\n```", wantErr: false},
		{name: "uni", fields: fields{name: "uni", dictionary: UNI}, args: args{args: "uniだよ"}, want: "```\nu\t記号,文字,*,*,*,*,ユー,Ｕ,u,ユー,u,ユー,記号,*,*,*,*\nn\t記号,文字,*,*,*,*,エヌ,Ｎ,n,エヌ,n,エヌ,記号,*,*,*,*\ni\t記号,文字,*,*,*,*,アイ,Ｉ,i,アイ,i,アイ,記号,*,*,*,*\nだ\t助動詞,*,*,*,助動詞-ダ,終止形-一般,ダ,だ,だ,ダ,だ,ダ,和,*,*,*,*\nよ\t助詞,終助詞,*,*,*,*,ヨ,よ,よ,ヨ,よ,ヨ,和,*,*,*,*\n```", wantErr: false},
		{name: "neologd", fields: fields{name: "neologd", dictionary: NEOLOGD}, args: args{args: "neologdだよ"}, want: "```\nneologd\t名詞,固有名詞,一般,*,*,*,NEologd,ネオログディー,ネオログディー\nだ\t助動詞,*,*,*,特殊・ダ,基本形,だ,ダ,ダ\nよ\t助詞,終助詞,*,*,*,*,よ,ヨ,ヨ\n```", wantErr: false},
		{name: "ipa search", fields: fields{name: "ipa", dictionary: IPA}, args: args{args: "search ipaだよ"}, want: "```\nipa\t名詞,一般,*,*,*,*,*\nだ\t助動詞,*,*,*,特殊・ダ,基本形,だ,ダ,ダ\nよ\t助詞,終助詞,*,*,*,*,よ,ヨ,ヨ\n```", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := KagomeAction{
				name:       tt.fields.name,
				dictionary: tt.fields.dictionary,
			}
			got, err := a.Run(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}
