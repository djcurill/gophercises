# bbolt Guide — Buckets

## What Is a Bucket?

A bucket is a **named container** for key-value pairs inside your database. Think of it like a table in SQL, or a folder on a filesystem.

One database file can hold many buckets:

```
tasks.db
├── Bucket: "Tasks"       ← incomplete tasks
└── Bucket: "Completed"   ← finished tasks
```

You can use multiple buckets to organize different types of data without needing multiple database files.

---

## Creating a Bucket

Buckets must be created inside a **read-write transaction** (`db.Update`):

```go
err := db.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucket([]byte("Tasks"))
    return err
})
```

However, `CreateBucket` returns an error if the bucket **already exists**. For most cases, you want the safe version:

```go
err := db.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
    return err
})
```

`CreateBucketIfNotExists` creates the bucket if it's new, or silently does nothing if it already exists. This is the idiomatic choice.

---

## When to Create Buckets

Create all your buckets **once, at startup**, before doing anything else. Don't scatter `CreateBucketIfNotExists` calls throughout your code — have one dedicated setup function:

```go
func setupDB(db *bolt.DB) error {
    return db.Update(func(tx *bolt.Tx) error {
        buckets := []string{"Tasks", "Completed"}
        for _, name := range buckets {
            if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
                return err
            }
        }
        return nil
    })
}
```

Call `setupDB(db)` right after `bolt.Open`. After that, you can safely assume your buckets exist for the rest of the program's life.

---

## Accessing a Bucket

To use a bucket inside a transaction, call `tx.Bucket()`:

```go
db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("Tasks"))
    if b == nil {
        return fmt.Errorf("bucket 'Tasks' not found")
    }
    // use b...
    return nil
})
```

`tx.Bucket()` returns `nil` if the bucket doesn't exist. Always check for nil if you haven't confirmed the bucket was created — though if you use `setupDB` at startup, you can be confident.

---

## Deleting a Bucket

```go
err := db.Update(func(tx *bolt.Tx) error {
    return tx.DeleteBucket([]byte("Tasks"))
})
```

This deletes the bucket **and all key-value pairs inside it**. It returns an error if the bucket doesn't exist. You probably won't need this in the task manager, but good to know.

---

## Nested Buckets

bbolt supports buckets inside buckets. You likely won't need this for the task manager, but it's worth knowing they exist:

```go
db.Update(func(tx *bolt.Tx) error {
    parent, _ := tx.CreateBucketIfNotExists([]byte("Users"))
    
    // Create a sub-bucket for a specific user
    child, _ := parent.CreateBucketIfNotExists([]byte("alice"))
    child.Put([]byte("task:1"), []byte("Buy milk"))

    return nil
})
```

---

## Summary

| Method | Transaction Type | Behavior |
|---|---|---|
| `tx.CreateBucket(name)` | Update only | Error if already exists |
| `tx.CreateBucketIfNotExists(name)` | Update only | Safe — no error if exists |
| `tx.Bucket(name)` | Update or View | Returns nil if not found |
| `tx.DeleteBucket(name)` | Update only | Deletes bucket and all contents |
