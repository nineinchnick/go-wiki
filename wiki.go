package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

const frontPage = "FrontPage"
const dataDir = "data"

func filename(dataDir, title string) string {
	return path.Join(dataDir, html.EscapeString(title)+".md")
}

func (p *Page) save(dataDir string) error {
	filename := filename(dataDir, p.Title)
	err := ioutil.WriteFile(filename, p.Body, 0600)
	if err != nil {
		log.Printf("ERROR Saving %s", filename)
	}
	return err
}

func loadPage(dataDir, title string) (*Page, error) {
	filename := filename(dataDir, title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("INFO Missing file %s", filename)
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

type Context struct {
	Page    *Page
	BaseUrl string
}

var pageTitle = regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")

func linkPages(body []byte, baseUrl string) template.HTML {
	return template.HTML(pageTitle.ReplaceAllFunc(body, func(title []byte) []byte {
		return []byte(fmt.Sprintf("<a href=\"%s%s\">%s</a>", baseUrl, title[1:len(title)-1], title[1:len(title)-1]))
	}))
}

func autoIndex(dataDir, baseUrl string) template.HTML {
	allfiles, _ := filepath.Glob(path.Join(dataDir, "*.md"))
	files := make([]string, 0)
	for _, v := range allfiles {
		if v != "" && v != frontPage {
			basename := filepath.Base(v)
			name := strings.TrimSuffix(basename, filepath.Ext(basename))
			files = append(files, name)
		}
	}
	t := fmt.Sprintf("{{if .}}<ul>{{range .}}<li><a href=\"%s{{.}}\">{{.}}</a></li>{{end}}</ul>{{else}}No pages{{end}}", baseUrl)
	var result bytes.Buffer
	template.Must(template.New("main").Parse(t)).Execute(&result, files)
	return template.HTML(result.String())
}

var templates = make(map[string]*template.Template)

func loadTemplates() {
	layoutFiles, err := filepath.Glob("templates/*.html")
	if err != nil {
		panic(err)
	}

	includeFiles, err := filepath.Glob("templates/*.tpl")
	if err != nil {
		panic(err)
	}

	mainTemplate := template.New("main").Funcs(template.FuncMap{
		"linkPages": linkPages,
		"autoIndex": func(baseUrl string) template.HTML {
			return autoIndex(dataDir, baseUrl)
		},
	})

	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)
		templates[fileName] = template.Must(template.Must(mainTemplate.Clone()).ParseFiles(files...))
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, c *Context) {
	log.Printf("DEBUG Executing template %s on %s with baseUrl: %s", tmpl, c.Page.Title, c.BaseUrl)
	err := templates[tmpl+".tpl"].ExecuteTemplate(w, tmpl+".tpl", c)
	if err != nil {
		log.Printf("ERROR Executing template %s on %s: %s", tmpl, c.Page.Title, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)/?$")

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(dataDir, title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", &Context{Page: p, BaseUrl: "//" + r.Host + "/view/"})
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(dataDir, title)
	if err != nil {
		log.Printf("INFO Creating %s", title)
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", &Context{Page: p})
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeDefaultHandler(fn func(http.ResponseWriter, *http.Request, string), defaultTitle string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("INFO", r.RemoteAddr, r.Proto, r.Method, r.URL, r.UserAgent(), r.Referer())
		title := defaultTitle
		if defaultTitle == "" {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				log.Printf("ERROR Getting title from URL %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}
			title = m[2]
		}
		fn(w, r, title)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return makeDefaultHandler(fn, "")
}

func main() {
	loadTemplates()
	http.HandleFunc("/", makeDefaultHandler(viewHandler, frontPage))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
