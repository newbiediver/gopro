package main

import (
	"nm.go/go-server/tools/updaterserver/config"
	"strings"
)

type FileAttribute struct {
	FileName	string `json:"name"`
	FileSize	int64 `json:"size"`
	ModTimeYear	int `json:"modyear"`
	ModTimeMonth int `json:"modmonth"`
	ModTimeDay int `json:"modday"`
	ModTimeHour int `json:"modhour"`
	ModTimeMin int `json:"modmin"`
	ModTimeSec int `json:"modsec"`
	ModTimeNSec int `json:"modnsec"`
}

type FileList []FileAttribute

type UpdateList struct {
	TotalCnt	int `json:"count"`
	TotalSize	int64 `json:"size"`
	Files		FileList `json:"list"`
}

func (ul *UpdateList) MakeList(info *Contents) {
	sc := config.GetSystemConfig()
	ul.TotalCnt = len(info.AllContents)
	ul.Files = make(FileList, ul.TotalCnt)


	for i, v := range info.AllContents {
		ul.Files[i].FileName = v.FilePath
		ul.Files[i].FileSize = v.FileInfo.Size()
		ul.Files[i].ModTimeYear = v.FileInfo.ModTime().Year()
		ul.Files[i].ModTimeMonth = int(v.FileInfo.ModTime().Month())
		ul.Files[i].ModTimeDay = v.FileInfo.ModTime().Day()
		ul.Files[i].ModTimeHour = v.FileInfo.ModTime().Hour()
		ul.Files[i].ModTimeMin = v.FileInfo.ModTime().Minute()
		ul.Files[i].ModTimeSec = v.FileInfo.ModTime().Second()
		ul.Files[i].ModTimeNSec = v.FileInfo.ModTime().Nanosecond()

		strs := strings.Split(v.FilePath, sc.ContentsPath)
		ul.Files[i].FileName = strs[1]

		ul.TotalSize = ul.TotalSize + v.FileInfo.Size()
	}
}