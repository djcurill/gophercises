# bbolt Guide — Overview

## What Is bbolt?

bbolt is an **embedded key-value store** — meaning it's not a separate server process you connect to (like Postgres or Redis). The database lives entirely inside your application as a library, and it stores everything in **a single file** on disk.

This is perfect for CLI tools and small applications where you don't want to manage a separate database server.

---

## Why bbolt for a CLI Task Manager?

- No server to start or stop
- No connection strings or credentials
- Data persists between runs automatically
- It's just a file — easy to back up, move, or delete
- Zero configuration

---

## The Big Picture

Before diving into individual topics, here's the mental model that everything else builds on:

```
bolt.Open()  →  gives you a *bolt.DB

db.Update()  →  read-write transaction  →  tx.Bucket()  →  Put / Delete / NextSequence
db.View()    →  read-only transaction   →  tx.Bucket()  →  Get / Cursor

db.Close()   →  flushes and releases the file lock
```

Every interaction with the database flows through a **transaction function**. You never touch the database directly — you always work inside `db.Update()` or `db.View()`.

---

## The Data Model

bbolt has exactly two concepts:

**Buckets** — named containers, similar to a table or a folder. You must create a bucket before storing anything in it.

**Key-Value Pairs** — everything inside a bucket is a `[]byte` key mapped to a `[]byte` value.

```
Database
└── Bucket: "Tasks"
    ├── key: []byte{0,0,0,0,0,0,0,1} → value: []byte(`{"desc":"Buy milk"}`)
    ├── key: []byte{0,0,0,0,0,0,0,2} → value: []byte(`{"desc":"Walk dog"}`)
    └── key: []byte{0,0,0,0,0,0,0,3} → value: []byte(`{"desc":"Write code"}`)
```

**Both keys and values are raw bytes.** This means you are responsible for serialization — if you want to store a struct, you encode it into bytes first (JSON, gob, etc.) and decode it back when you read it.

---

## Files in This Guide

| File | Topic |
|---|---|
| `00_overview.md` | This file — the big picture |
| `01_opening_closing.md` | Opening and closing the database |
| `02_transactions.md` | How transactions work |
| `03_buckets.md` | Creating and accessing buckets |
| `04_writing.md` | Writing data (Put, Delete, NextSequence) |
| `05_reading.md` | Reading data (Get, Cursor) |
| `06_serialization.md` | Encoding structs as bytes |
| `07_best_practices.md` | Best practices and common mistakes |
| `08_full_example.md` | A complete worked example |
