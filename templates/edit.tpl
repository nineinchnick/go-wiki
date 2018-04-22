{{ template "layout" . }}
{{define "title"}}Edit {{.Page.Title}}{{end}}
{{define "content"}}
<h1>Editing {{.Page.Title}}</h1>
<p>[<a href="/">front page</a>] [<a href="/view/{{.Page.Title}}">cancel</a>]</p>
<form action="/save/{{.Page.Title}}" method="POST">
    <div><textarea name="body" rows="20" cols="80">{{printf "%s" .Page.Body}}</textarea></div>
    <div><input type="submit" value="Save"></div>
</form>
{{end}}
