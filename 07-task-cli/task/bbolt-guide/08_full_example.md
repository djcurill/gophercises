# bbolt Guide — Full Worked Example

This is a complete, working database layer for the CLI task manager exercise.
It covers `add`, `list`, and `do` (complete task) — the three core commands.

This is NOT the answer to the exercise — it's intentionally just the database
layer. The cobra CLI wiring is left for you to build.

---

## File Structure

```
task/
├── main.go
├── cmd/
│   ├── root.go
│   ├── add.go
│   ├── list.go
│   └── do.go
└── db/
    └── db.go    ← this file
```

---

## db/db.go

```go
package db

import (
    "encoding/binary"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"

    bolt "go.etcd.io/bbolt"
)

// ----- Types -----

type Task struct {
    ID          uint64    `json:"id"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}

var tasksBucket = []byte("Tasks")

// ----- Setup -----

// Open opens (or creates) the database in the user's home directory.
func Open() (*bolt.DB, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("could not find home directory: %w", err)
    }

    dbDir := filepath.Join(home, ".task")
    if err := os.MkdirAll(dbDir, 0700); err != nil {
        return nil, fmt.Errorf("could not create db directory: %w", err)
    }

    dbPath := filepath.Join(dbDir, "tasks.db")
    db, err := bolt.Open(dbPath, 0600, &bolt.Options{
        Timeout: 1 * time.Second,
    })
    if err != nil {
        return nil, fmt.Errorf("could not open db: %w", err)
    }

    if err := setup(db); err != nil {
        db.Close()
        return nil, err
    }

    return db, nil
}

// setup creates required buckets if they don't exist.
func setup(db *bolt.DB) error {
    return db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists(tasksBucket)
        return err
    })
}

// ----- Core Operations -----

// AddTask inserts a new task and returns its assigned ID.
func AddTask(db *bolt.DB, description string) (uint64, error) {
    var id uint64

    err := db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(tasksBucket)

        seq, err := b.NextSequence()
        if err != nil {
            return err
        }
        id = seq

        task := Task{
            ID:          id,
            Description: description,
            CreatedAt:   time.Now(),
        }

        encoded, err := json.Marshal(task)
        if err != nil {
            return err
        }

        return b.Put(itob(id), encoded)
    })

    return id, err
}

// ListTasks returns all tasks in order.
func ListTasks(db *bolt.DB) ([]Task, error) {
    var tasks []Task

    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(tasksBucket)

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

// DeleteTask removes a task by ID. Returns an error if the task doesn't exist.
func DeleteTask(db *bolt.DB, id uint64) (Task, error) {
    var task Task

    err := db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(tasksBucket)

        // Fetch it first so we can return the description in the success message
        v := b.Get(itob(id))
        if v == nil {
            return fmt.Errorf("task %d not found", id)
        }

        if err := json.Unmarshal(v, &task); err != nil {
            return err
        }

        return b.Delete(itob(id))
    })

    return task, err
}

// ----- Helper -----

// itob converts a uint64 to an 8-byte big-endian slice.
// Big-endian ensures keys sort in correct numeric order.
func itob(v uint64) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, v)
    return b
}
```

---

## How the Cobra Commands Would Use This

This is pseudocode showing the pattern — not the full cobra wiring:

```go
// cmd/add.go
var addCmd = &cobra.Command{
    Use:   "add",
    Short: "Add a new task",
    Run: func(cmd *cobra.Command, args []string) {
        description := strings.Join(args, " ")

        db, err := db.Open()
        if err != nil { log.Fatal(err) }
        defer db.Close()

        id, err := db.AddTask(db, description)
        if err != nil { log.Fatal(err) }

        fmt.Printf("Added \"%s\" to your task list.\n", description)
        _ = id // you could print the ID if you wanted
    },
}
```

```go
// cmd/list.go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all tasks",
    Run: func(cmd *cobra.Command, args []string) {
        db, err := db.Open()
        if err != nil { log.Fatal(err) }
        defer db.Close()

        tasks, err := db.ListTasks(db)
        if err != nil { log.Fatal(err) }

        if len(tasks) == 0 {
            fmt.Println("You have no tasks.")
            return
        }

        fmt.Println("You have the following tasks:")
        for i, t := range tasks {
            fmt.Printf("%d. %s\n", i+1, t.Description)
        }
    },
}
```

```go
// cmd/do.go
var doCmd = &cobra.Command{
    Use:   "do [task number]",
    Short: "Mark a task as complete",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // Note: the number the user sees is the LIST position (1, 2, 3...)
        // but the database key is the actual ID (could be 1, 5, 12...)
        // You need to list tasks first, find the nth one, then delete by its ID.

        db, err := db.Open()
        if err != nil { log.Fatal(err) }
        defer db.Close()

        tasks, err := db.ListTasks(db)
        if err != nil { log.Fatal(err) }

        // parse args[0] as an integer to get the position
        // bounds check it against len(tasks)
        // get the actual task: tasks[position-1]
        // call db.DeleteTask(db, tasks[position-1].ID)
    },
}
```

### The "do" Command Subtlety

Notice the comment in `doCmd`. The user types `task do 1` meaning "the first task in the list." But the database ID of that task might be `7` if tasks 1–6 were already deleted.

The pattern to handle this:
1. Call `ListTasks` to get all tasks in order
2. Parse the user's number as a list index (`1`-based)
3. Look up `tasks[index-1].ID` to get the real database ID
4. Call `DeleteTask` with that real ID

---

## Key Things to Notice in This Example

- `Open()` handles path setup, directory creation, and bucket initialization — callers don't need to think about any of that
- `DeleteTask` fetches the task before deleting it, so it can return the task description for a friendly success message
- All bbolt code is in `db/db.go` — the cobra commands just call simple functions
- The `itob` helper lives alongside the database code, not scattered around
- Errors are wrapped with `fmt.Errorf("context: %w", err)` for better messages
