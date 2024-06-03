# client-server-api
Estudo sobre webserver http, contextos, banco de dados e manipulação de arquivos com Go.

# SQLite
No Windowns, precisei instalar o [TDM-GCC](https://jmeubank.github.io/tdm-gcc/), devido ter gerado o erro: 

```
Error on opening database connection: %sBinary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

Passos após a instalação:
- Abrir o _MinGW Command Prompt_
- Entrar na masta do projeto `cd C:\pasta\do\projeto`
- Baixar a dependência do sqlite: `go get -u github.com/mattn/go-sqlite3`
- Instalar a dependência do sqlite: `go install github.com/mattn/go-sqlite3`
- Rodar a aplicação: `go run server\server.go`