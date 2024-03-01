package owntone

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		want    *SearchQuery
		wantErr bool
	}{
		{name: "Term only", args: args{target: "term"}, want: &SearchQuery{Terms: []string{"term"}, Offset: -1, Limit: -1}},
		{name: "2 Terms", args: args{target: "日本語 twice"}, want: &SearchQuery{Terms: []string{"日本語", "twice"}, Offset: -1, Limit: -1}},
		{name: "Term and offset", args: args{target: "term offset:1"}, want: &SearchQuery{Terms: []string{"term"}, Offset: 1, Limit: -1}},
		{name: "Term and offset, limit", args: args{target: "term offset:1 limit:2"}, want: &SearchQuery{Terms: []string{"term"}, Offset: 1, Limit: 2}},
		{name: "Term and offset, limit, types", args: args{target: "term offset:1 limit:2 type:album type:artist"}, want: &SearchQuery{Terms: []string{"term"}, Offset: 1, Limit: 2, Types: []SearchType{album, artist}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
