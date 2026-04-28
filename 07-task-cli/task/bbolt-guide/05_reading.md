# bbolt Guide — Reading Data

## The Two Read Operations

| Method | What it does |
|---|---|
| `b.Get(key)` | Fetch a single value by key |
| `b.Cursor()` | Iterate over key-value pairs in sorted order |

Both work inside `db.View` (and also inside `db.Update` if needed).

---

## Get — Fetching a Single Item

```go
err := db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))

    v := b.Get([]byte("my-key"))
    if v == nil {
        fmt.Println("Key not found")
        return nil
    }

    fmt.Println(string(v))
    return nil
})
```

- Returns `nil` if the key doesn't exist
- Returns `[]byte` if found

### ⚠️ Critical: Byte Slice Lifetime

The `[]byte` returned by `Get` is **only valid for the duration of the transaction**. The moment `db.View` returns, that memory may be reused or freed.

```go
// WRONG — using v after the transaction closes
var v []byte
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    v = b.Get([]byte("key"))  // v points into bbolt's memory
    return nil
})
fmt.Println(string(v)) // DANGEROUS — memory may be invalid
```

```go
// RIGHT — copy the bytes before the transaction closes
var result []byte
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    v := b.Get([]byte("key"))
    if v != nil {
        result = make([]byte, len(v))
        copy(result, v)
    }
    return nil
})
fmt.Println(string(result)) // safe
```

In practice, if you immediately unmarshal `v` into a struct inside the transaction, you don't need to copy — the struct owns its own memory. Copying is only needed if you're storing the raw `[]byte` to use later.

---

## Cursor — Iterating Over All Items

To list every key-value pair in a bucket, use a cursor:

```go
err := db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))

    c := b.Cursor()
    for k, v := c.First(); k != nil; k, v = c.Next() {
        fmt.Printf("key=%v, value=%s\n", k, v)
    }

    return nil
})
```

### How the Cursor Works

| Method | Moves to |
|---|---|
| `c.First()` | The first key-value pair |
| `c.Last()` | The last key-value pair |
| `c.Next()` | The next key-value pair |
| `c.Prev()` | The previous key-value pair |
| `c.Seek(key)` | The key >= the given key |

Each method returns `(key []byte, value []byte)`. When `key` is `nil`, you've gone past the end (or the bucket is empty).

Items are returned in **sorted byte order by key** — which is why using big-endian integers for keys means you get tasks in numeric order (`1, 2, 3, ...`).

---

## A Complete List Tasks Example

```go
type Task struct {
    ID          uint64
    Description string
}

func listTasks(db *bolt.DB) ([]Task, error) {
    var tasks []Task

    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("Tasks"))

        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            var task Task
            if err := json.Unmarshal(v, &task); err != nil {
                return err
            }
            tasks = append(tasks, task)
        }

        return nil
    })

    return tasks, err
}
```

Notice that `tasks` is declared outside the transaction, and we append to it inside. After the transaction closes, the slice is fully populated and safe to use.

---

## A Complete Get Single Task Example

```go
func getTask(db *bolt.DB, id uint64) (Task, error) {
    var task Task

    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("Tasks"))

        v := b.Get(itob(id))
        if v == nil {
            return fmt.Errorf("task %d not found", id)
        }

        return json.Unmarshal(v, &task)
        // Safe: json.Unmarshal copies the data into task's fields
    })

    return task, err
}
```

---

## Counting Items

bbolt doesn't have a `Count()` method. The idiomatic way is to use bucket stats:

```go
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    stats := b.Stats()
    fmt.Printf("Number of tasks: %d\n", stats.KeyN)
    return nil
})
```

Or just count while iterating if you need to do both at once.
