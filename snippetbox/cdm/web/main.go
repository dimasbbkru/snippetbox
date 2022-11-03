package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"

	//"os/signal"
	//"syscall"
	"time"
	// "io/ioutil"
	"os" //для логов
	"path/filepath"

	//"github.com/elastic/go-elasticsearch/v8"
	_ "github.com/go-sql-driver/mysql"
	"golangify.com/snippetbox/pkg/models/mysql"

	//cron "github.com/robfig/cron/v3"
	//"github.com/go-co-op/gocron"
	"context"
	"fmt"

	"github.com/alitto/pond"
	"github.com/segmentio/kafka-go"

	//"github.com/segmentio/kafka-go/protocol/consumer"
	"encoding/json"

	"golangify.com/snippetbox/pkg/models"
	//	"github.com/elastic/go-elasticsearch/v8/esapi"
	//Mongo
	"io/ioutil"
	"strings"
)

// Создаем структуру `application` для хранения зависимостей всего веб-приложения.
// Пока, что мы добавим поля только для двух логгеров, но
// мы будем расширять данную структуру по мере усложнения приложения.

// Добавляем поле snippets в структуру application. Это позволит
// сделать объект SnippetModel доступным для наших обработчиков.

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

func main() {

	// Создаем новый флаг командной строки, значение по умолчанию: ":4000".
	// Добавляем небольшую справку, объясняющая, что содержит данный флаг.
	// Значение флага будет сохранено в переменной addr.
	addr := flag.String("addr", ":4001", "Сетевой адрес HTTP")
	// Определение нового флага из командной строки для настройки MySQL подключения.
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "Название MySQL источника данных")

	// Мы вызываем функцию flag.Parse() для извлечения флага из командной строки.
	// Она считывает значение флага из командной строки и присваивает его содержимое
	// переменной. Вам нужно вызвать ее *до* использования переменной addr
	// иначе она всегда будет содержать значение по умолчанию ":4000".
	// Если есть ошибки во время извлечения данных - приложение будет остановлено.
	flag.Parse()

	// Используйте log.New() для создания логгера для записи информационных сообщений. Для этого нужно
	// три параметра: место назначения для записи логов (os.Stdout), строка
	// с префиксом сообщения (INFO или ERROR) и флаги, указывающие, какая
	// дополнительная информация будет добавлена. Обратите внимание, что флаги
	// соединяются с помощью оператора OR |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Создаем логгер для записи сообщений об ошибках таким же образом, но используем stderr как
	// место для записи и используем флаг log.Lshortfile для включения в лог
	// названия файла и номера строки где обнаружилась ошибка.
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
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Инициализируем новую структуру с зависимостями приложения.
	// Инициализируем экземпляр mysql.SnippetModel и добавляем его в зависимостях.

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
	}
	// Используем методы из структуры в качестве обработчиков маршрутов.
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)
	mux.HandleFunc("/snippet/MakeRequest", app.MakeRequest)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Инициализируем новую структуру http.Server. Мы устанавливаем поля Addr и Handler, так
	// что сервер использует тот же сетевой адрес и маршруты, что и раньше, и назначаем
	// поле ErrorLog, чтобы сервер использовал наш логгер
	// при возникновении проблем.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		//Handler:  mux,
		Handler: app.routes(), // Вызов нового метода app.routes()
	}

	// Значение, возвращаемое функцией flag.String(), является указателем на значение
	// из флага, а не самим значением. Нам нужно убрать ссылку на указатель
	// то есть перед использованием добавьте к нему префикс *. Обратите внимание, что мы используем
	// функцию log.Printf() для записи логов в журнал работы нашего приложения.
	//log.Printf("Запуск сервера на %s", *addr)
	// Применяем созданные логгеры к нашему приложению.
	infoLog.Printf("Запуск сервера на %s", *addr)

	// КРОН - КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН -КРОН- КРОН -КРОН -- -
	//3

	ctx := context.Background()

	pool := pond.New(1, 2)

	for i := 0; i < 10; i++ {
		n := i
		pool.Submit(func() {
			fmt.Printf("Running task #%d\n ", n)
		})

		go consume(ctx)

		time.Sleep(1 * time.Second)

	}

	pool.StopAndWait()

	//err := http.ListenAndServe(*addr, mux)
	// Вызываем метод ListenAndServe() от нашей новой структуры http.Server
	//err := srv.ListenAndServe()
	//errorLog.Fatal(err)
	// Поскольку переменная `err` уже объявлена в приведенном выше коде, нужно
	// использовать оператор присваивания =
	// вместо оператора := (объявить и присвоить)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)

	// MakeRequest()

	// КРОН - КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН -КРОН- КРОН -КРОН -- -
	// КРОН - КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН -КРОН- КРОН -КРОН -- -
	// КРОН - КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН- КРОН -КРОН- КРОН -КРОН -- -
	//1
	// MskTime,_ := time.LoadLocation("Europe/Moscow")
	// scheduler := cron.New(cron.WithLocation(MskTime))
	//	 defer scheduler.Stop()
	// scheduler.AddFunc("0 0 1 1 *", func() { SendAutomail("New Year") })
	//	 scheduler.AddFunc("0 07 10 * *", app.MakeRequest)
	// scheduler.AddFunc("0 09 * * 1-5", NotifyDailyAgenda)
	//	 scheduler.AddFunc("*/10 * * * *", app.MakeRequest)
	// start scheduler
	//	 go scheduler.Start()
	// trap SIGINT untuk trigger shutdown.
	//	 sig := make(chan os.Signal, 1)
	//	 signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//	 <-sig
	//2
	// инициализируем объект планировщика
	//  s := gocron.NewScheduler(time.UTC)
	// добавляем одну задачу на каждую минуту
	//  s.Cron("* * * * *").Do(app.MakeRequest)

	// запускаем планировщик с блокировкой текущего потока
	//  s.StartBlocking()

}

