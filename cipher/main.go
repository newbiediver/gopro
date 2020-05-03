package main

import (
	"github.com/newbiediver/golib/aes256"
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	encryption = iota
	decryption
)

type argumentContainer struct {
	cryptoType int
	path       string
	save       string
}

var aesKey string
var aesIV string

var argsCont *argumentContainer
var cip *aes256.Cipher

func parseArgument(args []string) {
	argsCont = new(argumentContainer)
	vAll := list.New()
	for _, str := range args {
		vAll.PushBack(str)
	}

	argsCont.parse(vAll)
}

func (arg *argumentContainer) parse(all *list.List) {
	for elem := all.Front(); elem != nil; elem = elem.Next() {
		if elem.Value.(string) == "-e" {
			arg.cryptoType = encryption
		} else if elem.Value.(string) == "-d" {
			arg.cryptoType = decryption
		} else {
			if len(arg.path) == 0 {
				arg.path = elem.Value.(string)
			} else {
				arg.save = elem.Value.(string)
			}

		}
	}
}

func (arg *argumentContainer) isEncryption() bool {
	return arg.cryptoType == encryption
}

func (arg *argumentContainer) getPath() string {
	return arg.path
}

func (arg *argumentContainer) getSave() string {
	return arg.save
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	if len(os.Args) < 4 {
		panic("Usage > cipher <-e|-d> SourceFile TargetFile")
	}

	aesKey = "cdd23d6804dd6e7f89fc375925a152c7"
	aesIV = "h5KsCjyx=&aC2G8@"

	parseArgument(os.Args[1:])

	cip = new(aes256.Cipher)
	cip.SetKey(aesKey)
	cip.SetIV(aesIV)

	if argsCont.isEncryption() {
		bytes, err := ioutil.ReadFile(argsCont.getPath())
		if err != nil {
			panic(err)
		}

		bytes, err = cip.Encode(bytes)
		if err != nil {
			panic(err)
		}

		ioutil.WriteFile(argsCont.getSave(), bytes, os.ModePerm)
	} else {
		bytes, err := ioutil.ReadFile(argsCont.getPath())
		if err != nil {
			panic(err)
		}

		bytes, err = cip.Decode(bytes)
		if err != nil {
			panic(err)
		}

		ioutil.WriteFile(argsCont.getSave(), bytes, os.ModePerm)
	}

}
