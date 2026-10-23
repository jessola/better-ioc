# BetterIOC

BetterIOC is a reflection-based dependency injection container for Go. It uses
reflection to resolve dependencies from constructor function parameters at runtime.

> **Status:** Experimental pre-release (`v0.4.0`). Public APIs may change
> without notice before a stable release.

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

MakeScopedFunc turns an ordinary function into a framework callback while
creating a scope and registering callback arguments for each invocation.

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

func listItems(
	lister ItemLister,
	w http.ResponseWriter,
) {
	items, err := lister.ListItems()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Package-level helpers are shorthand for DefaultContainer methods.
	// ioc.AddSingleton(...) is equivalent to
	// ioc.DefaultContainer.AddSingleton(...).
	ioc.AddSingleton[*sqlx.DB](func() *sqlx.DB {
		return sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	})
	ioc.AddScoped[*AppService](newAppService)
	ioc.BindScoped[ItemLister, *AppService]()

	http.HandleFunc(
		"/items",
		ioc.MustMakeScopedFunc[http.HandlerFunc](
			ioc.DefaultContainer,
			listItems,
		),
	)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Watermill consumers

MakeScopedFunc also adapts plain functions to Watermill consumers. Router and
subscriber construction are omitted; any Watermill subscriber can be passed to
registerConsumers. Registering multiple transient ConsumerHandler providers
lets the container collect every consumer configuration as a slice.

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
			Run: ioc.MustMakeScopedFunc[message.NoPublishHandlerFunc](
				ioc.DefaultContainer,
				handleItemCreated,
			),
		}
	})
	c.AddTransient[ConsumerHandler](func(subscriber message.Subscriber) ConsumerHandler {
		return ConsumerHandler{
			Name:       "audit_item_created",
			Topic:      "items.created.audit",
			Subscriber: subscriber,
			Run: ioc.MustMakeScopedFunc[message.NoPublishHandlerFunc](
				ioc.DefaultContainer,
				auditItemCreated,
			),
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

## Dispatching application requests

The dispatch package is useful when different transports represent the same
application action. Define the action once as a request and handler function;
HTTP and message consumers then only translate their input into that request.
This example builds on the database and handler/consumer adapters above.

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jessola/better-ioc"
	"github.com/jessola/better-ioc/dispatch"
	"github.com/jmoiron/sqlx"
)

type CreateItem struct {
	Name string
}

type CreateItemResult struct {
	ID int
}

func createItem(
	request CreateItem,
	db *sqlx.DB,
) (CreateItemResult, error) {
	var result CreateItemResult
	err := db.QueryRowx(
		"INSERT INTO items (name) VALUES ($1) RETURNING id",
		request.Name,
	).Scan(&result.ID)
	return result, err
}

func createItemHTTP(
	w http.ResponseWriter,
	r *http.Request,
	dispatcher *dispatch.Dispatcher,
) {
	var request CreateItem
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := dispatcher.Send[CreateItemResult](request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createItemMessage(
	msg *message.Message,
	dispatcher *dispatch.Dispatcher,
) error {
	var request CreateItem
	if err := json.Unmarshal(msg.Payload, &request); err != nil {
		return err
	}

	_, err := dispatcher.Send[CreateItemResult](request)
	return err
}

// During application startup, after constructing router and subscriber:
c := ioc.DefaultContainer
dispatch.AddDispatcher(c)
dispatch.AddHandlerFunc[CreateItem, CreateItemResult](c, createItem)

http.HandleFunc(
	"POST /items",
	ioc.MustMakeScopedFunc[http.HandlerFunc](
		ioc.DefaultContainer,
		createItemHTTP,
	),
)
router.AddConsumerHandler(
	"create_item",
	"items.create",
	subscriber,
	ioc.MustMakeScopedFunc[message.NoPublishHandlerFunc](
		ioc.DefaultContainer,
		createItemMessage,
	),
)
```

## Advanced: custom callback scopes

MakeScopedFunc is the usual choice for framework callbacks. When an invocation
needs custom scope setup—such as a transaction—create and close the scope in a
wrapper, then use MakeFuncWithScope. It automatically registers the framework
callback arguments in your supplied scope.

```go
func transactionalHandler(fn any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := ioc.DefaultContainer.CreateScope()
		defer s.Close()

		// Register transaction-specific providers on s.
		// MakeFuncWithScope registers w and r before resolving fn.
		ioc.MustMakeFuncWithScope[http.HandlerFunc](s, fn)(w, r)
	}
}
```

MakeFuncWithScope does not create or close the supplied scope; its caller owns
that lifetime. Its generic function type allows wrappers for any callback
shape, not only HTTP handlers.

## Middleware

WithMiddleware is a low-level composition primitive for callbacks that need to
participate in the current dependency scope. It is most useful when middleware
creates or replaces a dependency for the next callback—for example, a database
transaction, authenticated identity, or request-local configuration. Ordinary
logging, recovery, and other framework concerns are usually better served by
the framework's native middleware.

Middleware receives and returns `any` because Go has no catch-all function
type. An application chooses a callback shape and keeps every middleware
compatible with it. For an Echo application, using handlers that return an
`error` makes `MustCallErr` a natural way to invoke the next callback.

```go
func transactional(next any) any {
	return func(c *ioc.Container, db *sqlx.DB, ctx echo.Context) error {
		tx, err := db.BeginTxx(ctx.Request().Context(), nil)
		if err != nil {
			return err // The final handler is not called.
		}

		c.AddScoped[*sqlx.Tx](ioc.Value(tx))
		if err := ioc.MustCallErr(c, next); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit() // This is perhaps best left to the actual handler!
	}
}

func scopedHandler(fn any, middlewares ...ioc.MiddlewareFunc) echo.HandlerFunc {
	handler := ioc.Must(ioc.WithMiddleware(fn, middlewares...))
	return ioc.MustMakeScopedFunc[echo.HandlerFunc](ioc.DefaultContainer, handler)
}
```

Register a transactional route during application setup:

```go
e.POST("/items", scopedHandler(createItem, transactional))
```

MakeScopedFunc creates a fresh scope for each request. The transactional
middleware receives that scope, registers the transaction in it, and invokes
the next callback. A final handler can therefore depend on `*sqlx.Tx` without
knowing how the transaction was created or finalized.

### Manual invocation and overrides

Use MustCall when you want to resolve and invoke a function immediately rather
than adapt it into another function type. Create a scope manually when an
invocation needs custom lifetime management or conditional registrations.

```go
s := ioc.CreateScope()
defer s.Close()

s.AddScoped[*http.Request](ioc.Value(r))
s.AddScoped[http.ResponseWriter](ioc.Value(w))

// Invoke fn after resolving its dependencies from s.
err := ioc.MustCallErr(
	s,
	ioc.WithOverrides(
		fn,
		ioc.ProvideNamed[ItemLister](0, "read-only"),
	),
)
```

Overrides select a function parameter by its zero-based index. Use Provide for
ordinary type resolution, ProvideNamed to select a named registration, and
ProvideValue to supply a specific value. Use WithOverrides to attach overrides
before passing a function to registration methods, MakeScopedFunc,
MakeFuncWithScope, or the Call helpers.
