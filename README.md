# Catalog Service (Go + MongoDB)

## Endpoints
- `GET /health` — service health.
- `GET /books` — list all books.
- `POST /books` — create a book: { title, author, price, stock }.

## Local dev (without docker)
1. Install go 1.22+
2. Run: `go run .`

## Tests
```
go test ./...
```

- `PUT /books/{id}` — update existing book.
- `DELETE /books/{id}` — delete by id.
