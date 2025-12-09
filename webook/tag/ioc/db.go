package ioc

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	gromProm "gorm.io/plugin/prometheus"
	prometheus2 "mbook/webook/pkg/gormx"
	"mbook/webook/pkg/logger"
	"mbook/webook/tag/repository/dao"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	c := Config{
		DSN: "root:root@tcp(localhost:3306)/mysql",
	}
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(fmt.Errorf("初始化配置失败 %v1, 原因 %w", c, err))
	}
	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		// 使用 DEBUG 来打印
		//Logger: glogger.New(gormLoggerFunc(l.Debug),
		//	glogger.Config{
		//		SlowThreshold: 0,
		//		LogLevel:      glogger.Info,
		//	}),
	})
	if err != nil {
		panic(err)
	}

	// 接入 prometheus
	err = db.Use(gromProm.New(gromProm.Config{
		DBName: "webook",
		// 每 15 秒采集一些数据
		RefreshInterval: 15,
		MetricsCollector: []gromProm.MetricsCollector{
			&gromProm.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		}, // user defined metrics
	}))
	if err != nil {
		panic(err)
	}
	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics()))
	if err != nil {
		panic(err)
	}

	cb := prometheus2.NewCallbacks(prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "gorm",
		Help:      "gorm DB 查询",
		ConstLabels: map[string]string{
			"instance_id": "my-instance-1",
		},

		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	err = cb.Initialize(db)
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Val: args})
}
