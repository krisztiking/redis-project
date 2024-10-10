package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	fmt.Println("Start of readLine")
	for {
		b, err := r.reader.ReadByte()
		fmt.Println("b=", b, "Ascii")
		fmt.Printf("b: %c\n", b)

		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		fmt.Println(string(line), "n= ", n, "\r", line)

		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	fmt.Println("End of readLine\n")

	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	fmt.Println("Start of readIntiger")
	line, n, err := r.readLine() //beolvassa az egesz sort, byteonkent
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	fmt.Println("End of readIntiger. i64= ", i64)
	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	fmt.Println("Start of Read method")

	fmt.Printf("%c", _type)
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		println("\nNext: return readArray")
		return r.readArray()

	case BULK:
		println("\nNext: return readBulk")
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"
	fmt.Println(v.typ, "\n")

	// read length of array
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// foreach line, parse and read the value
	v.array = make([]Value, 0)
	for i := 0; i < len; i++ {
		fmt.Println("AAAAAA To Read ", i, "-jara   AAAAAAAA")
		val, err := r.Read()
		fmt.Println("val:", val)
		if err != nil {
			return v, err
		}

		// append parsed value to array
		v.array = append(v.array, val)
	}
	fmt.Println("v.array = ", v.array)

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}

	v.typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)

	r.reader.Read(bulk)

	v.bulk = string(bulk)
	fmt.Println(v.bulk, "- In the readBulk")

	// Read the trailing CRLF
	r.readLine()

	return v, nil
}

func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	len := len(v.array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
