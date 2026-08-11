package subcommand

import "testing"

func TestValidateResolvedArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		def    Definition
		args   string
		valid  bool
		reason string
	}{
		{name: "start music without args", def: NewStartMusicCmdDefinition(), args: "", valid: true, reason: argsValidationReasonValid},
		{name: "start music artist", def: NewStartMusicCmdDefinition(), args: "artist", valid: true, reason: argsValidationReasonValid},
		{name: "start music genre", def: NewStartMusicCmdDefinition(), args: "genre", valid: true, reason: argsValidationReasonValid},
		{name: "start music rejects invented prefix", def: NewStartMusicCmdDefinition(), args: "mode:random", valid: false, reason: argsValidationReasonInvalidPrefix},
		{name: "start music rejects unknown enum", def: NewStartMusicCmdDefinition(), args: "random", valid: false, reason: argsValidationReasonInvalidEnum},
		{name: "search and play keyword and type", def: NewSearchAndPlayMusicCmdDefinition(), args: "Meja type:artist", valid: true, reason: argsValidationReasonValid},
		{name: "search and play rejects keyword prefix", def: NewSearchAndPlayMusicCmdDefinition(), args: "keyword:Meja", valid: false, reason: argsValidationReasonInvalidPrefix},
		{name: "search and play requires keyword", def: NewSearchAndPlayMusicCmdDefinition(), args: "type:artist", valid: false, reason: argsValidationReasonMissingRequired},
		{name: "search and play rejects unknown type", def: NewSearchAndPlayMusicCmdDefinition(), args: "Meja type:performer", valid: false, reason: argsValidationReasonInvalidEnum},
		{name: "search and play preserves colon in free-form keyword", def: NewSearchAndPlayMusicCmdDefinition(), args: "Re:Re:", valid: true, reason: argsValidationReasonValid},
		{name: "search and play preserves optional search controls as free-form keyword", def: NewSearchAndPlayMusicCmdDefinition(), args: "Meja limit:2", valid: true, reason: argsValidationReasonValid},
		{name: "search and play rejects duplicate type", def: NewSearchAndPlayMusicCmdDefinition(), args: "Meja type:artist type:album", valid: false, reason: argsValidationReasonDuplicateArgument},
		{name: "no-args command", def: Definition{Name: "start ps5"}, args: "", valid: true, reason: argsValidationReasonValid},
		{name: "no-args command rejects args", def: Definition{Name: "start ps5"}, args: "unexpected", valid: false, reason: argsValidationReasonUnexpected},
		{name: "tokenizer text and mode", def: NewTokenizeIpaDefinition(), args: "解析する文章 -search", valid: true, reason: argsValidationReasonValid},
		{name: "tokenizer accepts text without mode", def: NewTokenizeIpaDefinition(), args: "解析する文章", valid: true, reason: argsValidationReasonValid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reason, valid := validateResolvedArgs(tt.def, tt.args)
			if valid != tt.valid {
				t.Fatalf("validateResolvedArgs() valid = %v, want %v (reason=%q)", valid, tt.valid, reason)
			}
			if reason != tt.reason {
				t.Fatalf("validateResolvedArgs() reason = %q, want %q", reason, tt.reason)
			}
		})
	}
}
