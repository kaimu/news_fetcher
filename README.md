# Агрегатор новостей

## Запуск

```bash
go run . --feed=https://meduza.io/rss/all
```

или

```bash
go run . --feed=https://zona.media/rss
```

После этого страница поиска будет доступна на [http://127.0.0.1:8090](http://127.0.0.1:8090) (и остальных интерфейсах)

Полный список параметров:

```bash
  -feed string
        RSS URL
  -rule-date string
        date XPath expression (default "/rss/channel/item/pubDate")
  -rule-guid string
        GUID XPath expression (default "/rss/channel/item/guid")
  -rule-link string
        link XPath expression (default "/rss/channel/item/link")
  -rule-title string
        title XPath expression (default "/rss/channel/item/title")
```

**Примечание.** Приложение использует `github.com/mattn/go-sqlite3`, которому необходим `gcc` для сборки, т.е.
в Windows для этого требуется установить [TDM_GCC](https://jmeubank.github.io/tdm-gcc/) (или аналог), или использовать WSL.

## Ограничения

1. Так как в решении используются XPath-выражения в качестве правила парсинга, то с XML проблем не будет,
а вот найти HTML, на котором парсер не запнется, будет непросто (открытые теги, JS). Но т.к. большинство новостных сайтов
в наше время все равно подгружают новости динамически с помощью JS, а не генерируют HTML с новостями на стороне сервера,
то я решил обойтись этим наиболее простым решением.

2. У всех новостей с агрегируемого сайта должны быть заданы поля: *guid*, *title*, *pubDate*, *link*.

3. База новостей удаляется при перезагрузке системы.

4. Новости сортируются не по дате, а по порядку, в котором они были добавлены в базу.
