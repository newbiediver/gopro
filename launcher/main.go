package main

import (
	"github.com/newbiediver/golib/aes256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

var baseURL string
var keyFile string
var ivFile string

/*type historyInfo struct {
	Version int      `json:"version"`
	Files   []string `json:"files"`
}

type versionInfo struct {
	Latest  int           `json:"latest"`
	Root    string        `json:"root"`
	History []historyInfo `json:"history"`
}*/

type historyInfo struct {
	Version int
	Files   []string
}

type versionInfo struct {
	Latest  int
	Root    string
	History []historyInfo
}

type writeProgresser struct {
	totalSize uint64
}

func (pr *writeProgresser) Write(p []byte) (int, error) {
	n := len(p)
	pr.totalSize += uint64(n)
	pr.printProgress()

	return n, nil
}

func (pr *writeProgresser) printProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s", humanize.Bytes(pr.totalSize))
}

func (vi *versionInfo) extract(filename string) error {
	fullPath := filename
	if runtime.GOOS != "windows" {
		fullPath = getTmpPath()
		fullPath += filename
	}
	jsonData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, vi)
	if err != nil {
		return err
	}

	return nil
}

func (vi *versionInfo) localVersion() int {
	fullPath := getTmpPath()
	fullPath += "version"

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return 0
	}

	bytes, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return 0
	}

	str := string(bytes)
	ver, _ := strconv.Atoi(str)

	return ver
}

func (vi *versionInfo) verifyVersion() (bool, error) {
	fullPath := getTmpPath()
	fullPath += "version"

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return true, nil
	}

	bytes, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return false, err
	}

	str := string(bytes)
	ver, err := strconv.Atoi(str)

	if err != nil {
		return false, err
	}

	if vi.Latest > ver {
		return true, nil
	}

	return false, nil
}

func (vi *versionInfo) writeToLocal() {
	str := strconv.Itoa(vi.Latest)
	fullPath := getTmpPath()
	fullPath += "/version"

	ioutil.WriteFile(fullPath, []byte(str), 777)
}

func getFullPath() string {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)

	exPath += "/"

	return exPath
}

func getTmpPath() string {
	if runtime.GOOS == "windows" {
		return "./"
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	tmp := usr.HomeDir + "/Applications/FilsoSupported/"

	if _, err := os.Stat(tmp); os.IsNotExist(err) {
		os.MkdirAll(tmp, os.ModePerm)
	}

	return tmp
}

func isExistFile(container []string, file string) bool {
	for _, str := range container {
		if str == file {
			return true
		}
	}

	return false
}

func encodeFileName(filename string) (string, error) {
	aes := new(aes256.Cipher)
	aes.SetKey(keyFile)
	aes.SetIV(ivFile)

	b, err := aes.Encode([]byte(filename))
	if err != nil {
		return "", err
	}

	str := hex.EncodeToString(b)

	return str, nil
}

func decodeFileName(filename string) (string, error) {
	aes := new(aes256.Cipher)
	aes.SetKey(keyFile)
	aes.SetIV(ivFile)

	b, err := hex.DecodeString(filename)
	if err != nil {
		return "", err
	}

	b, err = aes.Decode(b)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (vi *versionInfo) collectDownloadableFiles(localVer int) []string {
	files := make([]string, 0)

	for _, info := range vi.History {
		if info.Version > localVer {
			for _, file := range info.Files {
				if !isExistFile(files, file) {
					files = append(files, file)
				}
			}
		}
	}

	return files
}

func makeRemoteURL(root, filename string) string {
	str := fmt.Sprintf("%s/%s/%s", baseURL, root, filename)
	return str
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}

	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	defer destination.Close()
	nByte, err := io.Copy(destination, source)

	return nByte, err
}

func copyDownloadedFiles(files []string) {
	/*if runtime.GOOS == "windows" {
		return
	}*/

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	tmpPath := getTmpPath()
	appPath := getFullPath()

	for _, str := range files {

		f, _ := decodeFileName(str)

		tmp := tmpPath + str
		app := appPath + f

		_, err := copyFile(tmp, app)
		if err != nil {
			panic(err)
		}

		os.Remove(tmp)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	keyFile = "cdd23d6804dd6e7f89fc375925a152c7"
	ivFile = "h5KsCjyx=&aC2G8@"

	if len(os.Args) > 1 {
		file, err := encodeFileName(os.Args[1])
		if err != nil {
			panic(err)
		}
		fmt.Printf("Encoded : %s\n", file)
		return
	}
	binName := "filso"

	fmt.Printf("Update contents of \"%s\" started\n", binName)

	// 집
	/*if runtime.GOOS == "windows" {
		baseURL = "http://10.10.10.2:9000/filso/win"
	} else {
		baseURL = "http://10.10.10.2:9000/filso/mac"
	}*/

	// Local
	/*if runtime.GOOS == "windows" {
		baseURL = "http://localhost:9000/filso/win"
	} else {
		baseURL = "http://localhost:9000/filso/mac"
	}*/

	// GKE
	if runtime.GOOS == "windows" {
		baseURL = "http://update.fil.so/filso/win"
	} else {
		baseURL = "http://update.fil.so/filso/mac"
	}

	fileURL := fmt.Sprintf("%s/versioninfo.json", baseURL)
	err := downloadFile("versioninfo.json", fileURL)
	if err != nil {
		panic(err)
	}

	vi := new(versionInfo)
	err = vi.extract("versioninfo.json")
	if err != nil {
		panic(err)
	}

	isDownload, err := vi.verifyVersion()
	if err != nil {
		panic(err)
	}

	if isDownload {
		localVer := vi.localVersion()
		files := vi.collectDownloadableFiles(localVer)
		for _, str := range files {
			fileURL = makeRemoteURL(vi.Root, str)
			err = downloadFile(str, fileURL)
			if err != nil {
				panic(err)
			}
		}
		vi.writeToLocal()

		copyDownloadedFiles(files)
	}

	tmp := getTmpPath() + "versioninfo.json"
	os.Remove(tmp)

	fmt.Println("Updating complete!!")

	time.Sleep(1 * time.Second)

	if runtime.GOOS == "windows" {
		cmd := exec.Command("filso.exe")
		cmd.Start()
	} else {
		exePath := getFullPath()
		exePath += "filso"
		cmd := exec.Command(exePath)
		cmd.Start()
	}

	time.Sleep(2 * time.Second)
}

func downloadFile(localPath, url string) error {
	fullPath := localPath
	if runtime.GOOS != "windows" {
		exPath := getTmpPath()

		fullPath = exPath + localPath
	}
	if localPath == "versioninfo.json" {
		fmt.Println("Download informaton file...")
	} else {
		fmt.Printf("Download %s\n", localPath)
	}

	hr, e := http.Head(url)
	if e != nil {
		return e
	}
	if hr.StatusCode != http.StatusOK {
		return errors.New("No versioninfo file")
	}

	out, err := os.Create(fullPath + ".tmp")
	if err != nil {
		return err
	}

	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	counter := &writeProgresser{}

	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	out.Close()

	fmt.Print("\n")

	r, err := ioutil.ReadFile(fullPath + ".tmp")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fullPath, r, 0644)
	if err != nil {
		return err
	}

	err = os.Remove(fullPath + ".tmp")

	if err != nil {
		fmt.Println(err)
	}

	return nil
}
