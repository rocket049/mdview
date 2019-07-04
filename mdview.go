package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/russross/blackfriday"
	"github.com/sourcegraph/go-webkit2/webkit2"

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
	//buf1, _ := ioutil.TempFile(path.Dir(name1), ".md")
	//defer os.Remove(buf1.Name())
	buf1 := bytes.NewBufferString("")
	t.Execute(buf1, data)
	//uri1 := getURI(buf1.Name())
	//buf1.Close()
	delete(data, "title")
	delete(data, "css")
	delete(data, "body")
	var builder *gtk.Builder
	builder, _ = gtk.BuilderNew()

	err := builder.AddFromString(string(resUi))
	if err != nil {
		panic(err)
	}
	iobj1, _ := builder.GetObject("window1")
	appwin1, ok := iobj1.(*gtk.Window)
	if !ok {
		panic("Fail load Widget 'window1' from mdv.ui")
	}
	defer appwin1.Destroy()
	appwin1.SetIconName("applications-internet")
	iobj2, _ := builder.GetObject("webkit1")
	frame1, ok := iobj2.(*gtk.Frame)
	if !ok {
		panic("Fail load Widget 'webkit1' from mdv.ui")
	}
	defer frame1.Destroy()
	web1 := webkit2.NewWebView()
	web1.Context().SetCacheModel(webkit2.DocumentViewerCacheModel)
	web1.Settings().SetAutoLoadImages(true)

	//web1.LoadURI(uri1)
	web1.LoadHTML("", "about:blank")

	appwin1.Connect("destroy", onDestroy)
	web1.Connect("load-failed", func() {
		fmt.Println("Load failed.")
	})
	web1.Connect("load-changed", func(_ *glib.Object, i int) {
		loadEvent := webkit2.LoadEvent(i)
		switch loadEvent {
		case webkit2.LoadFinished:
			if len(web1.Title()) > 0 {
				appwin1.SetTitle(web1.Title())
			} else {
				web1.LoadHTML(buf1.String(), "")
			}

		}
	})
	web1.SetVisible(true)
	frame1.Add(web1)
	frame1.ShowAll()
	gtk.Main()
}

func onDestroy() {
	gtk.MainQuit()
}
