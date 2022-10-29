package settings

import (
	"gopkg.in/ini.v1"
)

// 统一定义设置变量
var (
	Cfg           *ini.File
	ListenAddress string
	StorageRoot   string
)

func init() {
	source := "conf/app.ini"
	Cfg, err := ini.Load(source)
	if err != nil {
		panic(err)
	}
	ListenAddress = Cfg.Section("server").Key("LISTEN_ADDRESS").MustString(":8080")
	StorageRoot = Cfg.Section("server").Key("STORAGE_ROOT").MustString("./data")
}
