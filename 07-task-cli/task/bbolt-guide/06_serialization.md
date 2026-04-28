# bbolt Guide — Serialization

## The Problem

bbolt only stores `[]byte`. Your tasks are structs. You need a way to convert between them.

```go
type Task struct {
    ID          uint64
    Description string
    CreatedAt   time.Time
    Done        bool
}
```

This needs to become bytes to go in, and be reconstructed from bytes when it comes out.

This process is called **serialization** (encoding to bytes) and **deserialization** (decoding from bytes).

---

## Option 1: JSON (Recommended for Beginners)

JSON is the easiest to get started with and has a huge advantage: **you can read the raw bytes in the database file** if you ever need to inspect or debug it.

```go
import "encoding/json"

// Encoding a struct → bytes (before Put)
task := Task{ID: 1, Description: "Walk the dog", CreatedAt: time.Now()}
encoded, err := json.Marshal(task)
if err != nil {
    return err
}
b.Put(itob(task.ID), encoded)

// Decoding bytes → struct (after Get)
v := b.Get(itob(1))
var task Task
err = json.Unmarshal(v, &task)
```

### JSON Field Control

By default, `json.Marshal` uses the struct field names. You can customize with struct tags:

```go
type Task struct {
    ID          uint64    `json:"id"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    Done        bool      `json:"done,omitempty"` // omitted if false
}
```

### JSON Gotchas

- Unexported fields (lowercase) are **ignored** by encoding/json
- `time.Time` marshals to RFC3339 string format — works fine, just be aware
- JSON is slightly verbose; for a task manager, that's totally fine

---

## Option 2: encoding/gob

Go's built-in binary format. More efficient than JSON but produces binary output you can't easily read.

```go
import (
    "bytes"
    "encoding/gob"
)

// Encoding
var buf bytes.Buffer
enc := gob.NewEncoder(&buf)
if err := enc.Encode(task); err != nil {
    return err
}
b.Put(itob(task.ID), buf.Bytes())

// Decoding
buf := bytes.NewReader(v)
dec := gob.NewDecoder(buf)
var task Task
if err := dec.Decode(&task); err != nil {
    return err
}
```

Gob is faster and more compact than JSON for large datasets. For a task manager, the difference is negligible — use JSON unless you have a specific reason not to.

---

## Wrapping Serialization in Helper Functions

The cleanest approach is to keep serialization out of your database functions entirely:

```go
func marshalTask(t Task) ([]byte, error) {
    return json.Marshal(t)
}

func unmarshalTask(data []byte) (Task, error) {
    var t Task
    err := json.Unmarshal(data, &t)
    return t, err
}
```

Then your database functions become very readable:

```go
func addTask(db *bolt.DB, description string) error {
    return db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("Tasks"))

        id, err := b.NextSequence()
        if err != nil {
            return err
        }

        task := Task{
            ID:          id,
            Description: description,
            CreatedAt:   time.Now(),
        }

        encoded, err := marshalTask(task)
        if err != nil {
            return err
        }

        return b.Put(itob(id), encoded)
    })
}
```

---

## What Goes in the Key vs the Value?

A common question: if the Task struct has an ID field, should the key just be the ID, or should you also store the ID inside the JSON value?

**Store the ID in both places.**

- The **key** is what bbolt uses to find and sort your data
- The **value** (your JSON) should be self-contained so that when you decode it, you have everything you need without having to look at the key separately

```go
// Key: itob(task.ID)   ← used for lookup and ordering
// Value: json(task)    ← includes ID field too, for convenience when decoding
```

When you iterate with a cursor and unmarshal each value, you get back a fully populated `Task` struct including the ID — you don't need to also decode the key.
