# gojeep 🚙 – A Minimalist Go REST API Starter Kit

[![Go Report Card](https://goreportcard.com/badge/github.com/ferdiebergado/gojeep)](https://goreportcard.com/report/github.com/ferdiebergado/gojeep)

**gojeep** is a lightweight and efficient Go REST API starter kit inspired by the iconic **jeepney**—fast, reliable, and built for the long haul. Designed with simplicity in mind, it follows Go’s standard library-first approach, avoiding unnecessary dependencies while providing a solid foundation for building RESTful services.

## Features

✅ **Minimal Dependencies** – Uses only the Go standard library where possible.  
✅ **Fast & Lightweight** – Just like a jeepney, it keeps things simple and efficient.  
✅ **Production-Ready Structure** – Clean architecture with sensible defaults.  
✅ **Easy to Extend** – Add features as needed without unnecessary bloat.

Whether you're building a new API or need a solid starting point, **gojeep** gets you where you need to go—quickly and reliably. 🚀

## Requirements

-   [Go](https://go.dev)
-   [docker](https://www.docker.com) or [podman](https://podman.io)
-   [make](https://www.gnu.org/software/make)

## Dependencies

This project relies on the following libraries for core functionality:
| **Functionality** | **Library** |
|-----------------------|-------------------------------------------------------------------------------|
| Database | `github.com/jackc/pgx/v5`<br>PostgreSQL driver |
| HTTP Routing | `github.com/ferdiebergado/goexpress`<br>Routing HTTP Requests |
| Validation | `github.com/go-playground/validator/v10`<br>Struct validation |
| Security | `github.com/golang-jwt/jwt/v5`<br>Signing and verifying JSON Web Tokens |
| | `golang.org/x/crypto`<br>Hashing passwords with Argon2 |
| Utilities | `github.com/ferdiebergado/gopherkit`<br>Loading .env files and sending http responses |
| Testing | `github.com/stretchr/testify`<br>Test assertions |
| | `go.uber.org/mock`<br>Interface mocking |
| | `github.com/DATA-DOG/go-sqlmock`<br>Mock SQL database |

## Getting started

1. Copy `.env.example` to `.env`.

```sh
cp .env.example .env
```

2. Generate an APP_KEY.

```sh
make app-key
```

3. Set the values in `.env` based on your environment.

4. Change the placeholder values in `config.json`.

5. Start the app in development mode:

```sh
make dev
```

## Tasks

Common development tasks are automated using the Makefile. To view the available tasks, run the make command.

```sh
make
```
