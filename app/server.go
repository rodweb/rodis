package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	value, err := DecodeRESP(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Failed to decode RESP")
		return
	}
	command := value.Array()[0].String()
	args := value.Array()[1:]

	switch command {
	case "ping":
		conn.Write([]byte("+PONG\r\n"))
	case "echo":
		conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[0].String()), args[0].String())))
	default:
		conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
	}
}

type Type byte

const (
	SimpleString Type = '+'
	BulkString   Type = '$'
	Array        Type = '*'
)

type Value struct {
	typ   Type
	bytes []byte
	array []Value
}

func (v *Value) String() string {
	if v.typ == BulkString || v.typ == SimpleString {
		return string(v.bytes)
	}
	return ""
}

func (v *Value) Array() []Value {
	if v.typ == Array {
		return v.array
	}
	return []Value{}
}

func DecodeRESP(reader *bufio.Reader) (Value, error) {
	dataTypeByte, err := reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch string(dataTypeByte) {
	case "+":
		return decodeSimpleString(reader)
	case "$":
		return decodeBulkString(reader)
	case "*":
		return decodeArray(reader)
	}
	return Value{}, fmt.Errorf("Invalid RESP data type byte: %s", string(dataTypeByte))
}

func decodeSimpleString(reader *bufio.Reader) (Value, error) {
	bytes, err := readUntilCRLF(reader)
	if err != nil {
		return Value{}, err
	}
	return Value{
		typ:   SimpleString,
		bytes: bytes,
	}, nil
}

func decodeBulkString(reader *bufio.Reader) (Value, error) {
	countBytes, err := readUntilCRLF(reader)
	if err != nil {
		return Value{}, err
	}
	count, err := strconv.Atoi(string(countBytes))
	if err != nil {
		return Value{}, err
	}
	bytes := make([]byte, count+2)
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return Value{}, err
	}
	return Value{
		typ:   BulkString,
		bytes: bytes[:count],
	}, nil

}

func decodeInteger() {}

func decodeError() {}

func decodeArray(reader *bufio.Reader) (Value, error) {
	countBytes, err := readUntilCRLF(reader)
	if err != nil {
		return Value{}, err
	}
	count, err := strconv.Atoi(string(countBytes))
	if err != nil {
		return Value{}, err
	}
	array := []Value{}
	for i := 1; i <= count; i++ {
		value, err := DecodeRESP(reader)
		if err != nil {
			return Value{}, err
		}
		array = append(array, value)
	}
	return Value{
		typ:   Array,
		array: array,
	}, nil
}

func readUntilCRLF(reader *bufio.Reader) ([]byte, error) {
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		return []byte{}, err
	}
	fmt.Print(fmt.Printf("bytes: %s", bytes))
	return bytes[:len(bytes)-2], nil
}
