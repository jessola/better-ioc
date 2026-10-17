# BetterIOC

BetterIOC is a reflection-based dependency injection container for Go. It uses
reflection to resolve dependencies from constructor function parameters at runtime.

## Example

This example registers a database and an application service. BetterIOC sees
that newItemService needs a \*sqlx.DB and supplies the registered database
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
	c.BindScoped[ItemLister, *AppService]()

	lister := c.MustGet[ItemLister]()

	items, err := lister.ListItems()
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		fmt.Printf("%d: %s\n", item.ID, item.Name)
	}
}
```

## HTTP handlers

The container can also encapsulate an application request. Here, handler turns
an ordinary function into an http.HandlerFunc: it creates a scope for each
request, registers the HTTP values as scoped dependencies, and resolves the
function's arguments in any order.

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func handler(fn any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := ioc.DefaultContainer.CreateScope()
		defer s.Close()

		s.AddScoped[http.ResponseWriter](ioc.Value(w))
		s.AddScoped[*http.Request](ioc.Value(r))

		err := ioc.MustCallErr(s, fn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func listItems(lister ItemLister, w http.ResponseWriter) error {
	items, err := lister.ListItems()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(items)
}

func main() {
	c := ioc.DefaultContainer
	c.AddSingleton[*sqlx.DB](func() *sqlx.DB {
		return sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	})
	c.AddScoped[*AppService](newAppService)
	c.BindScoped[ItemLister, *AppService]()

	http.HandleFunc("/items", handler(listItems))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Watermill consumers

The same request-scope pattern applies to message consumers. This example keeps
the Watermill adapter small: it scopes the incoming message and delegates to a
plain application function. Router and subscriber construction are omitted;
any Watermill subscriber can be passed to registerConsumers. Registering
multiple transient ConsumerHandler providers lets the container collect every
consumer configuration as a slice.

```go
package main

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jessola/better-ioc"
)

type ItemCreated struct {
	ID int
}

type ItemCreatedHandler interface {
	HandleItemCreated(ItemCreated) error
}

type ConsumerHandler struct {
	Name       string
	Topic      string
	Subscriber message.Subscriber
	Run        message.NoPublishHandlerFunc
}

type AppService struct{}

func newAppService() *AppService {
	return &AppService{}
}

func (s *AppService) HandleItemCreated(event ItemCreated) error {
	// Update application state for event.ID.
	return nil
}

func consumer(fn any) message.NoPublishHandlerFunc {
	return func(msg *message.Message) error {
		s := ioc.DefaultContainer.CreateScope()
		defer s.Close()

		s.AddScoped[*message.Message](ioc.Value(msg))

		return ioc.MustCallErr(s, fn)
	}
}

func handleItemCreated(
	msg *message.Message,
	handler ItemCreatedHandler,
) error {
	var event ItemCreated
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}

	return handler.HandleItemCreated(event)
}

func auditItemCreated(msg *message.Message) error {
	// Record an audit entry from msg.Payload.
	return nil
}

func registerConsumers(
	router *message.Router,
	subscriber message.Subscriber,
) {
	c := ioc.DefaultContainer
	c.AddSingleton[message.Subscriber](ioc.Value(subscriber))
	c.AddScoped[*AppService](newAppService)
	c.BindScoped[ItemCreatedHandler, *AppService]()

	c.AddTransient[ConsumerHandler](func(subscriber message.Subscriber) ConsumerHandler {
		return ConsumerHandler{
			Name:       "handle_item_created",
			Topic:      "items.created",
			Subscriber: subscriber,
			Run:        consumer(handleItemCreated),
		}
	})
	c.AddTransient[ConsumerHandler](func(subscriber message.Subscriber) ConsumerHandler {
		return ConsumerHandler{
			Name:       "audit_item_created",
			Topic:      "items.created.audit",
			Subscriber: subscriber,
			Run:        consumer(auditItemCreated),
		}
	})

	for _, handler := range c.MustGet[[]ConsumerHandler]() {
		router.AddConsumerHandler(
			handler.Name,
			handler.Topic,
			handler.Subscriber,
			handler.Run,
		)
	}
}
```
