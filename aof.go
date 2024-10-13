package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	myResp "github.com/krisztiking/go-module-test"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	// Start a goroutine to sync AOF to disk every 1 second
	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

func (aof *Aof) Write(value myResp.Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read(resp *myResp.Resp) {
	origReader := resp.Reader
	resp.Reader = aof.rd
	defer func() { resp.Reader = origReader }()

	for {
		value, err := resp.Read()
		if err != nil {
			fmt.Println(aof.file.Name(), " restored")
			return
		}
		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		handler, ok := myResp.Handlers[command]
		if !ok {
			fmt.Println("Invalid command in aof file: ", command)
		}
		handler(args)
	}

}
