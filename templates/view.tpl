{{ template "layout" . }}
{{define "title"}}View {{.Page.Title}}{{end}}
{{define "content"}}
<h1>{{.Page.Title}}</h1>
<p>[<a href="/">front page</a>] [<a href="/edit/{{.Page.Title}}">edit</a>]</p>
<div>{{linkPages .Page.Body .BaseUrl}}</div>
{{end}}
