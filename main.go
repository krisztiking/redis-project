package main

import (
	"fmt"
	"net"
	"strings"

	myResp "github.com/krisztiking/go-module-test"
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

	resp := myResp.NewResp(conn)
	aof.Read(resp)
	for {
		value, err := resp.Read()
		fmt.Println("value: ", value)
		if err != nil {
			fmt.Println(err)
			return
		}

		if value.Typ != "array" {
			fmt.Println("Incorrect request, expected array")
			continue
		}

		if len(value.Array) == 0 {
			fmt.Println("Incorrect request, expected non-empty array")
			continue
		}

		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		fmt.Println("Command: ", command, " Args: ", args)

		//writer := NewWriter(conn)
		writer := myResp.NewWriter(conn)
		handler, ok := myResp.Handlers[command]
		if !ok {
			fmt.Println("Unknown command: ", command)
			writer.Write(myResp.Value{Typ: "error", Str: "ERR unknown command"})
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
