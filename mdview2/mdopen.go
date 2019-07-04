package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gotk3/gotk3/gtk"
	"github.com/russross/blackfriday"
	"github.com/skratchdot/open-golang/open"

	"path"
	"strings"
)

func getCss(filename string) string {
	return style
}

func makeRes(dir, name string) (res string) {
	filename := filepath.Join(dir, name)
	parts := strings.Split(name, ".")
	typ := mime.TypeByExtension("." + parts[len(parts)-1])
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err.Error()
	}
	b64s := base64.StdEncoding.EncodeToString(data)
	res = fmt.Sprintf("data:%s;base64,%s", typ, b64s)
	return res
}

func getSrcs(body string) []string {
	res := []string{}
	rx := regexp.MustCompile(`[Ss][Rr][Cc]=["']([^"']+)["']`)
	m := rx.FindAllStringSubmatch(body, -1)
	for _, v := range m {
		if strings.HasPrefix(v[1], "data:") == false {
			res = append(res, v[1])
		}
	}
	return res
}

func replaceSrc(body, url, data string) string {
	rx := regexp.MustCompile(`[Ss][Rr][Cc]=["']` + url + `["']`)
	return rx.ReplaceAllString(body, `src="`+data+`"`)
}

//getBody 把markdown文本翻译成html
func getBody(filename string) string {
	b1, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	r := blackfriday.HtmlRenderer(blackfriday.HTML_SKIP_HTML, "", "")
	opts := blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_TABLES
	body := string(blackfriday.Markdown(b1, r, opts))
	srcs := getSrcs(body)
	for _, v := range srcs {
		body = replaceSrc(body, v, makeRes(filepath.Dir(filename), v))
	}
	return body
}

const tmp1 = `<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
<title>{{.title}}</title>
<style>{{.css}}</style>
</head>
<body>
{{.body}}
</body>
</html>
`

//getURI 把文件路径翻译成URI形式（file://...）
func getURI(pathname string) string {
	if strings.HasPrefix(pathname, "/") {
		return "file://" + pathname
	} else {
		cwd, _ := os.Getwd()
		return "file://" + path.Join(cwd, pathname)
	}
}
func main() {
	gtk.Init(&os.Args)
	var h = flag.Bool("h", false, "Show help infomation")
	flag.Parse()
	if *h {
		fmt.Printf("mdview <markdown file>\n")
		return
	}
	name1 := flag.Arg(0)
	if name1 == "" {
		res := gtk.OpenFileChooserNative("Choose A Markdown File", nil)
		name1 = *res
		if name1 == "" {
			fmt.Printf("mdview <markdown file>\n")
			return
		}
	}
	exe1, _ := os.Executable()
	dir1 := path.Dir(exe1)
	var data = make(map[string]string)
	data["title"] = path.Base(name1)
	data["css"] = getCss(path.Join(dir1, "ui/style.css"))
	data["body"] = getBody(name1)
	t := template.New("")
	t.Parse(tmp1)
	buf1, _ := os.Create(path.Join(path.Dir(name1), ".md.htm"))
	//defer os.Remove(buf1.Name())
	t.Execute(buf1, data)
	uri1 := getURI(buf1.Name())
	buf1.Close()
	delete(data, "title")
	delete(data, "css")
	delete(data, "body")
	open.Run(uri1)
}
