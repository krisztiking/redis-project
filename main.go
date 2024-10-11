package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("Listening on port :9090")

	// Create a new server
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Println(err)
		return
	}

	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Print(err)
		return
	}
	defer aof.Close()

	// Listen for connections
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected to the client successfully!")

	defer conn.Close()

	resp := NewResp(conn)
	aof.Read(resp)
	for {
		value, err := resp.Read()
		fmt.Println("value: ", value)
		if err != nil {
			fmt.Println(err)
			return
		}

		if value.typ != "array" {
			fmt.Println("Incorrect request, expected array")
			continue
		}

		if len(value.array) == 0 {
			fmt.Println("Incorrect request, expected non-empty array")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		fmt.Println("Command: ", command, " Args: ", args)

		writer := NewWriter(conn)
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Unknown command: ", command)
			writer.Write(Value{typ: "error", str: "ERR unknown command"})
			continue
		}
		result := handler(args)
		writer.Write(result)

		fmt.Println("filnal result2: ", result)

		//Add 'SET' and 'HSET' to aof file,
		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}

	}
}
