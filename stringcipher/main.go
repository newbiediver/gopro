package main

import (
	"encoding/hex"
	"fmt"
	"github.com/newbiediver/golib/aes256"
	"os"
)

type myCipher struct {
	aesCipher 	aes256.Cipher
	baseString 	string
	result 		string
}

var (
	cipherKey string = "00cb0ba6b5af9731146b2bde676c8831"
	cipherInitVector string = "7e913e79dc442275"
)

func main() {
	args := os.Args[1:]

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	if len(args) < 2 {
		panic("Usage> stringcipher {-e|-d} string")
	}

	cp := myCipher{ aesCipher: aes256.Cipher{}, baseString: args[1] }
	cp.aesCipher.SetKey(cipherKey)
	cp.aesCipher.SetIV(cipherInitVector)

	switch args[0] {
	case "-e":
		cp.encode()
	case "-d":
		cp.decode()
	default:
		panic(fmt.Sprintf("Unknown option %s", args[0]))
	}

	fmt.Printf("Result: %s\n", cp.result)
}

func (c *myCipher) encode() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	bytes, err := c.aesCipher.Encode([]byte(c.baseString))
	if err != nil {
		panic(err)
	}

	c.result = hex.EncodeToString(bytes)
}

func (c *myCipher) decode() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	bytes, err := hex.DecodeString(c.baseString)
	if err != nil {
		panic(err)
	}

	result, err := c.aesCipher.Decode(bytes)
	if err != nil {
		panic(err)
	}

	c.result = string(result)
}
