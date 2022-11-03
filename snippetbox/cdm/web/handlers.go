package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"

	// "io/ioutil"
	//"log"
	//"io/ioutil"
	//"bytes"
	//"encoding/json"

	"net/http"
	"strconv"

	"golangify.com/snippetbox/pkg/models"
)

// Создается функция-обработчик "home", которая записывает байтовый слайс, содержащий
// текст "Привет из Snippetbox" как тело ответа.
func (app *application) home(w http.ResponseWriter, r *http.Request) {

	// Проверяется, если текущий путь URL запроса точно совпадает с шаблоном "/". Если нет, вызывается
	// функция http.NotFound() для возвращения клиенту ошибки 404.
	// Важно, чтобы мы завершили работу обработчика через return. Если мы забудем про "return", то обработчик
	// продолжит работу и выведет сообщение "Привет из SnippetBox" как ни в чем не бывало.
	if r.URL.Path != "/" {
		app.notFound(w) // Использование помощника notFound()
		return
	}
	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "home.page.tmpl", &templateData{
		Snippets: s,
	})
	// Создаем экземпляр структуры templateData,
	// содержащий срез с заметками.
	data := &templateData{Snippets: s}

	// Инициализируем срез содержащий пути к двум файлам. Обратите внимание, что
	// файл home.page.tmpl должен быть *первым* файлом в срезе.
	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	// Используем функцию template.ParseFiles() для чтения файлов шаблона.
	// Если возникла ошибка, мы запишем детальное сообщение ошибки и
	// используя функцию http.Error() мы отправим пользователю
	// ответ: 500 Internal Server Error (Внутренняя ошибка на сервере)
	ts, err := template.ParseFiles(files...)
	if err != nil {
		// Поскольку обработчик home теперь является методом структуры application
		// он может получить доступ к логгерам из структуры.
		// Используем их вместо стандартного логгера от Go.
		//app.errorLog.Println(err.Error())
		//http.Error(w, "Внутренняя ошибка сервера", 500)
		app.serverError(w, err) // Использование помощника serverError()
		return
	}

	// Используем функцию template.ParseFiles() для чтения файла шаблона.
	// Если возникла ошибка, мы запишем детальное сообщение ошибки и
	// используя функцию http.Error() мы отправим пользователю
	// ответ: 500 Internal Server Error (Внутренняя ошибка на сервере)
	// ts, err := template.ParseFiles("./ui/html/home.page.tmpl")
	// if err != nil {
	// 	log.Println(err.Error())
	// 	http.Error(w, "Internal Server Error", 500)
	// 	return
	// }

	// Затем мы используем метод Execute() для записи содержимого
	// шаблона в тело HTTP ответа. Последний параметр в Execute() предоставляет
	// возможность отправки динамических данных в шаблон.
	err = ts.Execute(w, data)
	if err != nil {
		// Обновляем код для использования логгера-ошибок
		// из структуры application.
		// app.errorLog.Println(err.Error())
		// http.Error(w, "Внутренняя ошибка сервера", 500)
		app.serverError(w, err) // Использование помощника serverError()
	}
}

// w.Write([]byte("Привет из Snippetbox"))
// }

// Обработчик для отображения содержимого заметки.

// Меняем сигнатуру обработчика showSnippet, чтобы он был определен как метод
// структуры *application
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Отображение заметки..."))

	// Извлекаем значение параметра id из URL и попытаемся
	// конвертировать строку в integer используя функцию strconv.Atoi(). Если его нельзя
	// конвертировать в integer, или значение меньше 1, возвращаем ответ
	// 404 - страница не найдена!
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		// http.NotFound(w, r)
		app.notFound(w) // Использование помощника notFound()
		return
	}

	// Вызываем метода Get из модели Snipping для извлечения данных для
	// конкретной записи на основе её ID. Если подходящей записи не найдено,
	// то возвращается ответ 404 Not Found (Страница не найдена).
	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	// Используем помощника render() для отображения шаблона.
	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
	// Создаем экземпляр структуры templateData, содержащей данные заметки.
	data := &templateData{Snippet: s}

	// Инициализируем срез, содержащий путь к файлу show.page.tmpl
	// Добавив еще базовый шаблон и часть футера, который мы сделали ранее.
	files := []string{
		"./ui/html/show.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	// Парсинг файлов шаблонов...
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// А затем выполняем их. Обратите внимание на передачу заметки с данными
	// (структура models.Snippet) в качестве последнего параметра.
	err = ts.Execute(w, data)
	if err != nil {
		app.serverError(w, err)
	}

	// Используем функцию fmt.Fprintf() для вставки значения из id в строку ответа
	// и записываем его в http.ResponseWriter.
	// fmt.Fprintf(w, "Отображение выбранной заметки с ID %d...", id)
	// Отображаем весь вывод на странице.
	fmt.Fprintf(w, "%v", s)
}

