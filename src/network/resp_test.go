package network

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserReadsInlineCommand(t *testing.T) {
	parser := NewParser(strings.NewReader("PING\r\n"))

	args, err := parser.ReadCommand()
	assert.NoError(t, err)
	assert.Equal(t, []string{"PING"}, args)
}

func TestParserReadsRESPArrayCommand(t *testing.T) {
	parser := NewParser(strings.NewReader("*2\r\n$4\r\nPING\r\n$5\r\nhello\r\n"))

	args, err := parser.ReadCommand()
	if errors.Is(err, ErrRESPArrayParsingNotImplemented) || errors.Is(err, ErrRESPBulkStringParsingNotImplemented) {
		t.Skip("TODO: Alan 需要实现 RESP array / bulk string 解析")
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, []string{"PING", "hello"}, args)
}

func TestParserReadPayload(t *testing.T) {
	parser := NewParser(strings.NewReader("hello\r\n"))

	payload, err := parser.readPayload(5)
	assert.NoError(t, err)
	assert.Equal(t, "hello", payload)
}

func TestParserReadsEmptyBulkString(t *testing.T) {
	parser := NewParser(strings.NewReader("*2\r\n$3\r\nGET\r\n$0\r\n\r\n"))

	args, err := parser.ReadCommand()
	assert.NoError(t, err)
	assert.Equal(t, []string{"GET", ""}, args)
}
