package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/env"
	"github.com/rickslab/ares/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	mysqlClients = map[string]*gorm.DB{}
	mysqlMutex   = sync.RWMutex{}
)

func NewMysql(addr, username, password, db string) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, addr, db)
	cli, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	util.AssertError(err)

	sqlDB, err := cli.DB()
	util.AssertError(err)

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if env.IsTest() {
		return cli.Debug()
	}
	return cli
}

func MySQL(name string) *gorm.DB {
	cli := getMySQLCli(name)
	if cli != nil {
		return cli
	}
	return initMySQLCli(name)
}

func SetMySQL(name string, cli *gorm.DB) {
	mysqlClients[name] = cli
}

func initMySQLCli(name string) *gorm.DB {
	mysqlMutex.Lock()
	defer mysqlMutex.Unlock()

	cli, ok := mysqlClients[name]
	if ok {
		return cli
	}

	conf := config.YamlEnv().Sub(fmt.Sprintf("mysql.%s", name))

	cli = NewMysql(conf.GetString("address"), conf.GetString("username"), conf.GetString("password"), conf.GetString("db"))
	SetMySQL(name, cli)
	return cli
}

func getMySQLCli(name string) *gorm.DB {
	mysqlMutex.RLock()
	defer mysqlMutex.RUnlock()

	return mysqlClients[name]
}
