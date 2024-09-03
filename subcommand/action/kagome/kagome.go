package kagome

import (
	"bytes"
	"fmt"
	"github.com/ikawaha/kagome-dict-ipa-neologd"
	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/filter/ja"
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
	var dict *dict.Dict
	if a.dictionary == UNI {
		dict = uni.Dict()
	} else if a.dictionary == NEOLOGD {
		dict = ipaneologd.Dict()
	} else {
		dict = ipa.Dict()
	}
	parsedArgs := a.parseArgs(args)

	t, err := tokenizer.New(dict)
	if err != nil {
		return "", fmt.Errorf("tokenizer initialization failed, %w", err)
	}
	// TODO Slack用の返信をできるようにしたい

	tokens := t.Analyze(parsedArgs.text, parsedArgs.mode)
	var buf bytes.Buffer
	if parsedArgs.filter {
		f, err := ja.NewFilter()
		if err != nil {
			return "", fmt.Errorf("filter initialization failed, %w", err)
		} else {
			f.Drop(&tokens)
		}
	}
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		fmt.Fprintf(&buf, "%s\t%s\n", token.Surface, strings.Join(token.Features(), ","))
	}
	return "```\n" + buf.String() + "```", nil
}

type Args struct {
	mode   tokenizer.TokenizeMode
	text   string
	filter bool
}

func (a KagomeAction) parseArgs(args string) Args {
	inputs := strings.Fields(args)
	var text string
	filter := false
	mode := tokenizer.Normal
	if len(inputs) == 0 {
		text = ""
	} else if len(inputs) == 1 {
		text = inputs[0]
	} else {
		var tmp []string
		for _, input := range inputs {
			if strings.HasPrefix(input, "-") {
				input = input[1:]
				if input == tokenizer.Search.String() {
					mode = tokenizer.Search
				} else if input == tokenizer.Extended.String() {
					mode = tokenizer.Extended
				} else if input == "filter" {
					filter = true
				} else {
					tmp = append(tmp, input)
				}
			} else {
				tmp = append(tmp, input)
			}
		}
		text = strings.Join(tmp, " ")
	}
	return Args{
		mode:   mode,
		text:   text,
		filter: filter,
	}
}

func NewKagomeAction(dict Dict) KagomeAction {
	return KagomeAction{
		name:       "Tokenize by Kagome with some Dictionaries",
		dictionary: dict,
	}
}
