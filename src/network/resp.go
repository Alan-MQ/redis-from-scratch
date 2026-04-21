package network

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	// ErrRESPArrayParsingNotImplemented 提醒下一步先把 RESP array 读通。
	ErrRESPArrayParsingNotImplemented = errors.New("RESP array parsing not implemented")

	// ErrRESPBulkStringParsingNotImplemented 提醒下一步补 bulk string。
	ErrRESPBulkStringParsingNotImplemented = errors.New("RESP bulk string parsing not implemented")
)

// Parser 负责把客户端字节流解析成命令参数列表。
type Parser struct {
	reader *bufio.Reader
}

// NewParser 创建一个 RESP/inline 命令解析器。
func NewParser(r io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(r),
	}
}

// ReadCommand 读取下一条命令。
// 当前先保留 inline command 兼容，这样你在补 RESP 前，PING 还能跑通。
func (parser *Parser) ReadCommand() ([]string, error) {
	if parser == nil || parser.reader == nil {
		return nil, io.EOF
	}

	prefix, err := parser.reader.Peek(1)
	if err != nil {
		return nil, err
	}

	if prefix[0] == '*' {
		return parser.readArrayCommand()
	}

	return parser.readInlineCommand()
}

func (parser *Parser) readInlineCommand() ([]string, error) {
	line, err := parser.readLine()
	if err != nil {
		return nil, err
	}

	args := strings.Fields(line)
	if len(args) == 0 {
		return nil, fmt.Errorf("ERR empty command")
	}

	return args, nil
}

func (parser *Parser) readArrayCommand() ([]string, error) {
	// 推荐顺序：
	// 1. 先用 readLine() 读取形如 "*3" 的数组头。
	// 2. 用 parseIntegerAfterPrefix(line, '*') 得到元素个数。
	// 3. 依次读取 count 个 bulk string，每个都调用 readBulkString()。
	// 4. 把结果按顺序 append 到 args，并返回。
	//
	// 为什么这是关键练习：
	// - Redis 客户端（如 redis-cli）默认发的就是 RESP Array。
	// - 命令执行器拿到的其实只是 argv，真正把“网络协议”翻译成 argv 的地方就在这里。
	args := []string{}
	commandCountLine, err := parser.readLine()
	if err != nil {
		return nil, err
	}
	commandCount, err := parseIntegerAfterPrefix(commandCountLine, '*')
	if err != nil {
		return nil, err
	}
	for i := 0; i < commandCount; i++ {
		arg, err := parser.readBulkString()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func (parser *Parser) readBulkString() (string, error) {
	// 推荐顺序：
	// 1. 读取形如 "$5" 的头。
	// 2. 用 parseIntegerAfterPrefix(line, '$') 得到 payload 长度。
	// 3. 再调用 readPayload(length) 把后面的精确字节读出来。
	// 4. 注意处理 $0 和非法长度。
	commandCountLine, err := parser.readLine()
	if err != nil {
		return "", err
	}
	payloadLength, err := parseIntegerAfterPrefix(commandCountLine, '$')
	if err != nil {
		return "", err
	}
	if payloadLength < 0 {
		return "", fmt.Errorf("protocol error: invalid bulk length %d", payloadLength)
	}
	payload, err := parser.readPayload(payloadLength)
	if err != nil {
		return "", err
	}
	return payload, nil
}

func (parser *Parser) readLine() (string, error) {
	line, err := parser.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("protocol error: expected CRLF terminator")
	}

	return strings.TrimSuffix(line, "\r\n"), nil
}

func (parser *Parser) readPayload(length int) (string, error) {
	if parser == nil || parser.reader == nil {
		return "", io.EOF
	}
	if length < 0 {
		return "", fmt.Errorf("protocol error: invalid bulk length %d", length)
	}

	buf := make([]byte, length+2)
	if _, err := io.ReadFull(parser.reader, buf); err != nil {
		return "", err
	}
	if buf[length] != '\r' || buf[length+1] != '\n' {
		return "", fmt.Errorf("protocol error: bulk string missing CRLF")
	}

	return string(buf[:length]), nil
}

func parseIntegerAfterPrefix(line string, prefix byte) (int, error) {
	if len(line) == 0 || line[0] != prefix {
		return 0, fmt.Errorf("protocol error: expected %q prefix", prefix)
	}

	value, err := strconv.Atoi(line[1:])
	if err != nil {
		return 0, fmt.Errorf("protocol error: invalid integer %q", line[1:])
	}

	return value, nil
}
