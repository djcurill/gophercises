# bbolt Guide — Best Practices

## 1. Create Buckets Once, at Startup

Don't scatter `CreateBucketIfNotExists` throughout your code. Have one `setupDB` function that runs all bucket creation right after opening the database.

```go
// Good
func main() {
    db, err := openDB()
    // ...
    if err := setupDB(db); err != nil {
        log.Fatal(err)
    }
    // Now run your commands — buckets are guaranteed to exist
}

func setupDB(db *bolt.DB) error {
    return db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
        return err
    })
}
```

---

## 2. Keep Transactions Short

A read-write transaction (`db.Update`) blocks all other writers while it's open. Do any heavy computation *before* opening the transaction, then use the transaction only for the actual database operation.

```go
// Bad — expensive work inside the transaction
db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    result := doExpensiveComputation() // blocks all writers while this runs
    return b.Put([]byte("key"), result)
})

// Good — prepare data first, write quickly
result := doExpensiveComputation()
db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    return b.Put([]byte("key"), result)
})
```

For a CLI task manager, this distinction barely matters in practice — operations are so fast it's irrelevant. But it's a good habit to build.

---

## 3. Don't Hold Raw Byte Slices Across Transaction Boundaries

The memory that bbolt returns from `Get` is managed by bbolt, not Go's garbage collector. After the transaction closes, it's unsafe to read.

```go
// Bad
var raw []byte
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    raw = b.Get([]byte("key")) // points into bbolt memory
    return nil
})
doSomethingWith(raw) // unsafe — transaction is closed

// Good: unmarshal inside the transaction
var task Task
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    v := b.Get([]byte("key"))
    return json.Unmarshal(v, &task) // task has its own memory now
})
doSomethingWith(task) // safe
```

---

## 4. Use `Update` for Writes, `View` for Reads

Never try to write inside a `View` transaction — it will panic. Keep them clearly separated.

```go
// For listing → View
func listTasks(db *bolt.DB) ([]Task, error) { ... db.View ... }

// For adding → Update
func addTask(db *bolt.DB, desc string) error { ... db.Update ... }

// For deleting → Update
func deleteTask(db *bolt.DB, id uint64) error { ... db.Update ... }
```

---

## 5. Use BigEndian Integer Keys

If you're using numeric IDs as keys (which you should for a task list), always convert with `binary.BigEndian`. This ensures keys sort in correct numeric order when you iterate.

```go
func itob(v uint64) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, v)
    return b
}
```

---

## 6. One Database File, Multiple Buckets

Don't create separate `.db` files for different data types. Put everything in one file, organized into buckets.

```
~/.task/tasks.db
├── Bucket: "Tasks"      ← active tasks
└── Bucket: "Completed"  ← finished tasks (bonus feature)
```

---

## 7. Always Set a Timeout on Open

For CLI tools, a locked database should be a fast, clear error — not an infinite hang.

```go
db, err := bolt.Open(path, 0600, &bolt.Options{
    Timeout: 1 * time.Second,
})
if err == bolt.ErrTimeout {
    fmt.Fprintln(os.Stderr, "Error: database is locked by another process.")
    os.Exit(1)
}
```

---

## 8. Handle the Case Where a Bucket Returns Nil

`tx.Bucket()` returns `nil` if the bucket doesn't exist. If you call `setupDB` at startup this shouldn't happen, but defensive code is still a good practice:

```go
b := tx.Bucket([]byte("Tasks"))
if b == nil {
    return fmt.Errorf("tasks bucket not found — was the database initialized?")
}
```

---

## 9. Structure Your Code in Layers

Keep your bbolt code in its own file or package, separate from your CLI code. Your cobra commands should call simple functions like `addTask(db, description)` — they shouldn't know anything about bbolt buckets or byte conversion.

```
cmd/
  add.go      ← cobra command, calls db.AddTask()
  list.go     ← cobra command, calls db.ListTasks()
  do.go       ← cobra command, calls db.CompleteTask()
db/
  db.go       ← all bbolt code lives here
main.go
```

This makes your code easier to test and reason about.

---

## Common Pitfalls Summary

| Pitfall | Consequence | Fix |
|---|---|---|
| Forgetting `defer db.Close()` | File stays locked | Always defer immediately after Open |
| Writing inside `db.View` | Panic | Use `db.Update` for writes |
| Using `[]byte` from `Get` after transaction | Unsafe memory access | Copy or unmarshal immediately |
| Calling `db.Update` inside `db.Update` | Deadlock | Never nest transactions |
| No timeout on `bolt.Open` | CLI hangs forever if file is locked | Always set `Timeout` in options |
| Little-endian integer keys | Items sort out of numeric order | Always use `binary.BigEndian` |
