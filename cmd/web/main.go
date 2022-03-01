package main

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"snippetbox/pkg/models/mysql"
	"time"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "сетевой аддрес HTTP")

	//флаг для настойки MySQL подключения
	dsn := flag.String("dns", "web:Password123#@!@/snippetbox?parseTime=true", "название MySQL источника  данных")
	flag.Parse()

	//logger
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Чтобы функция main() была более компактной, мы поместили код для создания
	// пула соединений в отдельную функцию openDB(). Мы передаем в нее полученный
	// источник данных (DSN) из флага командной строки.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	// Мы также откладываем вызов db.Close(), чтобы пул соединений был закрыт
	// до выхода из функции main().
	// Подробнее про defer: https://golangs.org/errors#defer
	defer db.Close()

	// Инициализируем новый кэш шаблона...
	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	//init struct application
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &mysql.SnippetModel{
			DB: db,
		},
		templateCache: templateCache,
	}

	//struct server
	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(), //handlers.go
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		ErrorLog:     errorLog,
	}

	infoLog.Printf("server started to port: %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

//openDB
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

//что бы не дать пользователю скачать системные фаилы
type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
		}
	}

	return f, nil
}
