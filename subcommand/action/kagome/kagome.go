package kagome

import (
	"bytes"
	"fmt"
	"github.com/ikawaha/kagome-dict-ipa-neologd"
	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"strings"
)

type Dict string

const (
	IPA     Dict = "ipa"
	UNI     Dict = "uni"
	NEOLOGD Dict = "neologd"
)

type KagomeAction struct {
	name       string
	dictionary Dict
}

func (a KagomeAction) Run(args string) (string, error) {
	// TODO argsのパース
	var dict *dict.Dict
	if a.dictionary == UNI {
		dict = uni.Dict()
	} else if a.dictionary == NEOLOGD {
		dict = ipaneologd.Dict()
	} else {
		dict = ipa.Dict()
	}
	t, err := tokenizer.New(dict)
	if err != nil {
		return "", fmt.Errorf("tokenizer initialization failed, %w", err)
	}
	// TODO Slack用の返信をできるようにしたい

	tokens := t.Tokenize(args)
	var buf bytes.Buffer
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		fmt.Fprintf(&buf, "%s\t%s\n", token.Surface, strings.Join(token.Features(), ","))
	}
	return buf.String(), nil
}

func NewKagomeAction(dict Dict) KagomeAction {
	return KagomeAction{
		name:       "Tokenize by Kagome with IPA Dic",
		dictionary: dict,
	}
}
