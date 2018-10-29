package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"
	//"reflect"
	//"strings"

	"github.com/jinzhu/gorm"
)

import _ "SAP/go-hdb/driver"

//import _ "github.com/go-sql-driver/mysql"

var DB *gorm.DB

type Num int64

type CreditCard struct {
	ID        int8
	Number    string
	UserId    sql.NullInt64
	CreatedAt time.Time `sql:"not null"`
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"column:deleted_time"`
}

type Email struct {
	Id        int16
	UserId    int
	Email     string `sql:"type:varchar(100);"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Address struct {
	ID        int
	Address1  string
	Address2  string
	Post      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Language struct {
	gorm.Model
	Name  string
	Users []User `gorm:"many2many:user_languages;"`
}

type Role struct {
	Name string `gorm:"size:256"`
}

type Company struct {
	Id    int64
	Name  string
	Owner *User `sql:"-"`
}

type EncryptedData []byte

type User struct {
	Id  int64
	Age int64
	//UserNum           Num
	Name              string `sql:"size:255"`
	Email             string
	Birthday          *time.Time    // Time
	CreatedAt         time.Time     // CreatedAt: Time of record is created, will be insert automatically
	UpdatedAt         time.Time     // UpdatedAt: Time of record is updated, will be updated automatically
	Emails            []Email       // Embedded structs
	BillingAddress    Address       // Embedded struct
	BillingAddressID  sql.NullInt64 // Embedded struct's foreign key
	ShippingAddress   Address       // Embedded struct
	ShippingAddressId int64         // Embedded struct's foreign key
	CreditCard        CreditCard
	Latitude          float64
	Languages         []Language `gorm:"many2many:user_languages;"`
	CompanyID         *int
	Company           Company
	Role              Role
	/*Password          EncryptedData
	PasswordHash      []byte
	IgnoreMe          int64                 `sql:"-"`
	IgnoreStringSlice []string              `sql:"-"`
	Ignored           struct{ Name string } `sql:"-"`
	IgnoredPointer    *User                 `sql:"-"`*/
}

func getPreparedUser(name string, role string) *User {
	var company Company
	DB.Where(Company{Name: role}).FirstOrCreate(&company)

	return &User{
		Name:            name,
		Age:             20,
		Role:            Role{role},
		BillingAddress:  Address{Address1: fmt.Sprintf("Billing Address %v", name)},
		ShippingAddress: Address{Address1: fmt.Sprintf("Shipping Address %v", name)},
		CreditCard:      CreditCard{Number: fmt.Sprintf("123456%v", name)},
		Emails: []Email{
			{Email: fmt.Sprintf("user_%v@example1.com", name)}, {Email: fmt.Sprintf("user_%v@example2.com", name)},
		},
		Company: company,
		Languages: []Language{
			{Name: fmt.Sprintf("lang_1_%v", name)},
			{Name: fmt.Sprintf("lang_2_%v", name)},
		},
	}
}

func getPreloadUser(name string) *User {
	return getPreparedUser(name, "Preload")
}

func checkUserHasPreloadData(user User) {
	u := getPreloadUser(user.Name)
	if user.BillingAddress.Address1 != u.BillingAddress.Address1 {
		fmt.Println("Failed to preload user's BillingAddress")
	}

	if user.ShippingAddress.Address1 != u.ShippingAddress.Address1 {
		fmt.Println("Failed to preload user's ShippingAddress")
	}

	if user.CreditCard.Number != u.CreditCard.Number {
		fmt.Println("Failed to preload user's CreditCard")
	}

	if user.Company.Name != u.Company.Name {
		fmt.Println("Failed to preload user's Company")
	}

	if len(user.Emails) != len(u.Emails) {
		fmt.Println("Failed to preload user's Emails")
	} else {
		var found int
		for _, e1 := range u.Emails {
			for _, e2 := range user.Emails {
				if e1.Email == e2.Email {
					found++
					break
				}
			}
		}
		if found != len(u.Emails) {
			fmt.Println("Failed to preload user's email details")
		}
	}
}

func main() {
	var DSN string

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	flag.StringVar(&DSN, "dsn", "m98://SYSTEM:Sybase123@10.58.180.202:30215?CURRENTSCHEMA=ORM_TEST", "hdb dsn")

	DB, _ = gorm.Open("hdb", DSN)

	defer DB.Close()

	DB.LogMode(true)
	DB.SingularTable(true)

	values := []interface{}{&Company{}, &Email{}, &Address{}, &CreditCard{}, &Language{}, &Role{}, &User{}}
	for _, value := range values {
		DB.DropTableIfExists(value)
	}

	if err := DB.AutoMigrate(values...).Error; err != nil {
		panic(fmt.Sprintf("No error should happen when create table, but got %+v\n", err))
	}

	user1 := getPreloadUser("user1")
	//fmt.Printf("user1: %v\n", user1)
	DB.Save(user1)

	preloadDB := DB.Where("name = ?", "user1").Preload("BillingAddress").Preload("ShippingAddress").
		Preload("CreditCard").Preload("Emails").Preload("Company")
	var user User
	preloadDB.Find(&user)
	checkUserHasPreloadData(user)

	user2 := getPreloadUser("user2")
	DB.Save(user2)

	user3 := getPreloadUser("user3")
	DB.Save(user3)

	var users []User
	preloadDB.Find(&users)

	for _, user := range users {
		checkUserHasPreloadData(user)
	}

	var users2 []*User
	preloadDB.Find(&users2)

	for _, user := range users2 {
		checkUserHasPreloadData(*user)
	}

	var users3 []*User
	preloadDB.Preload("Emails", "email = ?", user3.Emails[0].Email).Find(&users3)

	for _, user := range users3 {
		if user.Name == user3.Name {
			if len(user.Emails) != 1 {
				fmt.Println("should only preload one emails for user3 when with condition")
			}
		} else if len(user.Emails) != 0 {
			fmt.Println("should not preload any emails for other users when with condition")
		} else if user.Emails == nil {
			fmt.Println("should return an empty slice to indicate zero results")
		}
	}
}
