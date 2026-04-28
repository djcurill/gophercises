# bbolt Guide — Transactions

## What Is a Transaction?

A transaction is a **wrapper around one or more database operations** that guarantees they either all succeed together or all fail together. You never partially write data — it's all or nothing.

In bbolt, **every single database operation happens inside a transaction**, no exceptions. You can't call `Get` or `Put` directly on the database object — you must open a transaction first.

---

## The Two Transaction Types

### `db.Update` — Read-Write

```go
err := db.Update(func(tx *bolt.Tx) error {
    // You can read AND write in here.
    // Return nil  → all changes are COMMITTED to disk.
    // Return err  → all changes are ROLLED BACK, nothing is saved.
    return nil
})
```

- Only **one** `Update` transaction can be open at a time
- Any other `Update` calls will wait until the current one finishes
- Use for: creating buckets, writing data, deleting data

### `db.View` — Read-Only

```go
err := db.View(func(tx *bolt.Tx) error {
    // You can only read in here.
    // Attempting a write will panic.
    // The return value doesn't affect the database — it's just for your error handling.
    return nil
})
```

- **Multiple** `View` transactions can run at the same time
- They never block each other
- Use for: fetching a single item, listing all items

---

## Why Transactions Matter for Error Handling

The return value of your transaction function controls commit vs rollback. This makes error handling very clean:

```go
err := db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))

    // Step 1
    if err := b.Put([]byte("key1"), []byte("value1")); err != nil {
        return err  // Rolls back everything — key1 is NOT saved
    }

    // Step 2
    if err := b.Put([]byte("key2"), []byte("value2")); err != nil {
        return err  // Rolls back everything — key1 is also NOT saved
    }

    return nil  // Both keys saved atomically
})
```

If step 2 fails, step 1 is also undone. You never end up with partial data.

---

## Common Mistakes

### Mistake 1: Returning nil when you mean to fail

```go
// WRONG — swallowing the error means changes get committed anyway
db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    b.Put([]byte("key"), []byte("value")) // error ignored!
    return nil // commits even if Put failed
})

// RIGHT
db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    return b.Put([]byte("key"), []byte("value"))
})
```

### Mistake 2: Writing inside a View

```go
// WRONG — this will panic at runtime
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    b.Put([]byte("key"), []byte("value")) // PANIC
    return nil
})
```

### Mistake 3: Nesting transactions

bbolt does not support nested transactions. Don't call `db.Update` or `db.View` from inside another transaction function — it will deadlock.

```go
// WRONG — deadlock
db.Update(func(tx *bolt.Tx) error {
    db.View(func(tx2 *bolt.Tx) error { // hangs forever
        return nil
    })
    return nil
})
```

---

## Manual Transactions (Advanced)

For most use cases `db.Update` and `db.View` are all you need. But bbolt also exposes manual transaction control:

```go
// Start a read-write transaction manually
tx, err := db.Begin(true) // true = writable
if err != nil {
    return err
}
defer tx.Rollback() // safe to call even after Commit

// ... do stuff ...

if err := tx.Commit(); err != nil {
    return err
}
```

You'd use this for batch operations where you want to commit in chunks. For the task manager exercise, stick with `db.Update` and `db.View`.
