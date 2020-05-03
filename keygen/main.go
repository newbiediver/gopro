package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	if len(os.Args) < 2 {
		panic("Usage > keygen -bit=256")
	}

	arg := os.Args[1]
	kv := strings.Split(arg, "=")

	if kv[0] != "-bit" {
		panic("Usage > keygen -bit=256")
	}

	if kv[1] != "64" && kv[1] != "128" && kv[1] != "256" {
		panic("Key length support > 64bit 128bit 256bit")
	}

	keyLength := 0
	if kv[1] == "64" {
		keyLength = 8
	} else if kv[1] == "128" {
		keyLength = 16
	} else {
		keyLength = 32
	}

	key := generate(keyLength)

	fmt.Println(key)
}

func generate(keyLength int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sha := sha256.New()
	sha.Write([]byte(strconv.FormatFloat(r.Float64(), 'f', 6, 32)))

	str := fmt.Sprintf("%x", sha.Sum(nil))

	trim := str[:keyLength]
	return trim
}
