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

// Page represents an article
type Page struct {
	Title string
	Body  []byte
}

const frontPage = "FrontPage"
const dataDir = "data"
const templatesDir = "templates"

func filename(dataDir, title string) string {
	return path.Join(dataDir, html.EscapeString(title)+".md")
}

// Save makes changes persistent
func (p *Page) Save(dataDir string) error {
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

var pageTitle = regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")

func linkPages(body []byte, baseURL string) template.HTML {
	return template.HTML(pageTitle.ReplaceAllFunc(body, func(title []byte) []byte {
		return []byte(fmt.Sprintf("<a href=\"%s%s\">%s</a>", baseURL, title[1:len(title)-1], title[1:len(title)-1]))
	}))
}

func getFileNamesExcept(pattern string, except []string) []string {
	m := make(map[string]struct{})
	for _, v := range except {
		m[v] = struct{}{}
	}
	allfiles, _ := filepath.Glob(pattern)
	files := make([]string, 0)
	for _, v := range allfiles {
		basename := filepath.Base(v)
		name := strings.TrimSuffix(basename, filepath.Ext(basename))
		if _, invalid := m[name]; invalid {
			continue
		}
		files = append(files, name)
	}
	return files
}

var indexTemplate = template.Must(template.New("main").Parse("{{if .Files}}<ul>{{range .Files}}<li><a href=\"{{ $.BaseURL }}{{.}}\">{{.}}</a></li>{{end}}</ul>{{else}}No pages{{end}}"))

func autoIndex(dataDir, baseURL string) template.HTML {
	context := struct {
		Files   []string
		BaseURL string
	}{
		getFileNamesExcept(path.Join(dataDir, "*.md"), []string{"", frontPage}),
		baseURL,
	}
	var result bytes.Buffer
	indexTemplate.Execute(&result, context)
	return template.HTML(result.String())
}

var templates map[string]*template.Template

func getFiles(pattern string) []string {
	files, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}
	return files
}

func getTemplateFiles(dir string) map[string][]string {
	layoutFiles := getFiles(path.Join(dir, "*.html"))
	includeFiles := getFiles(path.Join(dir, "/*.tpl"))

	var files = make(map[string][]string, 0)
	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files[fileName] = append(layoutFiles, file)
	}
	return files
}

func loadTemplates() map[string]*template.Template {
	mainTemplate := template.New("main").Funcs(template.FuncMap{
		"linkPages": linkPages,
		"autoIndex": func(baseURL string) template.HTML {
			return autoIndex(dataDir, baseURL)
		},
	})

	var templates = make(map[string]*template.Template)
	for fileName, files := range getTemplateFiles(templatesDir) {
		templates[fileName] = template.Must(template.Must(mainTemplate.Clone()).ParseFiles(files...))
	}
	return templates
}

func renderTemplate(w http.ResponseWriter, tmpl string, page *Page, baseURL string) {
	log.Printf("DEBUG Executing template %s on %s with baseURL: %s", tmpl, page.Title, baseURL)
	context := struct {
		Page    *Page
		BaseURL string
	}{
		page,
		baseURL,
	}
	err := templates[tmpl+".tpl"].ExecuteTemplate(w, tmpl+".tpl", context)
	if err != nil {
		log.Printf("ERROR Executing template %s on %s: %s", tmpl, page.Title, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)/?$")

func viewHandler(w http.ResponseWriter, r *http.Request, dataDir string, title string) {
	p, err := loadPage(dataDir, title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p, "//"+r.Host+"/view/")
}

func editHandler(w http.ResponseWriter, r *http.Request, dataDir string, title string) {
	p, err := loadPage(dataDir, title)
	if err != nil {
		log.Printf("INFO Creating %s", title)
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p, "")
}

func saveHandler(w http.ResponseWriter, r *http.Request, dataDir string, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.Save(dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeDefaultHandler(fn func(http.ResponseWriter, *http.Request, string, string), defaultTitle string) http.HandlerFunc {
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
		fn(w, r, dataDir, title)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return makeDefaultHandler(fn, "")
}

func main() {
	templates = loadTemplates()
	http.HandleFunc("/", makeDefaultHandler(viewHandler, frontPage))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
