package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	"github.com/kaimu/news_fetcher/fetch"
	"github.com/kaimu/news_fetcher/serve"
	_ "github.com/mattn/go-sqlite3"
)

var feed = flag.String("feed", "", "RSS URL")
var ruleGUID = flag.String("rule-guid", "/rss/channel/item/guid", "GUID XPath expression")
var ruleTitle = flag.String("rule-title", "/rss/channel/item/title", "title XPath expression")
var ruleLink = flag.String("rule-link", "/rss/channel/item/link", "link XPath expression")
var ruleDate = flag.String("rule-date", "/rss/channel/item/pubDate", "date XPath expression")

func main() {
	flag.Parse()

	// Создаем временную папку
	tmpDir := path.Join(os.TempDir(), "news_fetcher")
	_, err := os.Stat(tmpDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(tmpDir, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	// Готовим БД
	db, err := sqlx.Connect("sqlite3", dbPath(tmpDir))
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(fetch.Schema)

	// Добавляем новости
	if *feed != "" {
		rule := fetch.ParsingRule{
			GUID:  *ruleGUID,
			Title: *ruleTitle,
			Link:  *ruleLink,
			Date:  *ruleDate,
		}
		log.Printf("Fetching feed from %v with the parsing rule:\n%v\n", *feed, rule)
		err = fetch.News(*feed, rule, db)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Println("No feed provided to fetch news from")
	}

	// Запускаем web-сервер для поиска по существующим новостям
	err = createStaticIndex(tmpDir)
	if err != nil {
		log.Fatalln(err)
	}
	serve.Listen(":8090", tmpDir, db)
}

func dbPath(dbDir string) string {
	result := path.Join(dbDir, "data.db")
	log.Printf("Database is temporary stored at: %v", result)
	return result
}

func createStaticIndex(dir string) (err error) {
	file := []byte(indexHTML)
	filePath := path.Join(dir, "index.html")
	err = ioutil.WriteFile(filePath, file, 0644)
	return
}

const indexHTML = `
<html>
<body onload="sendReq()">
    <h1>Агрегатор новостей</h1>
    <label for="search">Поиск по заголовку:</label>
    <input type="text" id="search" oninput="sendReq()">
    <section></section>
    <script>
        let input = document.getElementsByTagName("input")[0];
        let section = document.getElementsByTagName("section")[0];
        function sendReq() {
            let searchParams = new URLSearchParams();
            searchParams.set("term", input.value);
            let url = new URL("search", window.location.origin);
            url.search = searchParams;
            fetch(url, {
                method: 'GET',
            })
                .then(response => {
                    return response.json();
                })
                .then(data => {
                    if (data.term == input.value) {
                        section.innerHTML = "";
                        if (data.results) {
                            for (let i = 0; i < data.results.length; i++) {
                                section.appendChild(newsBody(data.results[i]));
                            }
                        }
                    }
                }).catch(e => {
                    window.alert(e);
                });
        }
        function newsBody(newsItem) {
            let date = document.createElement("span");
            date.innerHTML = newsItem.Date;
            let title = document.createElement("span");
            title.innerHTML = newsItem.Title;
            let link = document.createElement("a");
            link.href = newsItem.Link;
            link.appendChild(title);
            let p = document.createElement("p");
            p.appendChild(date);
            p.appendChild(document.createElement("br"));
            p.appendChild(link);
            return p
        }
    </script>
</body>
</html>
`
