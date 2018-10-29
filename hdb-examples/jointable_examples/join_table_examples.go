package main

import (
	//"database/sql"
	"flag"
	"fmt"
	//"reflect"
	//"strings"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

import _ "SAP/go-hdb/driver"

//import _ "github.com/go-sql-driver/mysql"

type Address struct {
	ID        int
	Address1  string
	Address2  string
	Post      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Person struct {
	Id        int
	Name      string
	Addresses []*Address `gorm:"many2many:person_address;"`
}

type PersonAddress struct {
	gorm.JoinTableHandler
	PersonID  int
	AddressID int
	DeletedAt *time.Time
	CreatedAt time.Time
}

func (*PersonAddress) Add(handler gorm.JoinTableHandlerInterface, db *gorm.DB, foreignValue interface{}, associationValue interface{}) error {
	foreignPrimaryKey, _ := strconv.Atoi(fmt.Sprint(db.NewScope(foreignValue).PrimaryKeyValue()))
	associationPrimaryKey, _ := strconv.Atoi(fmt.Sprint(db.NewScope(associationValue).PrimaryKeyValue()))
	if result := db.Unscoped().Model(&PersonAddress{}).Where(map[string]interface{}{
		"person_id":  foreignPrimaryKey,
		"address_id": associationPrimaryKey,
	}).Update(map[string]interface{}{
		"person_id":  foreignPrimaryKey,
		"address_id": associationPrimaryKey,
		"deleted_at": gorm.Expr("NULL"),
	}).RowsAffected; result == 0 {
		return db.Create(&PersonAddress{
			PersonID:  foreignPrimaryKey,
			AddressID: associationPrimaryKey,
		}).Error
	}

	return nil
}

func (*PersonAddress) Delete(handler gorm.JoinTableHandlerInterface, db *gorm.DB, sources ...interface{}) error {
	return db.Delete(&PersonAddress{}).Error
}

func (pa *PersonAddress) JoinWith(handler gorm.JoinTableHandlerInterface, db *gorm.DB, source interface{}) *gorm.DB {
	table := pa.Table(db)
	return db.Joins("INNER JOIN person_address ON person_address.address_id = address.id").Where(fmt.Sprintf("%v.deleted_at IS NULL OR %v.deleted_at <= '0001-01-02'", table, table))
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

	values := []interface{}{&Address{}, &Person{}}
	for _, value := range values {
		DB.DropTableIfExists(value)
	}

	DB.DropTable("person_address")

	if err := DB.AutoMigrate(values...).Error; err != nil {
		panic(fmt.Sprintf("No error should happen when create table, but got %+v\n", err))
	}

	DB.SetJoinTableHandler(&Person{}, "Addresses", &PersonAddress{})

	address1 := &Address{Address1: "address 1"}
	address2 := &Address{Address1: "address 2"}
	person := &Person{Name: "person", Addresses: []*Address{address1, address2}}
	DB.Save(person)

	DB.Model(person).Association("Addresses").Delete(address1)

	if DB.Find(&[]PersonAddress{}, "person_id = ?", person.Id).RowsAffected != 1 {
		fmt.Println("Should found one address")
	}

	if DB.Model(person).Association("Addresses").Count() != 1 {
		fmt.Println("Should found one address")
	}

	if DB.Unscoped().Find(&[]PersonAddress{}, "person_id = ?", person.Id).RowsAffected != 2 {
		fmt.Println("Should found two addresses with Unscoped")
	}

	if DB.Model(person).Association("Addresses").Clear(); DB.Model(person).Association("Addresses").Count() != 0 {
		fmt.Println("Should deleted all address")
	}

}
