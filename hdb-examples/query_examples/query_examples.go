package main

import (
	"fmt"
	//"database/sql"
	"flag"
	"time"

	"github.com/jinzhu/gorm"
)

import _ "SAP/go-hdb/driver"

type Num int64

type Email struct {
	Id        int16
	UserId    int
	Email     string `sql:"type:varchar(100);"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Company struct {
	Id    int64
	Name  string
	Test  int   `sql:"-"`
	Owner *User `sql:"-"`
}

type User struct {
	Id        int64
	Age       int64
	UserNum   Num
	Name      string `sql:"size:255"`
	Email     string
	Birthday  *time.Time // Time
	CreatedAt time.Time  // CreatedAt: Time of record is created, will be insert automatically
	UpdatedAt time.Time  // UpdatedAt: Time of record is updated, will be updated automatically
	Emails    []Email    // Embedded structs
	/*BillingAddress    Address       // Embedded struct
	BillingAddressID  sql.NullInt64 // Embedded struct's foreign key
	ShippingAddress   Address       // Embedded struct
	ShippingAddressId int64         // Embedded struct's foreign key
	CreditCard        CreditCard
	Latitude          float64
	Languages         []Language `gorm:"many2many:user_languages;"`*/
	CompanyID *int
	Company   Company
	/*Role              Role
	Password          EncryptedData
	PasswordHash      []byte
	IgnoreMe          int64                 `sql:"-"`
	IgnoreStringSlice []string              `sql:"-"`
	Ignored           struct{ Name string } `sql:"-"`
	IgnoredPointer    *User                 `sql:"-"`*/
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

func main() {
	var DSN string

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	flag.StringVar(&DSN, "dsn", "m98://SYSTEM:Sybase123@10.58.180.202:30215?CURRENTSCHEMA=ORM_TEST", "hdb dsn")

	DB, err := gorm.Open("hdb", DSN)
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Println("Success to connect database")
	}

	defer DB.Close()

	DB.LogMode(true)
	DB.SingularTable(true)

	values := []interface{}{&User{}, &Email{}, &Company{}}
	for _, value := range values {
		//DB.DropTable(value)
		DB.DropTableIfExists(value)
	}

	if err := DB.AutoMigrate(values...).Error; err != nil {
		panic(fmt.Sprintf("No error should happen when create table, but got %+v", err))
	}

	DB.Save(&User{Name: "user1", Emails: []Email{{Email: "user1@example.com"}}})
	DB.Save(&User{Name: "user2", Emails: []Email{{Email: "user2@example.com"}}})

	var user1, user2, user3, user4 User
	DB.First(&user1)
	DB.Order("id").Limit(1).Find(&user2)

	ptrOfUser3 := &user3
	DB.Last(&ptrOfUser3)
	DB.Order("id desc").Limit(1).Find(&user4)
	if user1.Id != user2.Id || user3.Id != user4.Id {
		fmt.Printf("First and Last should by order by primary key")
	}

	var users []User
	DB.First(&users)
	if len(users) != 1 {
		fmt.Printf("Find first record as slice")
	}

	var user User
	if DB.Joins("left join email on email.user_id = user.id").First(&user).Error != nil {
		fmt.Printf("Should not raise any error when order with Join table")
	}

	if user.Email != "" {
		fmt.Printf("User's Email should be blank as no one set it")
	}

}
