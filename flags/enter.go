package flags

import (
	"flag"
	"fmt"
	"os"
)

type Option struct {
	File    string //配置文件
	DB      bool   //迁移数据库
	Version bool   //查看版本
	Type    string
	Sub     string
	ES      bool //建立索引
}

var FlagOptions = new(Option) //一个指针命令行参数指针

func Parse() { //注册然后解析
	flag.StringVar(&FlagOptions.File, "f", "setting.yaml", "配置文件")
	flag.BoolVar(&FlagOptions.DB, "db", false, "数据库迁移")
	flag.BoolVar(&FlagOptions.ES, "es", false, "es建立索引")
	flag.BoolVar(&FlagOptions.Version, "v", false, "版本")
	flag.StringVar(&FlagOptions.Type, "t", "", "操作类型")
	flag.StringVar(&FlagOptions.Sub, "s", "", "子类型/内容")
	flag.Parse() //解析部分
}

func Run() {
	if FlagOptions.DB {
		//执行数据库迁移
		FlagDB()
		os.Exit(0)
	}
	if FlagOptions.ES {
		EsIndex()
		os.Exit(0)
	}
	if FlagOptions.Version {
		fmt.Print("当前版本为:")
		fmt.Println(1)
		os.Exit(0)
	}
	switch FlagOptions.Type {
	case "user":
		switch FlagOptions.Sub {
		case "create":
			Create()
			os.Exit(0)
		}
	}
}
