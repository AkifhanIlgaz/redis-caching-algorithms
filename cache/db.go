package cache

type User struct {
	Id   string `json:"id" redis:"-"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var myDB = map[string]User{
	"1": {Id: "1", Name: "Alice", Age: 30},
	"2": {Id: "2", Name: "Bob", Age: 25},
	"3": {Id: "3", Name: "Charlie", Age: 35},
	"4": {Id: "4", Name: "Zozak", Age: 1},
	"5": {Id: "5", Name: "Enayi", Age: 35},
}

func getUserFromDb(id string) User {
	return myDB[id]
}
