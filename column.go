package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type my_1 struct {
	Name string `db:"nam"`
	List *colly.HTMLElement `db:lis`
}

var (
	ans my_1
	Da *sql.DB
)

func initial() {
	usr := "root"
	key := "**"
	sbase := "news"
	da,err := sql.Open("mysql",usr+":"+key+"@tcp(localhost:3306)/"+sbase)
	if err != nil {
		fmt.Println("连接数据库失败")
	}

	Da = da
}

func Login(user string,password string)(string,string,string) {
	client := &http.Client{}

	obj := "muser=" + user + "&passwd=" + password

	body := ioutil.NopCloser(strings.NewReader(obj))

	//第一次页面跳转
	req ,_ := http.NewRequest("POST","http://59.77.226.32/logincheck.asp",body)

	req.Header.Set("Referer","http://jwch.fzu.edu.cn/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36 Edg/87.0.664.52")
	req.Header.Set("Upgrade-Insecure-Requests","1")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	response, _ := client.Do(req)
	link,_ := response.Header["Location"]
	cookie,_ :=response.Header["Set-Cookie"]
	pat := "; path=/"
	re, _ := regexp.Compile(pat)
	str := re.ReplaceAllString(cookie[0], "")

	//第二次页面跳转
	req, _ = http.NewRequest("GET",link[0],nil)

	req.Header.Set("Cookie",str)
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36 Edg/87.0.664.52")

	res, _ := client.Do(req)
	link = res.Header["Location"]
	cookie = res.Header["Set-Cookie"]
	p := "default"
	pat = "; path=/; HttpOnly"
	r,_ := regexp.Compile(p)
	re, _ = regexp.Compile(pat)
	Cookie := re.ReplaceAllString(cookie[0], "")
	Link :=r.ReplaceAllString(link[0],"right")

	return Cookie,Link,link[0]
}

func GetHtml(c string,l string,r string)  {
	client := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36 Edg/87.0.664.52"))

	client.OnRequest(func(req *colly.Request) {
		req.Headers.Set("Cookie",c)
		req.Headers.Set("Referer",r)
	})

	client.OnResponse(func(response *colly.Response) {
		//fmt.Println(string(response.Body))
	})

	for  j:=1;j<=7;j++ { //竖列打印课程表
		for i:=1;i<=11;i++ {
			var b string
			if i==1 || i==5 || i==9 {
				b = strconv.Itoa(j+2)
			} else {
				b = strconv.Itoa(j+1)
			}
			client.OnHTML("#LB_kb>table>tbody>tr:nth-child("+strconv.Itoa(i+1)+")>td:nth-child("+b+")", func(e *colly.HTMLElement) {
				if e.Text!=""{
					fmt.Println(e.Text)
				} else {
					fmt.Println("--空课--")
				}
			})
		}
	}

	client.OnHTML("#LB_kb > table", func(e *colly.HTMLElement) {
		ans.Name = "Column"
		ans.List = e
		Da.Exec("insert into my_1(nam,lis)values (?,?)",ans.Name,ans.List.Text)
		//fmt.Println(e.Text)
	})

	client.Visit(l)
}

func main() {
	initial()
	defer Da.Close()

	var user string
	var password string
	fmt.Println("输入用户名：")
	//fmt.Scan(&user)
	fmt.Println("输入密码：")
	//fmt.Scan(&password)
	fmt.Println("(跳过输入用户名及密码部分-.-)")

	user = "031902202"
	password = "cyh051036"
	cookie,link,referer := Login(user,password)

	GetHtml(cookie,link,referer)
}
