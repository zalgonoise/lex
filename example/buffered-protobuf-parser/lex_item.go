package protofile

type ProtoToken int

const (
	TokenEOF ProtoToken = iota
	TokenIDENT
	TokenTYPE
	TokenVALUE
	TokenEQUAL
	TokenDQUOTE
	TokenSEMICOL
	TokenLBRACE
	TokenRBRACE
	TokenSYNTAX
	TokenPACKAGE
	TokenMESSAGE
	TokenENUM
	TokenREPEATED

	TokenBOOL
	TokenUINT32
	TokenUINT64
	TokenSINT32
	TokenSINT64
	TokenINT32
	TokenINT64
	TokenFIXED32
	TokenFIXED64
	TokenSFIXED32
	TokenSFIXED64
	TokenDOUBLE
	TokenFLOAT
	TokenSTRING
	TokenBYTES
)

var keywords = map[string]ProtoToken{
	"syntax":   TokenSYNTAX,
	"package":  TokenPACKAGE,
	"message":  TokenMESSAGE,
	"enum":     TokenENUM,
	"repeated": TokenREPEATED,
}

var types = map[string]struct{}{
	"bool":     {},
	"uint32":   {},
	"uint64":   {},
	"sint32":   {},
	"sint64":   {},
	"int32":    {},
	"int64":    {},
	"fixed32":  {},
	"fixed64":  {},
	"sfixed32": {},
	"sfixed64": {},
	"double":   {},
	"float":    {},
	"string":   {},
	"bytes":    {},
}
