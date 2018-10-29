package main

import (
	//"database/sql"
	"flag"
	"fmt"
	//"reflect"
	//"strings"

	"github.com/jinzhu/gorm"
)

import _ "SAP/go-hdb/driver"

type User struct {
	gorm.Model
	Profile   Profile //`gorm:"ForeignKey:ProfileID"`
	ProfileID int
}

type Profile struct {
	gorm.Model
	Name string
}

func main() {
	var DSN string

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	flag.StringVar(&DSN, "dsn", "m98://SYSTEM:Sybase123@10.58.180.202:30215?CURRENTSCHEMA=ORM_TEST", "hdb dsn")
	DB, err := gorm.Open("hdb", DSN)

	//flag.StringVar(&DSN, "dsn", "root:Sybase123@tcp(127.0.0.1:3306)/world?charset=utf8&parseTime=True&loc=Local", "mysql dsn")
	//DB, err := gorm.Open("mysql", DSN)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success to connect database")
	}

	defer DB.Close()

	DB.LogMode(true)
	DB.SingularTable(true)

	values := []interface{}{&Profile{}, &User{}}
	for _, value := range values {
		DB.DropTableIfExists(value)
	}

	if err := DB.AutoMigrate(values...).Error; err != nil {
		panic(fmt.Sprintf("No error should happen when create table, but got %+v\n", err))
	}

	user := User{
		Profile: Profile{
			Name: "test",
		},
	}
	DB.Create(&user)

	if DB.Model(&user).Association("Profile").Count() != 1 {
		fmt.Println("User's profile count should be 1")
	}

	// AddForeignKey
	// SELECT COUNT(*) FROM REFERENTIAL_CONSTRAINTS WHERE SCHEMA_NAME=CURRENT_SCHEMA AND TABLE_NAME='USER' AND COLUMN_NAME='PROFILE_ID'
	// DB.Model(&User{}).AddForeignKey("PROFILE_ID", "profile(ID)", "CASCADE", "CASCADE")
}
