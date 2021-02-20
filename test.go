package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"regexp"
	"strings"
)

var(
	Db1 *sql.DB
	ex goods
)
type goods struct {
	matched_str []string
	matched_str2 []string
	matched_str3 []string
}

type what struct {
	url string `db:"url"`
	name string `db:"name"`
	price string `db:"price"`
}

func Init() { //数据库初始化
	usr := "root"
	key := "chenyuhan123000"
	sbase := "news"

	var err error
	Db1, err = sql.Open("mysql",usr+":"+key+"@tcp(localhost:3306)/"+sbase)
	if err != nil {
		fmt.Println("Open Connection failed:", err)
	}
	//fmt.Println(Db1)
}

func Insertintodb() {
	for i, _ := range ex.matched_str {
		_,err := Db1.Exec("insert into what(url,name,price) values (?,?,?)", ex.matched_str[i], ex.matched_str2[i], ex.matched_str3[i])
		if err != nil {
			fmt.Println("error insert",err)
		}
		/*id,err := r.LastInsertId()
		if err != nil {
			fmt.Println("error id")
		}

		fmt.Println("id",id)*/
	}
}

func Getinfo(e *colly.HTMLElement){
	body := string(e.Response.Body[:])
	//图片获取
	match := regexp.MustCompile(`img width="220" height="220" data-img="1" data-lazy-img=".*?"`)
	tmp := match.FindAllString(body,-1)

	head := "https:"
	for _, match_str := range tmp {
		comma := strings.Index(match_str,"/")
		match_str = head + match_str[comma:]
		match_str = strings.TrimRight(match_str, "\"")
		fmt.Println(match_str)
		ex.matched_str = append(ex.matched_str,match_str)
	}


	//商品名获取
	match = regexp.MustCompile(`\W*?<font class="skcolor_ljg">零食</font>`)
	tmp = match.FindAllString(body,-1)

	for _, match_str := range tmp {
		comma := strings.Index(match_str,"\n")
		match_str = match_str[comma+1:]
		match_str = strings.TrimRight(match_str, "<font class=\"skcolor_ljg\">零食</font>")
		fmt.Println(match_str)
		ex.matched_str2 = append(ex.matched_str2,match_str)
	}

	//商品价格获取
	match = regexp.MustCompile(`<em>￥</em><i>.*?</i>`)
	tmp = match.FindAllString(body,-1)

	for _, match_str := range tmp {
		comma := strings.Index(match_str,"i")
		match_str = match_str[comma+2:]
		match_str = strings.TrimRight(match_str, "</i>")
		fmt.Println(match_str)
		ex.matched_str3 = append(ex.matched_str3,match_str)
	}
}

func getMessage(link string) {
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true),
	)

	extensions.RandomUserAgent(c)

	c.Limit(&colly.LimitRule{Parallelism: 20})


	c.OnHTML("#J_goodsList > ul", func(e *colly.HTMLElement) {
		Getinfo(e)
		//fmt.Println(string(e.Response.Body[:]))
	})

	c.Visit(link)

	c.Wait()
}

func main() {
	Init()
	defer Db1.Close()

	link := "https://search.jd.com/Search?keyword=零食&enc=utf-8&wq=零食&pvid=edcff423eebf476fa88cbfd4744e9ec7"

	getMessage(link)
	Insertintodb()
}
