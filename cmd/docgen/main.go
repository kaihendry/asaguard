package main

import (
	"html/template"
	"log"
	"os"
	"os/exec"
)

const modulePath = "github.com/kaihendry/asaguard/internal/"

var packages = []string{
	"hooks",
	"mcps",
	"perms",
	"policy",
	"result",
	"scorer",
	"secrets",
	"settings",
	"siem",
	"transcripts",
	"updater",
}

const pkgTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{.Name}} — asaguard</title>
<style>
body{font-family:system-ui,sans-serif;max-width:860px;margin:2rem auto;padding:0 1rem;color:#111}
h1{margin-bottom:.25rem}
nav{margin-bottom:1.5rem}
pre{background:#f5f5f5;padding:1rem;overflow-x:auto;border-radius:4px;font-size:.875rem;white-space:pre-wrap;word-break:break-word}
a{color:#0066cc}
</style>
</head>
<body>
<nav><a href="index.html">← asaguard guard rails</a></nav>
<h1>{{.Name}}</h1>
<pre>{{.Doc}}</pre>
</body>
</html>`

const idxTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>asaguard guard rails</title>
<style>
body{font-family:system-ui,sans-serif;max-width:860px;margin:2rem auto;padding:0 1rem;color:#111}
h1{margin-bottom:.25rem}
p{color:#555}
ul{line-height:2}
a{color:#0066cc}
</style>
</head>
<body>
<h1>asaguard guard rails</h1>
<p>Auto-generated from <code>go doc -all</code>. Run <code>make docs</code> to regenerate.</p>
<ul>
{{range .}}<li><a href="{{.Name}}.html">{{.Name}}</a></li>
{{end}}</ul>
</body>
</html>`

type pkgData struct {
	Name string
	Doc  string
}

func godoc(pkg string) (string, error) {
	out, err := exec.Command("go", "doc", "-all", modulePath+pkg).CombinedOutput()
	return string(out), err
}

func ensureDocsDir() error {
	return os.MkdirAll("docs", 0755)
}

func main() {
	if err := ensureDocsDir(); err != nil {
		log.Fatalf("create docs dir: %v", err)
	}

	pkgTpl := template.Must(template.New("pkg").Parse(pkgTmpl))
	idxTpl := template.Must(template.New("idx").Parse(idxTmpl))

	var index []pkgData
	for _, pkg := range packages {
		doc, err := godoc(pkg)
		if err != nil {
			log.Printf("warning: go doc for %s: %v", pkg, err)
		}
		data := pkgData{Name: pkg, Doc: doc}
		index = append(index, data)

		f, err := os.Create("docs/" + pkg + ".html")
		if err != nil {
			log.Fatalf("create %s.html: %v", pkg, err)
		}
		if err := pkgTpl.Execute(f, data); err != nil {
			log.Fatalf("render %s.html: %v", pkg, err)
		}
		f.Close()
		log.Printf("wrote docs/%s.html", pkg)
	}

	f, err := os.Create("docs/index.html")
	if err != nil {
		log.Fatalf("create index.html: %v", err)
	}
	defer f.Close()
	if err := idxTpl.Execute(f, index); err != nil {
		log.Fatalf("render index.html: %v", err)
	}
	log.Println("wrote docs/index.html")
}
