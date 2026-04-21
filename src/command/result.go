package command

import "fmt"

type resultKind int

const (
	simpleStringResult resultKind = iota
	bulkStringResult
	nullBulkStringResult
	integerResult
	errorResult
)

// Result 是命令执行后返回给客户端的 RESP 响应。
type Result struct {
	kind    resultKind
	text    string
	integer int64
}

func SimpleStringResult(text string) Result {
	return Result{
		kind: simpleStringResult,
		text: text,
	}
}

func BulkStringResult(text string) Result {
	return Result{
		kind: bulkStringResult,
		text: text,
	}
}

func NullBulkStringResult() Result {
	return Result{
		kind: nullBulkStringResult,
	}
}

func IntegerResult(value int64) Result {
	return Result{
		kind:    integerResult,
		integer: value,
	}
}

func ErrorResult(text string) Result {
	return Result{
		kind: errorResult,
		text: text,
	}
}

// Encode 把执行结果编码成 RESP2 响应。
func (result Result) Encode() []byte {
	switch result.kind {
	case simpleStringResult:
		return []byte("+" + result.text + "\r\n")
	case bulkStringResult:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(result.text), result.text))
	case nullBulkStringResult:
		return []byte("$-1\r\n")
	case integerResult:
		return []byte(fmt.Sprintf(":%d\r\n", result.integer))
	case errorResult:
		return []byte("-" + result.text + "\r\n")
	default:
		return []byte("-ERR unknown result kind\r\n")
	}
}
