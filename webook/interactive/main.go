package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"net/http"
)

func main() {
	initViper()
	app := InitApp()
	initPrometheus()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	// 当前工作目录的 config 子目录
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}
func initViperV1() {
	cfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")
	// 这一步之后，cfile 里面才有值
	pflag.Parse()
	// 默认值V1
	//viper.Set("db.dsn", "localhost:3306")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}

func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}
