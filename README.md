# config

A tiny INI-style config file package for Go, with live reload.

## Install

```sh
go get github.com/eqto/config
```

## File format

```ini
# comments start with # or //

[server]
host = localhost
port = 8080

# repeat a key for a list of values
[server]
domain = example.com
domain = example.org

# prefix a section with - for an array of objects
-[database]
host = db1
port = 5432

-[database]
host = db2
port = 5432
```

## Usage

Package-level functions operate on a shared file opened with `Open`:

```go
if err := config.Open("configs/app.conf"); err != nil {
    log.Fatal(err)
}

host := config.Get("server.host")          // "localhost"
port := config.GetInt("server.port")       // 8080
name := config.GetOr("server.name", "app") // default when missing

domains := config.GetStringArray("server.domain") // ["example.com", "example.org"]

for _, db := range config.GetArray("database") {
    fmt.Println(db.Get("host"), db.GetInt("port"))
}
```

Changes made with `Put` can be written back with `Save`:

```go
config.Put("server.port", "9090")
config.Save()
```

### Live reload

Parsed files are watched automatically. Register a callback to be notified
when the file changes on disk:

```go
config.OnChanges(func() {
    log.Println("config reloaded")
})
```

To watch a file in any format without parsing it, use `Watch` and keep the
handle so you can stop it later:

```go
w := config.Watch("data.json", func() {
    log.Println("data.json changed")
})
defer w.Close()
```

### Multiple files

Use `ParseFile` for independent instances instead of the shared one:

```go
cfg, err := config.ParseFile("other.conf")
if err != nil {
    log.Fatal(err)
}
defer cfg.Close()
fmt.Println(cfg.Get("server.host"))
```

## License

MIT
