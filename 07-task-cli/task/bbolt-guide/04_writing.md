# bbolt Guide — Writing Data

## The Three Write Operations

| Method | What it does |
|---|---|
| `b.Put(key, value)` | Insert or overwrite a key-value pair |
| `b.Delete(key)` | Remove a key-value pair |
| `b.NextSequence()` | Get an auto-incrementing integer ID |

All three must be called inside a `db.Update` transaction.

---

## Put

```go
err := db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    return b.Put([]byte("my-key"), []byte("my-value"))
})
```

- If the key **doesn't exist**, it's created
- If the key **already exists**, the value is **overwritten** — no warning, no error
- Both key and value must be `[]byte`

---

## Delete

```go
err := db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    return b.Delete([]byte("my-key"))
})
```

- If the key **exists**, it's deleted
- If the key **doesn't exist**, nothing happens and `nil` is returned
- This means `Delete` won't tell you whether it actually removed anything — if you need to confirm the item existed, do a `Get` first inside the same transaction

---

## NextSequence — Auto-Incrementing IDs

For a task list, you need each task to have a unique numeric ID. bbolt's `NextSequence()` gives you this:

```go
err := db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))

    id, err := b.NextSequence()
    if err != nil {
        return err
    }

    // id is a uint64: 1, 2, 3, ... (never resets, never repeats)
    key := itob(id)
    return b.Put(key, []byte("Walk the dog"))
})
```

`NextSequence()` returns a `uint64` that increments by 1 each time. It **never reuses IDs** — even if you delete task #3, the next task will be #4, not #3.

---

## Converting Integers to Bytes

Since keys must be `[]byte`, you need a helper to convert your `uint64` ID into bytes:

```go
import "encoding/binary"

func itob(v uint64) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, v)
    return b
}
```

### Why BigEndian?

bbolt stores keys in sorted byte order. If you used little-endian encoding:

- `1`  → `01 00 00 00 00 00 00 00`
- `2`  → `02 00 00 00 00 00 00 00`
- `10` → `0A 00 00 00 00 00 00 00`

That sorts fine. But `256` would become `00 01 00 00 00 00 00 00`, which sorts *before* `1`. The ordering breaks down.

With big-endian:

- `1`  → `00 00 00 00 00 00 00 01`
- `2`  → `00 00 00 00 00 00 00 02`
- `10` → `00 00 00 00 00 00 00 0A`
- `256`→ `00 00 00 00 00 00 01 00`

These sort in correct numeric order. When you iterate your tasks with a cursor, they come back as `1, 2, 3, ...` — which is exactly what you want.

---

## A Complete Add Task Example

```go
type Task struct {
    ID          uint64
    Description string
}

func addTask(db *bolt.DB, description string) error {
    return db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("Tasks"))

        id, err := b.NextSequence()
        if err != nil {
            return err
        }

        task := Task{ID: id, Description: description}
        encoded, err := json.Marshal(task)
        if err != nil {
            return err
        }

        return b.Put(itob(id), encoded)
    })
}
```

---

## A Complete Delete Task Example

```go
func deleteTask(db *bolt.DB, id uint64) error {
    return db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("Tasks"))

        // Verify the task exists before deleting
        if b.Get(itob(id)) == nil {
            return fmt.Errorf("task %d not found", id)
        }

        return b.Delete(itob(id))
    })
}
```
