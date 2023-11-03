package db_test

import (
	"fmt"
	"testing"

	"github.com/Drelf2018/webhook/service/data"
	"github.com/Drelf2018/webhook/service/db"
	"gorm.io/gorm/clause"
)

var Class = &db.DB{}

type Student struct {
	Name string `gorm:"primaryKey"`
	Male bool   `gorm:"column:gender"`
	Age  int

	FriendsName string    `gorm:"column:friends"`
	Friends     []Student `gorm:"foreignKey:FriendsName"`
}

func init() {
	Class.SetSqlite("./sqlite.db").AutoMigrate(&Student{})
	Class.Clauses(clause.OnConflict{DoNothing: true}).Create(&Student{
		Name: "m1dw1nter",
		Male: true,
		Age:  22,
		Friends: []Student{
			{Name: "FriendA", Male: false, Age: 16},
			{Name: "FriendB", Male: false, Age: 17},
			{Name: "FriendC", Male: false, Age: 18},
		},
	})
}

func TestDB(t *testing.T) {
	b := db.Exists[Student](Class, "gender = ?", true)
	if !b {
		t.Fatal(b)
	}

	var stu1 Student
	Class.First(&stu1, "age > ?", 18)
	if len(stu1.Friends) != 0 {
		t.Fatal(stu1)
	}

	var stu2 Student
	Class.Preload(&stu2, "name = ?", "m1dw1nter")
	if len(stu2.Friends) != 3 {
		t.Fatal(stu2)
	}

	var stus []Student
	Class.Preloads(&stus)
	if len(stus) != 4 {
		t.Fatal(stus)
	}
}

func TestRef(t *testing.T) {
	db.Ref.Init(data.Post{})
	fmt.Printf("db.Ref.Get(&[]data.Post{}): %v\n", db.Ref.Get(&[]*data.Post{}))
}