// Обработчик для создания новой заметки.
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {

	// Вызываем щаблон страницы create
	app.render(w, r, "create.page.tmpl", &templateData{
		// Snippet: s,
	})

	if r.Method != http.MethodGet {

		//	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {

		title := r.FormValue("title")
		content := r.FormValue("content")
		expires := r.FormValue("expires")

		// Если из формы отправить GET запрос, то URL можно вытащить параметры так
		//	title := r.URL.Query().Get("title")
		//	content := r.URL.Query().Get("content")
		//	expires := r.URL.Query().Get("expires")

		// fmt.Fprintf(w, "title: %s content: %s expires %s id %s", title, content, expires, id)

		// передаем в функцию insert параметры из формы, получаем ID записи в БД
		id, err := app.snippets.Insert(title, content, expires)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// fmt.Fprintf(w, "title: %s content: %s expires %s  go to snippet:", title, content, expires)

		// ПОказываем ссылку на страницу с заметкой
		http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)

	}
}

// w.Header().Set("Allow", http.MethodPost)

// app.clientError(w, http.StatusMethodNotAllowed) // Используем помощник clientError()

// добавляем в заголовок ответа указание о том, что Content-Type = application/json, в стандарте GO JSON не определяет
// w.Header().Set("Content-Type", "application/json")
// w.Header().Set("Cache-Control", "public, max-age=31536000")
// w.Header()["Date"] = nil
// w.Write([]byte(`{"name":"Alex"}`))
// http.Error(w, "Метод запрещен!", 405)
// return
// }

// Создаем несколько переменных, содержащих тестовые данные. Мы удалим их позже.

// title := "testtitl"
// content := "testconten"
// expires := "1"

// Передаем данные в метод SnippetModel.Insert(), получая обратно
// ID только что созданной записи в базу данных.

// ЗАДАЕМ ЗАГОЛОВОК ОТВЕТА
//добавляем в заголовок ответа указание о том, что Content-Type = application/json, в стандарте GO JSON не определяет
//	 w.Header().Set("Content-Type", "application/json")
// ЗАДАЕМ ТЕЛО ОТВЕТА
//		w.Write([]byte(`{"name":"Alex"}`))
//		w.Write([]byte("Форма для создания новой заметки..."))

// Запрос GET без тела на моковый адрес
// func (app *application) MakeRequest(w http.ResponseWriter, r *http.Request) {
// resp, err := http.Get("https://httpbin.org/get")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	fmt.Fprintf(w, string(body))
// log.Println(string(body))
// }

// Запрос в JSON формате
func (app *application) MakeRequest(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("дддд")

	message := map[string]interface{}{
		// "hello": "world",
		// "life":  42,
		// "embedded": map[string]string{
		//	"yes": "of course!",

		"data": map[string]string{
			"login":    "nodluga@mail.ru",
			"password": "gBc09",
		},
		"meta": map[string]string{},
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post("http://api-http.test.goods.local/api/market/v2/securityService/session/start",
		"application/json", bytes.NewBuffer(bytesRepresentation))
	bytes.NewBuffer(bytesRepresentation)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	result := models.SessionStartResult{}

	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(jsonDataFromHttp), &result) 

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintln(w, result.Data.SessionID)

	sessionIDcol := result.Data.SessionID

	id, err := app.snippets.InsertSessionID(sessionIDcol)
	if err != nil {
		app.serverError(w, err)
		return
	}
	// Выводит ID, убрать когда будет крон
	fmt.Fprintln(w, id)
	fmt.Fprintf(w, result.Data.SessionID)
}
