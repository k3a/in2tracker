package main

import (
	"time"

	"k3a.me/money/backend/store"
	//"ndemiccreations.com/rest"
	_ "github.com/mattn/go-sqlite3" //sqlite driver
)

type User struct {
	ID           uint
	Name         string `gorm:"type:varchar(60)"`
	EMail        string `gorm:"not null;varchar(128);unique"`
	Password     string `gorm:"type:varchar(60)"`
	Created      time.Time
	LastLoggedIn time.Time
}

func main() {
	/*stor :=*/ store.New("sqlite3", "/tmp/qtest.db")

	//rest.ListenAndServe(":3434", nil)
}
