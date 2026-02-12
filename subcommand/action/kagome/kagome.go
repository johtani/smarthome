/*
Package kagome provides an action to tokenize Japanese text using the Kagome tokenizer.
It supports multiple dictionaries like IPA, Uni, and Neologd.
*/
package kagome

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ikawaha/kagome-dict-ipa-neologd"
	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/filter/ja"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"strings"

	"go.opentelemetry.io/otel"
)

// Dict represents the type of dictionary used by Kagome.
type Dict string

const (
	// IPA dictionary.
	IPA Dict = "ipa"
	// UNI dictionary.
	UNI Dict = "uni"
	// NEOLOGD dictionary.
	NEOLOGD Dict = "neologd"
)

// KagomeAction represents an action that tokenizes text using Kagome.
type KagomeAction struct {
	name       string
	dictionary Dict
}

// Run executes the Kagome tokenization action.
func (a KagomeAction) Run(ctx context.Context, args string) (string, error) {
	_, span := otel.Tracer("kagome").Start(ctx, "KagomeAction.Run")
	defer span.End()
	var dict *dict.Dict
	switch {
	case a.dictionary == UNI:
		dict = uni.Dict()
	case a.dictionary == NEOLOGD:
		dict = ipaneologd.Dict()
	default:
		dict = ipa.Dict()
	}
	parsedArgs := a.parseArgs(args)

	t, err := tokenizer.New(dict)
	if err != nil {
		return "", fmt.Errorf("tokenizer initialization failed, %w", err)
	}

	tokens := t.Analyze(parsedArgs.text, parsedArgs.mode)
	var buf bytes.Buffer
	if parsedArgs.filter {
		f, err := ja.NewFilter()
		if err != nil {
			return "", fmt.Errorf("filter initialization failed, %w", err)
		}
		f.Drop(&tokens)
	}
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		fmt.Fprintf(&buf, "%s\t%s\n", token.Surface, strings.Join(token.Features(), ","))
	}
	return "```\n" + buf.String() + "```", nil
}

// Args represents the arguments for KagomeAction.
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
	switch len(inputs) {
	case 0:
		text = ""
	case 1:
		text = inputs[0]
	default:
		var tmp []string
		for _, input := range inputs {
			if strings.HasPrefix(input, "-") {
				option := input[1:]
				switch option {
				case tokenizer.Search.String():
					mode = tokenizer.Search
				case tokenizer.Extended.String():
					mode = tokenizer.Extended
				case "filter":
					filter = true
				default:
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

// NewKagomeAction creates a new KagomeAction with the specified dictionary.
func NewKagomeAction(dict Dict) KagomeAction {
	return KagomeAction{
		name:       "Tokenize by Kagome with some Dictionaries",
		dictionary: dict,
	}
}
