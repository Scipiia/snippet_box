package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox/pkg/models"
	"strconv"
)

// Обработчик главной страницы.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		//helpers.go / func notFound
		app.notFound(w)
		return
	}

	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Используем помощника render() для отображения шаблона.
	app.render(w, r, "home.page.tmpl", &templateData{
		Snippet:  nil,
		Snippets: s,
	})
}

func (app *application) showSnippets(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		//helpers.go / func notFound
		app.notFound(w)
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
		Snippet:  s,
		Snippets: nil,
	})
}

// Обработчик для создания новой заметки.
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		//helpers.go / func clientError
		app.clientError(w, http.StatusMethodNotAllowed)
		//http.Error(w, "метод запрещен", 405)
		return
	}

	// Создаем несколько переменных, содержащих тестовые данные. Мы удалим их позже.
	title := "История про улитку"
	content := "Улитка выползла из раковины,\nвытянула рожки,\nи опять подобрала их."
	expires := "7"

	// Передаем данные в метод SnippetModel.Insert(), получая обратно
	// ID только что созданной записи в базу данных.
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
	}

	app.render(w, r, "create.page.tmpl", &templateData{
		Snippet:  nil,
		Snippets: nil,
	})
	// Перенаправляем пользователя на соответствующую страницу заметки.
	http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
}
