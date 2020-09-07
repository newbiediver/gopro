package main

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
