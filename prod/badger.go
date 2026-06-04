package prod

import (
	"github.com/dgraph-io/badger/v3"
)

func h() {
	opts := badger.DefaultOptions("badgerdb")
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}
