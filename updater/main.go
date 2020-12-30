package main

import (
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/newbiediver/gopro/updater/config"
	"github.com/shibukawa/configdir"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type writeProgresser struct {
	totalSize	uint64
}

type downloadFileInfo struct {
	fileName	string
	lastModTime	time.Time
}

var (
	baseUrl string
	downloadAllTotalSize uint64
	isDownload bool
	uIndex int
)

func (wp *writeProgresser) Write(p []byte) (int, error) {
	n := len(p)
	wp.totalSize += uint64(n)
	wp.printProgress()
	return n, nil
}

func (wp *writeProgresser) printProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s / %s", humanize.Bytes(wp.totalSize), humanize.Bytes(downloadAllTotalSize))
}

func createDownloadDirectory(path string) {
	_ = os.Mkdir(path, 0777)
}

func main() {
	sc := config.GetSystemConfig()
	sc.DefaultConfig()

	args := os.Args[1]

	uIndex, _ = strconv.Atoi(args)
	fmt.Printf( "Current uindex: %d\n", uIndex)

	fmt.Printf("Loading config file...updater.yml")
	yaml, isLoaded := sc.LoadYaml("updater.yml")
	if isLoaded {
		err := sc.SetConfig(yaml)
		if err != nil {
			panic(err)
		}
		fmt.Printf(" succeed\n")
	} else {
		fmt.Printf(" failed\n")
	}

	baseUrl = fmt.Sprintf("http://%s:%d", sc.UpdaterServer, sc.UpdaterPort)

	createDownloadDirectory(sc.DownloadPath)

	allList := listing()
	if allList != nil {
		newList := compareList(allList)

		if newList != nil {
			downloadFiles(newList)
		}
		updateFiles(sc.DownloadPath, ".")
	}

	launchProgram(sc.LaunchPath, sc.LaunchArg)
}

func listing() *UpdateList {
	fmt.Println("Generate remote contents information...")
	uri := baseUrl + "/list"
	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	updateList := new(UpdateList)
	err = json.Unmarshal(data, updateList)
	if err != nil {
		panic(err)
	}

	return updateList
}

func compareList(serverList *UpdateList) []downloadFileInfo {
	var downloadList []downloadFileInfo
	tmpLocation := time.UTC
	sc := config.GetSystemConfig()
	for _, v := range serverList.Files {
		filePath := "." + v.FileName
		downPath := sc.DownloadPath + v.FileName
		file, err := os.Open(filePath)
		downFile, err2 := os.Open(downPath)

		curFilePath := strings.ReplaceAll(v.FileName, "\\", "/")
		serverFileTime := time.Date(v.ModTimeYear, time.Month(v.ModTimeMonth), v.ModTimeDay, v.ModTimeHour, v.ModTimeMin, v.ModTimeSec, v.ModTimeNSec, tmpLocation)

		if err2 == nil {
			// 다운로드 임시 폴더에 파일이 있으면 패스
			downStat, _ := downFile.Stat()
			downFileTime := downStat.ModTime()
			downFileSize := downStat.Size()
			if downFileTime.Equal(serverFileTime) && downFileSize == v.FileSize {
				if file != nil {
					_ = file.Close()
				}
				_ = downFile.Close()
				continue
			}
		}

		if err != nil {
			curFileInfo := downloadFileInfo{fileName: curFilePath, lastModTime: serverFileTime}
			downloadList = append(downloadList, curFileInfo)
			downloadAllTotalSize = downloadAllTotalSize + uint64(v.FileSize)
			if downFile != nil {
				_ = downFile.Close()
			}
			continue
		}

		loc, _ := time.LoadLocation("UTC")
		stat, _ := file.Stat()
		localFileTime := stat.ModTime().In(loc)

		strLocalFileTime := fmt.Sprintf("%04d.%02d.%02d %02d:%02d:%02d", localFileTime.Year(), localFileTime.Month(), localFileTime.Day(), localFileTime.Hour(), localFileTime.Minute(), localFileTime.Second() )
		strServerFileTime := fmt.Sprintf("%04d.%02d.%02d %02d:%02d:%02d", serverFileTime.Year(), serverFileTime.Month(), serverFileTime.Day(), serverFileTime.Hour(), serverFileTime.Minute(), serverFileTime.Second() )

		if strLocalFileTime != strServerFileTime {
		//if !serverFileTime.Equal(localFileTime) {
			curFileInfo := downloadFileInfo{fileName: curFilePath, lastModTime: serverFileTime}
			downloadList = append(downloadList, curFileInfo)
			downloadAllTotalSize = downloadAllTotalSize + uint64(v.FileSize)
		}
		_ = file.Close()
		if downFile != nil {
			_ = downFile.Close()
		}
	}

	return downloadList
}

func downloadFiles(files []downloadFileInfo) {
	fmt.Println("Download updatable contents...")
	sc := config.GetSystemConfig()
	uri := baseUrl + "/contents"
	downloadDir := sc.DownloadPath

	counter := &writeProgresser{}
	isDownload = true

	for _, str := range files {
		s := uri + str.fileName
		resp, err := http.Get(s)
		if err != nil {
			panic(err)
		}

		localPath := downloadDir + str.fileName
		dirPath := filepath.Dir(localPath)
		if _, err = os.Stat(dirPath); os.IsNotExist(err) {
			err = os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		out, err := os.Create(localPath)
		if err != nil {
			panic(err)
		}


		_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
		if err != nil {
			panic(err)
		} else {

			_ = out.Close()
			_ = os.Chtimes(localPath, str.lastModTime.UTC(), str.lastModTime.UTC())
		}
	}
}

func updateIndex() {
	strUIndex := strconv.Itoa(uIndex)
	bytes := []byte(strUIndex)

	if runtime.GOOS == "windows" {
		dirPath := configdir.New("", "PADICP")
		cf := dirPath.QueryCacheFolder()
		_ = cf.WriteFile("uindex.bin", bytes)
	} else {
		dirPath := configdir.New("", "PADICP")
		cf := dirPath.QueryFolders(configdir.Global)
		_ = cf[0].WriteFile("uindex.bin", bytes)
	}
}

func updateFiles(srcPath, dstPath string) {
	sc := config.GetSystemConfig()
	if isDownload {
		fmt.Printf("\n")
	}
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {

		handler, err := os.Stat(path)
		if err != nil {
			return err
		}

		if handler.Name() == srcPath {
			return nil
		}

		sp := strings.Split(path, srcPath)
		if handler.IsDir() {
			dstDir := dstPath + sp[1]
			if _, err = os.Stat(dstDir); os.IsNotExist(err) {
				err = os.MkdirAll(dstDir, os.ModePerm)
				if err != nil {
					return err
				}
			}

			return nil
		}

		fmt.Printf("Update %s ...\n", sp[1])

		dst := dstPath + sp[1]
		_, err = copyFile(path, dst)

		return err
	})

	if err != nil {
		panic(err)
	}

	tmpPath := ".\\" + sc.DownloadPath + "\\"
	err = os.RemoveAll(tmpPath)

	if err != nil {
		fmt.Println(err)
	}

	updateIndex()
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

	if err == nil {
		_ = os.Chtimes(dst, sourceFileStat.ModTime(), sourceFileStat.ModTime())
	}

	return nByte, err
}

func launchProgram(programPath, launchArgs string) {
	fmt.Println("Launch nm...")
	if launchArgs == "" {
		cmd := exec.Command(programPath)
		if err := cmd.Start(); err != nil {
			panic(err)
		}
	} else {
		cmd := exec.Command(programPath, launchArgs)
		if err := cmd.Start(); err != nil {
			panic(err)
		}
	}
}