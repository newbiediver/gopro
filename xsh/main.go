package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"github.com/newbiediver/golib/socket"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

const request uint8 = 1
const response uint8 = 2
const login uint8 = 3
const doLogin uint8 = 4
const loginResult uint8 = 5
const consoleLog uint8 = 6
const cshowlog uint8 = 7
const sshowlog uint8 = 8

type xshHandler struct {
	xshConnection *socket.TCP
	username      string
	server        string
	logShow       bool
}

var xsh *xshHandler
var doneLogin chan bool
var feedback chan bool
var tryLogin int

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	tryLogin = 0

	if len(os.Args) < 2 {
		panic("Usage > xsh address [port(default is 3214)]")
	}

	address := os.Args[1]
	port := 0
	if len(os.Args) > 2 {
		port, _ = strconv.Atoi(os.Args[2])
	} else {
		port = 3214
	}

	//address := "127.0.0.1"
	//port := 3214

	xsh = new(xshHandler)
	xsh.logShow = false
	if !xsh.connectServer(address, uint(port)) {
		red := color.New(color.FgHiRed)
		red.Println("xsh server is not running!")
		os.Exit(0)
	}

	doneLogin = make(chan bool, 1)
	<-doneLogin

	feedback = make(chan bool, 1)
	procMainLooop()
}

func (x *xshHandler) connectServer(address string, port uint) bool {
	x.xshConnection = new(socket.TCP)
	if !x.xshConnection.Connect(address, port) {
		return false
	}

	go x.xshConnection.ConnectionHandler(func(p []byte) {
		msg := p[0]
		m := uint8(msg)

		switch m {
		case response:
			onRecvString(string(p[3:]))
		case login:
			onRecvNeedLogin(string(p[3:]))
		case loginResult:
			if !xsh.parseAuthentication(string(p[3:])) {
				onRecvNeedLogin(string(p[3:]))
			} else {
				doneLogin <- true
			}
		case consoleLog:
			onConsoleLog(string(p[3:]))
		case sshowlog:
			onRecvString(string(p[3:]))
		default:
			fmt.Println("Unknown message")
		}
	}, func() {
		os.Exit(0)
	})

	return true
}

func (x *xshHandler) write(p []byte) {
	x.xshConnection.Send(p)
}

func (x *xshHandler) parseAuthentication(str string) bool {
	all := strings.Split(str, "\t")
	if all[0] != "ok" {
		return false
	}

	x.username = all[1]
	x.server = all[2]

	return true
}

func onRecvString(str string) {
	fmt.Println(str)
	feedback <- true
}

func onRecvNeedLogin(welcome string) {
	fmt.Println(welcome)
	if tryLogin == 3 {
		red := color.New(color.FgHiRed)
		red.Println("You failed to login three count!")
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("username : ")
	username := readString(reader)

	fmt.Print("password : ")
	pwdByte, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	pwd := string(pwdByte)

	//username := "root"
	//pwd := "xodlsl78\r\n"

	if runtime.GOOS == "windows" {
		pwd = strings.TrimRight(pwd, "\r\n")
	} else {
		pwd = strings.TrimRight(pwd, "\n")
	}

	fmt.Print("\n")
	tryLogin++

	auth := fmt.Sprintf("%s\t%s", username, pwd)

	sendPacket(auth, doLogin)
}

func sendPacket(str string, msg uint8) {
	str = fmt.Sprintf("%s\000", str)
	header := make([]byte, 3)
	packet := []byte(str)
	size := uint16(len(str) + 3)

	header[0] = byte(msg)
	binary.LittleEndian.PutUint16(header[1:], size)

	bytes := append(header, packet...)

	xsh.write(bytes)
}

func procMainLooop() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nWelcome!!")

	green := color.New(color.FgHiGreen)
	cyan := color.New(color.FgHiCyan)

	for {
		green.Print(xsh.username)
		fmt.Print("@[")
		cyan.Print(xsh.server)
		fmt.Print("]$ ")

		text := readString(reader)
		args := strings.Split(text, " ")

		if text == "" {
			continue
		}

		if text == "help" {
			showCommandList()
			continue
		}

		if text == "quit" || text == "exit" {
			fmt.Println("See you again!")
			break
		}

		if len(args) > 1 {
			text = args[0]
		}

		if text == "log" {
			sendPacket(args[1], cshowlog)
			<-feedback
			continue
		}

		if text == "terminate" {
			fmt.Println("Server will be terminated! Are you sure?")
			for {
				fmt.Print("Type yes or no => ")
				text = readString(reader)
				if text == "yes" {
					sendPacket("exit", request)
					break
				} else if text == "no" {
					break
				}
			}
		} else {
			sendPacket(text, request)
			<-feedback
		}
	}
}

func readString(reader *bufio.Reader) string {
	text, _ := reader.ReadString('\n')
	if runtime.GOOS == "windows" {
		text = strings.TrimRight(text, "\r\n")
	} else {
		text = strings.TrimRight(text, "\n")
	}

	return text
}

func showCommandList() {
	fmt.Println("Default command list for xsh")
	fmt.Println("Exit xsh : exit or quit")
	fmt.Println("Setting env value : set id value")
	fmt.Println("Listing env : list [*,id*]")
	fmt.Println("Switch remote console log : log show/hide")
	fmt.Println("Terminate Server : terminate")
}

func onConsoleLog(msg string) {
	logStrings := strings.Split(msg, "\x7F")

	timeString := logStrings[0]
	logString := logStrings[1]

	blue := color.New(color.FgHiBlue)
	blue.Print(timeString)
	fmt.Print(" ")
	fmt.Println(logString)
}
