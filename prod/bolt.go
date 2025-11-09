package prod

import (
	"go.etcd.io/bbolt"
)

func main() {
	db, err := bbolt.Open("mydb.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}
