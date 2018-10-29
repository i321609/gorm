package main

import (
	"fmt"
	//"database/sql"
	"flag"
	"time"

	"github.com/jinzhu/gorm"
)

import _ "SAP/go-hdb/driver"

type Animal struct {
	Counter    uint64    `gorm:"primary_key:yes"`
	Name       string    `sql:"DEFAULT:'galeone'"`
	From       string    `sql:"column:from_where"`
	Age        time.Time `sql:"DEFAULT:current_timestamp"`
	unexported string    // unexported value
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Product struct {
	Id                    int64
	Code                  string
	Price                 int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
	AfterFindCallTimes    int64
	BeforeCreateCallTimes int64
	AfterCreateCallTimes  int64
	BeforeUpdateCallTimes int64
	AfterUpdateCallTimes  int64
	BeforeSaveCallTimes   int64
	AfterSaveCallTimes    int64
	BeforeDeleteCallTimes int64
	AfterDeleteCallTimes  int64
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

	values := []interface{}{&Product{}, &Animal{}}
	for _, value := range values {
		//DB.DropTable(value)
		DB.DropTableIfExists(value)
	}

	if err := DB.AutoMigrate(values...).Error; err != nil {
		panic(fmt.Sprintf("No error should happen when create table, but got %+v", err))
	}

	product1 := Product{Code: "product1code"}
	product2 := Product{Code: "product2code"}

	DB.Save(&product1).Save(&product2).Update("code", "product2newcode")

	if product2.Code != "product2newcode" {
		fmt.Printf("Record should be updated\n")
	}

	DB.First(&product1, product1.Id)
	DB.First(&product2, product2.Id)

	if DB.First(&Product{}, "code = ?", product1.Code).RecordNotFound() {
		fmt.Printf("Product1 should not be updated\n")
	}

	if !DB.First(&Product{}, "code = ?", "product2code").RecordNotFound() {
		fmt.Printf("Product2's code should be updated\n")
	}

	if DB.First(&Product{}, "code = ?", "product2newcode").RecordNotFound() {
		fmt.Printf("Product2's code should be updated\n")
	}

	updatedAt1 := product1.UpdatedAt

	DB.Table("product").Where("code in (?)", []string{"product1code"}).Update("code", "product1newcode")

	var product4 Product
	DB.First(&product4, product1.Id)
	if updatedAt1.Format(time.RFC3339Nano) != product4.UpdatedAt.Format(time.RFC3339Nano) {
		fmt.Printf("updatedAt should be updated if something changed\n")
	}

	if !DB.First(&Product{}, "code = 'product1code'").RecordNotFound() {
		fmt.Printf("Product1's code should be updated\n")
	}

	if DB.First(&Product{}, "code = 'product1newcode'").RecordNotFound() {
		fmt.Printf("Product should not be changed to 789\n")
	}

	if DB.Model(product2).Update("CreatedAt", time.Now().Add(time.Hour)).Error != nil {
		fmt.Printf("No error should raise when update with CamelCase\n")
	}

	if DB.Model(&product2).UpdateColumn("CreatedAt", time.Now().Add(time.Hour)).Error != nil {
		fmt.Printf("No error should raise when update_column with CamelCase\n")
	}

	var products []Product
	DB.Find(&products)

	if count := DB.Model(Product{}).Update("CreatedAt", time.Now().Add(2*time.Hour)).RowsAffected; count != int64(len(products)) {
		fmt.Printf("RowsAffected should be correct when do batch update\n")
	}

	DB.First(&product4, product4.Id)
	updatedAt4 := product4.UpdatedAt
	DB.Model(&product4).Update("price", gorm.Expr("price + ? - ?", 100, 50))
	var product5 Product
	DB.First(&product5, product4.Id)
	if product5.Price != product4.Price+100-50 {
		fmt.Println("Update with expression")
	}
	if product4.UpdatedAt.Format(time.RFC3339Nano) == updatedAt4.Format(time.RFC3339Nano) {
		fmt.Printf("Update with expression should update UpdatedAt\n")
	}

	animal := Animal{Name: "Ferdinand"}
	DB.Save(&animal)
	updatedAt5 := animal.UpdatedAt

	DB.Save(&animal).Update("name", "Francis")

	if updatedAt5.Format(time.RFC3339Nano) == animal.UpdatedAt.Format(time.RFC3339Nano) {
		fmt.Printf("updatedAt should not be updated if nothing changed\n")
	}

	var animals []Animal
	DB.Find(&animals)
	if count := DB.Model(Animal{}).Update("CreatedAt", time.Now().Add(2*time.Hour)).RowsAffected; count != int64(len(animals)) {
		fmt.Printf("RowsAffected should be correct when do batch update\n")
	}

	animal = Animal{From: "somewhere"}              // No name fields, should be filled with the default value (galeone)
	DB.Save(&animal).Update("From", "a nice place") // The name field shoul be untouched
	DB.First(&animal, animal.Counter)
	if animal.Name != "galeone" {
		fmt.Printf("Name fields shouldn't be changed if untouched, but got %v\n", animal.Name)
	}

	// When changing a field with a default value, the change must occur
	animal.Name = "amazing horse"
	DB.Save(&animal)
	DB.First(&animal, animal.Counter)
	if animal.Name != "amazing horse" {
		fmt.Printf("Update a filed with a default value should occur. But got %v\n", animal.Name)
	}

	// When changing a field with a default value with blank value
	animal.Name = ""
	DB.Save(&animal)
	DB.First(&animal, animal.Counter)
	if animal.Name != "" {
		fmt.Printf("Update a filed to blank with a default value should occur. But got %v\n", animal.Name)
	}
}
