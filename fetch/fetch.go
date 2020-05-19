package fetch

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"gopkg.in/xmlpath.v2"
)

const Schema = `
CREATE TABLE IF NOT EXISTS news
(guid TEXT PRIMARY KEY NOT NULL, title TEXT, date TEXT, link TEXT); 
`

// NewsItem новость
type NewsItem struct {
	GUID  string `db:"guid"`
	Title string `db:"title"`
	Date  string `db:"date"` // Кандидат для `time.Time` в реальном проекте
	Link  string `db:"link"`
}

// ParsingRule правило парсинга (просто alias NewsItem, т.к. тот подходит по структуре)
type ParsingRule NewsItem

// News читает Feed, парсерит и пишет результат в базу
func News(url string, rule ParsingRule, db *sqlx.DB) (err error) {
	feedRc, err := fetchFeed(url)
	if err != nil {
		return
	}
	defer feedRc.Close()
	// Читаем Feed в память, т.к. не можем просто передать Reader дальше,
	// иначе его надо сбрасывать для каждой категории (guid, заголовки, и т.д)
	feedStr := new(strings.Builder)
	_, err = io.Copy(feedStr, feedRc)
	if err != nil {
		return
	}
	news, err := parseIntoNewsItems(feedStr.String(), rule)
	if err != nil {
		return
	}
	// Пишем в БД
	return insertNews(news, db)
}

func fetchFeed(url string) (responseBody io.ReadCloser, err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := netClient.Get(url)
	if err != nil {
		return
	}
	responseBody = response.Body
	return
}

func parseIntoNewsItems(feed string, rule ParsingRule) (result []NewsItem, err error) {
	var guids, titles, dates, links []string
	// Извлекаем значения по категориям (GUID, заголовки, и т.д)
	for _, cat := range []struct {
		rule   string
		result *[]string
	}{
		{rule.GUID, &guids},
		{rule.Title, &titles},
		{rule.Date, &dates},
		{rule.Link, &links}} {
		*cat.result, err = readValues(strings.NewReader(feed), cat.rule)
		if err != nil {
			return
		}
	}
	// Проверяем, что все категории значений одной длины
	for _, cat := range []*[]string{&titles, &dates, &links} {
		// Если длинее или короче категории GUID, то возвращаем ошибку
		if len(*cat) != len(guids) {
			err = errors.New("GUID length differs from some other category")
			return
		}
	}
	for i, guid := range guids {
		item := NewsItem{
			GUID:  guid,
			Title: titles[i],
			Date:  dates[i],
			Link:  links[i],
		}
		result = append(result, item)
	}
	return
}

func readValues(data io.Reader, rule string) (values []string, err error) {
	if rule == "" {
		return
	}
	path, err := xmlpath.Compile(rule)
	if err != nil {
		return
	}
	root, err := xmlpath.Parse(data)
	if err != nil {
		return
	}
	items := path.Iter(root)
	for items.Next() {
		values = append(values, items.Node().String())
	}
	return
}

func insertNews(news []NewsItem, db *sqlx.DB) (err error) {
	// Используем REPLACE, а не IGNORE, на случай, если издание поменяло у новости заголовок или дату
	_, err = db.NamedExec(`
	INSERT OR REPLACE INTO news (guid, title, date, link)
	VALUES (:guid, :title, :date, :link);`, news)
	return
}
