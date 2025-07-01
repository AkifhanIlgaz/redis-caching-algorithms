package cache

import "fmt"

type User struct {
	Id   string `json:"id" redis:"-"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var myDB = map[string]User{
	"1": {Name: "Alice", Age: 30, Id: "1"},
	"2": {Name: "Bob", Age: 25, Id: "2"},
	"3": {Name: "Charlie", Age: 35, Id: "3"},
	"4": {Name: "Zozak", Age: 1, Id: "4"},
	"5": {Name: "Enayi", Age: 35, Id: "5"},
}

func getUserFromDb(id string) *User {
	fmt.Println("Getting user from database")
	user := myDB[id]
	return &user
}
