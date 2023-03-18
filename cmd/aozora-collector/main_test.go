package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
)

func TestFindEntries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.String())
		if r.URL.String() == "/" {
			w.Write([]byte(`
			<table summary="作家データ">
<tr><td class="header">作家名：</td><td><font size="+2">芥川 竜之介</font></td></tr>
<tr><td class="header">作家名読み：</td><td>あくたがわ りゅうのすけ</td></tr>
<tr><td class="header">ローマ字表記：</td><td>Akutagawa, Ryunosuke</td></tr>
</table>



<hr>

<div align="right">
［<a href="#sakuhin_list_1">公開中の作品</a>｜<a href="#sakuhin_list_2">作業中の作品</a>］
</div>

<h2><a name="sakuhin_list_1">公開中の作品</a></h2>

<ol>
<li><a href="../cards/000879/card4872.html">愛読書の印象</a>　（新字旧仮名、作品ID：4872）　</li> 
<li><a href="../cards/000879/card16.html">秋</a>　（新字旧仮名、作品ID：16）　</li> 
<li><a href="../cards/000879/card178.html">芥川竜之介歌集</a>　（新字旧仮名、作品ID：178）　</li> 
</ol>
`))
		} else {
			pat := regexp.MustCompile(`.*/cards/([0-9]+)/card([0-9]+).html$`)
			token := pat.FindStringSubmatch(r.URL.String())
			w.Write([]byte(fmt.Sprintf(`
			<h2>作家データ</h2>
<table summary="作家データ">
<tr><td class="header">分類：</td><td>著者</td></tr>
<tr><td class="header">作家名：</td><td><a href="../../index_pages/person879.html">芥川 竜之介</a></td></tr>
<tr><td class="header">作家名読み：</td><td>あくたがわ りゅうのすけ</td></tr>
<tr><td class="header">ローマ字表記：</td><td>Akutagawa, Ryunosuke</td></tr>
</table>


<h2>底本データ</h2>

<table summary="底本データ">
<tr><td class="header">底本：</td><td>芥川龍之介全集　第一巻</td></tr>
<tr><td class="header">出版社：</td><td>岩波書店</td></tr>
<tr><td class="header">初版発行日：</td><td>1996（平成8）年4月8日</td></tr>
<tr><td class="header">入力に使用：</td><td>1996（平成8）年4月8日</td></tr>
<tr><td class="header">校正に使用：</td><td>1996（平成8）年4月8日</td></tr>

</table>

<h2>工作員データ</h2>

<table summary="工作員データ">
<tr><td class="header">入力：</td><td>砂場清隆</td></tr><tr><td class="header">校正：</td><td>高柳典子</td></tr>
</table>

<h2><a name="download">ファイルのダウンロード</a></h2>

<table border="1" summary="ダウンロードデータ" class="download">
<tr>
    <th class="download">ファイル種別</th>
    <th class="download">圧縮</th>
    <th class="download">ファイル名（リンク）</th>
    <th class="download">文字集合／符号化方式</th>
    <th class="download">サイズ</th>
    <th class="download">初登録日</th>
    <th class="download">最終更新日</th>
</tr>
<tr bgcolor="white">
    <td><img src="../images/f1.png" width="16" height="16" border="0" alt="rtxtアイコン">
        テキストファイル(ルビあり)
    </td>
    <td>zip</td>
    <td><a href="./files/%[1]s_%[2]s.zip">%[1]s_%[2]s.zip</a></td>
    <td>JIS X 0208／ShiftJIS</td>
    <td>1873</td>
    <td>2006-02-21</td>
    <td>2006-04-05</td>
</tr>
</table>
`, token[1], token[2])))
		}
	}))
	tmp := pageURLFormat
	pageURLFormat = ts.URL + "/cards/%s/card%s.html"

	defer ts.Close()

	defer func() {
		pageURLFormat = tmp
	}()

	got, err := findEntries(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	want := []Entry{
		{
			AuthorID: "000879",
			Author:   "芥川 竜之介",
			TitleID:  "4872",
			Title:    "愛読書の印象",
			InfoURL:  ts.URL,
			ZipURL:   ts.URL + "/cards/000879/files/000879_4872.zip",
		},
		{
			AuthorID: "000879",
			Author:   "芥川 竜之介",
			TitleID:  "16",
			Title:    "秋",
			InfoURL:  ts.URL,
			ZipURL:   ts.URL + "/cards/000879/files/000879_16.zip",
		},
		{
			AuthorID: "000879",
			Author:   "芥川 竜之介",
			TitleID:  "178",
			Title:    "芥川竜之介歌集",
			InfoURL:  ts.URL,
			ZipURL:   ts.URL + "/cards/000879/files/000879_178.zip",
		},
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %+v, but got %+v\n", want, got)
	}
}
