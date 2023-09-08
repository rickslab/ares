package mock

import (
	"github.com/rickslab/ares/store"
	"github.com/rickslab/ares/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitMysqlCli() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("./mock.db"), &gorm.Config{})
	util.AssertError(err)

	store.SetMySQL("write", db)
	store.SetMySQL("read", db)
	return db
}

func InitModel(models ...any) {
	db := store.MySQL("write")

	db.Migrator().DropTable(models)
	db.Migrator().CreateTable(models)
}
