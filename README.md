# BetterIOC

BetterIOC is a reflection-based dependency injection container for Go. It uses
reflection to resolve dependencies from constructor function parameters at runtime.

## Example

This example registers a database and an application service. BetterIOC sees
that newItemService needs a *sqlx.DB and supplies the registered database
when it creates the service.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessola/better-ioc"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Item struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type ItemService struct {
	db *sqlx.DB
}

func newItemService(db *sqlx.DB) *ItemService {
	return &ItemService{db: db}
}

func (s *ItemService) ListItems() ([]Item, error) {
	var items []Item
	err := s.db.Select(&items, "SELECT id, name FROM items ORDER BY id")
	return items, err
}

func main() {
	c := ioc.NewContainer()
	c.AddSingleton[*sqlx.DB](func() *sqlx.DB {
		return sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	})
	c.AddScoped[*ItemService](newItemService)

	service := c.MustGet[*ItemService]()

	items, err := service.ListItems()
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		fmt.Printf("%d: %s\n", item.ID, item.Name)
	}
}
```

The example uses PostgreSQL through `sqlx`; install `github.com/jmoiron/sqlx`
and a PostgreSQL driver such as `github.com/lib/pq` in your application.

## Interfaces

Providers can be registered against interface types. This example exposes an
AppService as an ItemLister, keeping callers dependent only on the ListItems
behavior they need. AppService also has other capabilities, such as
ApplicationName, that are not part of ItemLister.

~~~go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessola/better-ioc"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Item struct {
	ID   int
	Name string
}

type ItemLister interface {
	ListItems() ([]Item, error)
}

type AppService struct {
	db *sqlx.DB
}

func newAppService(db *sqlx.DB) *AppService {
	return &AppService{db: db}
}

func (s *AppService) ListItems() ([]Item, error) {
	var items []Item
	err := s.db.Select(&items, "SELECT id, name FROM items ORDER BY id")
	return items, err
}

func (s *AppService) ApplicationName() string {
	return "Inventory"
}

func main() {
	c := ioc.NewContainer()
	c.AddSingleton[*sqlx.DB](func() *sqlx.DB {
		return sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	})
	c.AddScoped[*AppService](newAppService)
	c.AddScoped[ItemLister](func(app *AppService) ItemLister {
		return app
	})

	lister := c.MustGet[ItemLister]()

	items, err := lister.ListItems()
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		fmt.Printf("%d: %s\n", item.ID, item.Name)
	}
}
~~~
