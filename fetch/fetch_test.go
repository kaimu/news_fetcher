package fetch

import (
	"encoding/json"
	"testing"
)

func Test_parseIntoNewsItems(t *testing.T) {
	rule := ParsingRule{
		GUID:  "/rss/channel/item/guid",
		Title: "/rss/channel/item/title",
		Link:  "/rss/channel/item/link",
		Date:  "/rss/channel/item/pubDate",
	}
	feed := `
<rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
<channel>
<title>
<![CDATA[ Медиазона ]]>
</title>
<description>
<![CDATA[ Медиазона ]]>
</description>
<link>https://zona.media</link>
<image>
<url>https://zona.media/s/defaultShare.png</url>
<title>Медиазона</title>
<link>https://zona.media</link>
</image>
<generator>RSS for Node</generator>
<lastBuildDate>Tue, 19 May 2020 19:47:00 GMT</lastBuildDate>
<ttl>60</ttl>
<item>
<title>Коронавирус в России. Май</title>
<description>
<![CDATA[ Число зараженных превысило 300 тысяч ]]>
</description>
<link>https://zona.media/chronicle/spring</link>
<guid isPermaLink="false">https://zona.media/40281</guid>
<pubDate>Tue, 19 May 2020 19:45:13 GMT</pubDate>
<enclosure url="https://s3.zona.media/entry/b9edcee2d538ddd1478e041f1d48e101_1400x850" length="281717" type="image/jpeg"/>
</item>
<item>
<title>«Кому положено умереть — помрут» — глава инфоцентра по коронавирусу Мясников о смертях во время эпидемии</title>
<link>https://zona.media/news/2020/05/19/myasnikov</link>
<guid isPermaLink="false">https://zona.media/40682</guid>
<pubDate>Tue, 19 May 2020 19:39:51 GMT</pubDate>
</item>
</channel>
</rss>
	`
	got, err := parseIntoNewsItems(feed, rule)
	expected := []NewsItem{
		{GUID: "https://zona.media/40281",
			Title: "Коронавирус в России. Май",
			Date:  "Tue, 19 May 2020 19:45:13 GMT",
			Link:  "https://zona.media/chronicle/spring"},
		{GUID: "https://zona.media/40682",
			Title: "«Кому положено умереть — помрут» — глава инфоцентра по коронавирусу Мясников о смертях во время эпидемии",
			Date:  "Tue, 19 May 2020 19:39:51 GMT",
			Link:  "https://zona.media/news/2020/05/19/myasnikov"},
	}
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	gotJson, _ := json.Marshal(got)
	expectedJson, _ := json.Marshal(expected)
	if string(gotJson) != string(expectedJson) {
		t.Errorf("Got:\n %v\n, expected:\n %v\n", string(gotJson), string(expectedJson))
	}
}