type neuteredFileSystem struct {
	fs http.FileSystem
}

// Функция openDB() обертывает sql.Open() и возвращает пул соединений sql.DB
// для заданной строки подключения (DSN).

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

			return nil, err
		}
	}

	return f, nil
}

func consume(ctx context.Context) {

	const (
		topic          = "ouroboros.service.events"
		broker1Address = "ouroboros-kafka-01.test.cloud.sber-msk-az1.goods.local:9094"
		broker2Address = "ouroboros-kafka-02.test.cloud.sber-msk-az1.goods.local:9094"
		broker3Address = "ouroboros-kafka-03.test.cloud.sber-msk-az1.goods.local:9094"
	)
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker1Address, broker2Address, broker3Address},
		Topic:   topic,
		GroupID: "my-group",
	})
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		fmt.Println("received: ", string(msg.Value))

		//var i interface{} = msg.Value

		kfkresult := models.KfkOrderStatus{}

		//err = json.Unmarshal(msg.Value, &kfkresult)
		err = json.Unmarshal([]byte(msg.Value), &kfkresult)
		if err != nil {
			log.Fatalln(err)
		}

		predicateId := kfkresult.Predicate.ID
		predicateName := kfkresult.Predicate.Name
		orderId := kfkresult.Attributes.Order
		//deliveryId:=kfkresult.Attributes.Delivery
		//shipmentId:= kfkresult.Attributes.Shipment
		//"id":"FFCM00F","name":"Подтвержден продавцом"
		//"id":"DeliveryCreated","name":"Доставка создана"
		if predicateId == "DeliveryCreated" {
			fmt.Println("received: ", predicateName, " ", orderId)

		} else if predicateId == "FFCM00F" {
			fmt.Println("received: ", predicateName, " ", orderId)
		}

		//action","id":"FFCR00F

	}
}

func MongoFind() {
	url := "https://data.mongodb-api.com/app/data-qnbcv/endpoint/data/v1/action/findOne"
	method := "POST"

	payload := strings.NewReader(`{
        "collection":"Mongotest",
        "database":"Mongotest",
        "dataSource":"Cluster0",
        "projection": {"_id": 1}
    }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Access-Control-Request-Headers", "*")
	req.Header.Add("api-key", "<inNVJcbLbpRdaj1OtLSiuBfEEbQVNyUGtG6VBmQlffnVKVgUwBoBkvjHKL3U35AC>")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func MongoInsert() {
	url := "https://data.mongodb-api.com/app/data-qnbcv/endpoint/data/v1/action/insertOne"
	method := "POST"
//text:="Hello from the Data API!"
	payload := strings.NewReader(`{
        "collection":"Mongotest",
        "database":"Mongotest",
        "dataSource":"Cluster0",
		"document": {
			"test": "aa",
			"text": "Text"
		}
    }`)


	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Access-Control-Request-Headers", "*")
	req.Header.Add("api-key", "<inNVJcbLbpRdaj1OtLSiuBfEEbQVNyUGtG6VBmQlffnVKVgUwBoBkvjHKL3U35AC>")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}