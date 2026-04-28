# bbolt Guide — Opening & Closing the Database

## Opening the Database

```go
db, err := bolt.Open("mydata.db", 0600, nil)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### What each argument means

| Argument | Value | Meaning |
|---|---|---|
| path | `"mydata.db"` | File path. Created if it doesn't exist, opened if it does. |
| mode | `0600` | Unix file permissions. Owner can read/write, nobody else. |
| options | `nil` | Uses defaults. See below for useful overrides. |

### The file lock

When bbolt opens a file, it places an **exclusive lock** on it. This means:

- Only **one process** can have the database open at a time
- If your CLI is already running and you open a second terminal and run it again, the second instance will **wait forever** by default

For a CLI tool, hanging forever is terrible UX. Fix it with a timeout:

```go
db, err := bolt.Open("mydata.db", 0600, &bolt.Options{
    Timeout: 1 * time.Second,
})
if err != nil {
    // If the file is locked, err will be bolt.ErrTimeout
    if err == bolt.ErrTimeout {
        log.Fatal("Database is already in use by another process.")
    }
    log.Fatal(err)
}
```

---

## Closing the Database

```go
defer db.Close()
```

`Close()` does two things:

1. Flushes all pending writes to disk
2. Releases the file lock so other processes can open the file

Always call it. `defer` right after `Open` is the idiomatic Go pattern — it guarantees cleanup even if your program panics or returns early.

---

## Where to Store the Database File

For a CLI tool that works from any directory, you need an **absolute path** — not a relative one like `"tasks.db"`. If you use a relative path, the database file appears in whatever directory the user ran the command from, which means you'd have a different database in every folder.

The standard solution is to store it in the user's home directory:

```go
import "os"

home, err := os.UserHomeDir()
if err != nil {
    log.Fatal(err)
}

dbPath := filepath.Join(home, ".task", "tasks.db")
```

This puts the database at `~/.task/tasks.db` on Mac/Linux and the equivalent on Windows.

### Creating the directory if it doesn't exist

```go
err = os.MkdirAll(filepath.Dir(dbPath), 0700)
if err != nil {
    log.Fatal(err)
}

db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
```

`os.MkdirAll` is like `mkdir -p` — it creates the full path and does nothing if it already exists.

---

## Putting It Together

A clean `initDB()` function you might write once and call at startup:

```go
func initDB() (*bolt.DB, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    dbDir := filepath.Join(home, ".task")
    if err := os.MkdirAll(dbDir, 0700); err != nil {
        return nil, err
    }

    dbPath := filepath.Join(dbDir, "tasks.db")
    db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
    if err != nil {
        return nil, err
    }

    return db, nil
}
```
