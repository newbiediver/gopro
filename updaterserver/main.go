package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"github.com/newbiediver/gopro/updaterserver/config"
)

func main() {
	sc := config.GetSystemConfig()
	sc.DefaultConfig()
	yaml, isLoaded := sc.LoadYaml("server.yml")
	if isLoaded {
		err := sc.SetConfig(yaml)
		if err != nil {
			panic(err)
		}
	}

	contents := GetContents()
	contents.LoadContents(sc.ContentsPath)

	r := setupUri()
	strPort := fmt.Sprintf(":%d", sc.Port)

	log.Println("Start update service...")
	_ = r.Run(strPort)
}

func setupUri() *gin.Engine {
	sc := config.GetSystemConfig()
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.MaxMultipartMemory = 2147483648

	r.GET("/list", onList)
	r.GET("/reload", onReload)
	r.StaticFS("/contents", http.Dir(sc.ContentsPath))

	return r
}

func onList(ctx *gin.Context) {
	updateList := new(UpdateList)
	contents := GetContents()
	updateList.MakeList(contents)

	ctx.JSON(http.StatusOK, updateList)
}

func onReload(ctx *gin.Context) {
	contents := GetContents()
	contents.Reload()

	ctx.String(http.StatusOK, "Reloaded!!")
}