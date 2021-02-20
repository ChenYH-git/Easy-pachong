package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"regexp"
	"strconv"
)

type data struct {
	id int `json:"Id"`
}

type My struct { //数据表
	Author string  `db:"author"`
	Rcnt string  `db:"rcnt"`
	Text *colly.HTMLElement `db:"tex"`
	Date string  `db:"dat"`
	Title string `db:"titl"`
}

var (
	Db1 *sql.DB
)

func Init() { //数据库初始化
	usr := "root"
	key := "chenyuhan123000"
	sbase := "news"
	db,err := sql.Open("mysql",usr+":"+key+"@tcp(localhost:3306)/"+sbase)
	if err != nil {
		fmt.Println("连接数据库失败")
	}

	Db1 = db
}

func FindId(i *int,a *colly.Collector) { //正则获取动态Id
	a.OnHTML("script", func(e *colly.HTMLElement) {
		reg := regexp.MustCompile(`\d{6}`)
		result := reg.FindString(e.Text)
		if result != "" {
			*i,_ = strconv.Atoi(result)
		}
	})
}

func getText(detail *colly.Collector,e *colly.HTMLElement,link string) (My) { //获取文章主体
	var ans My
	//爬取发布日期，作者,标题
	detail.OnHTML("#fbsj", func(element *colly.HTMLElement) {
		ans.Date = element.Text
		fmt.Printf("发布日期：%s  ",ans.Date)
	})

	detail.OnHTML("#author", func(element *colly.HTMLElement) {
		ans.Author = element.Text
		fmt.Printf("作者：%s  ",ans.Author)
	})

	detail.OnHTML("#main > div.right > div.detail_main_content > p", func(element *colly.HTMLElement) {
		ans.Title = element.Text
		fmt.Printf("%s\n",element.Text)
	})

	//爬取文章主体内容
	detail.OnHTML("#news_content_display", func(element *colly.HTMLElement) {
		ans.Text = element
		fmt.Printf("%q\n",element.Text)
		fmt.Println()
		fmt.Println()
	})

	detail.OnError(func(r *colly.Response, err error) {
		fmt.Println("detail Request URL:", r.Request.URL,"failed with response", r, "\nError",err)
	})

	detail.Visit(e.Request.AbsoluteURL(link)) //开始爬取

	return ans
}

func getAll(detail *colly.Collector,glink *colly.Collector) { //获取文章链接
	var Id int

	glink.OnHTML(".list_main_content ul li a[href]" , func(e *colly.HTMLElement) {
		link := e.Attr("href")
		ans := getText(detail,e,link) //访问链接并爬取
		dynamics := colly.NewCollector() //用于获取动态Id的容器
		dynamics.OnResponse(func(r *colly.Response) {
			ans.Rcnt = string(r.Body[:]) //得到阅读数
			fmt.Printf("阅读数：%s\n",string(r.Body[:]))
		})

		FindId(&Id,detail)//得到动态Id
		d:=data{Id }
		j,_:=json.Marshal(d)

		if Id != 0 { //去除不需要的Id
			dynamics.PostRaw("http://news.fzu.edu.cn/interFace/getDocReadCount.do?id="+strconv.Itoa(Id), j) //Post方法登录
		}
		Db1.Exec("insert into my(author, rcnt, tex, dat, titl) values (?,?,?,?,?)",ans.Author,ans.Rcnt,ans.Text.Text,ans.Date,ans.Title) //插入数据库中
	})

	glink.OnError(func(r *colly.Response, err error) {
		fmt.Println("glink Request URL:", r.Request.URL,"failed with response", r, "\nError",err)
	})

	glink.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	//主页面翻页
	for i := 1; i <= 13; i++ {
		page := fmt.Sprintf("http://news.fzu.edu.cn/html/fdyw/%d.html", i)
		glink.Visit(page)
	}
}

func main() {
	Init()
	defer Db1.Close() //数据库操作

	glink := colly.NewCollector()

	detail := colly.NewCollector(
		//colly.MaxDepth(1),
		//colly.Async(true),  //开启异步
	)

	//detail.Limit(&colly.LimitRule{Delay:1*time.Second, RandomDelay: 1*time.Second, Parallelism: 200})

	extensions.RandomUserAgent(detail)

	getAll(detail,glink)

	//detail.Wait()
}
