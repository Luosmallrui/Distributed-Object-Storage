package db

import (
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func init() {
	//var dsn string
	//env := os.Getenv("USER")
	//if env == "root" {
	//	dsn = "root:123456@tcp(127.0.0.1:3306)/xfzddataV2?charset=utf8mb4&parseTime=True&loc=Local"
	//} else {
	//	dsn = "root:123456@tcp(0.0.0.0:3306)/xfzddataV2?charset=utf8mb4&parseTime=True&loc=Local"
	//}
	//var err error
	//db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//if err != nil {
	//	fmt.Println(1)
	//}
	//sql, _ := db.DB()
	//sql.SetConnMaxLifetime(60 * time.Minute)
	//sql.SetMaxOpenConns(100)
	//sql.SetMaxIdleConns(50)
}

func Db() *gorm.DB {
	return db
}
