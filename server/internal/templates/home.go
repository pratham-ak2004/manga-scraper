package templates

import "html/template"

var PageTemplate = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
<title>File Browser</title>
<style>
	body { font-family: Arial; padding: 20px; }
	table { border-collapse: collapse; width: 60%; }
	td, th { padding: 8px; border-bottom: 1px solid #ccc; }
	a { text-decoration: none; color: blue; }
</style>
</head>
<body>

<h2>Browsing: /{{.CurrentPath}}</h2>

{{if ne .CurrentPath ""}}
	<a href="/">⬅ Back to root</a><br><br>
	<a href="/{{.ParentPath}}">⬆ Go Up</a><br><br>
{{end}}

<table>
<tr><th>Name</th><th>Size</th></tr>

{{range .Entries}}
<tr>
	<td>
		{{if .IsDir}}
			📁 <a href="/{{.Link}}">{{.Name}}</a>
		{{else}}
			📄 <a href="/download/{{.Link}}">{{.Name}}</a>
		{{end}}
	</td>
	<td>{{.Size}}</td>
</tr>
{{end}}

</table>

</body>
</html>
`))
