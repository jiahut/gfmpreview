package main

import (
	"bytes"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/github_flavored_markdown/gfmstyle"
	"github.com/vrischmann/hutil"
)

func pathForName(paths []string, path string) string {
	name := filepath.Base(path)
	for _, el := range paths {
		if strings.HasSuffix(el, name) {
			return el
		}
	}
	return ""
}

func renderList(w http.ResponseWriter, paths []string) {
	const tpl = `
<!doctype>
<html>
<head><title>GitHub Flavored Markdown Preview</title></head>
<body>
<ul>
	{{ range . }}<li><a href="/{{ . }}">{{ . }}</a></li>{{ end }}
</ul>
</body>
</html>`

	tmpl := template.Must(template.New("root").Parse(tpl))

	links := make([]string, len(paths))
	for i, el := range paths {
		links[i] = filepath.Base(el)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, links); err != nil {
		hutil.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	io.Copy(w, &buf)
}

func renderMarkdownHTML(w http.ResponseWriter, html []byte) {
	const tpl = `
<!doctype>
<html>
<head>
	<meta charset="utf-8">
	<link href="/assets/gfm.css" media="all" rel="stylesheet" type="text/css" />
	<link href="//cdnjs.cloudflare.com/ajax/libs/octicons/2.1.2/octicons.css" media="all" rel="stylesheet" type="text/css" />
</head>
<body>
	<article class="markdown-body entry-content" style="padding: 30px;">{{ . }}</article>
</body>
</html>`

	tmpl := template.Must(template.New("root").Parse(tpl))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, template.HTML(html)); err != nil {
		hutil.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	io.Copy(w, &buf)
}

func main() {
	var (
		flListenAddr = flag.String("l", ":3030", "The HTTP listen address")
	)

	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var paths []string
	err = filepath.Walk(cwd, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(fi.Name(), ".md") {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	//

	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(gfmstyle.Assets)))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			renderList(w, paths)
			return
		}

		path := pathForName(paths, req.URL.Path)
		if path == "" {
			hutil.WriteText(w, http.StatusNotFound, "Not Found")
			return
		}

		markdown, err := ioutil.ReadFile(path)
		if err != nil {
			hutil.WriteError(w, err)
			return
		}

		data := github_flavored_markdown.Markdown(markdown)
		renderMarkdownHTML(w, data)
	})

	log.Printf("all markdown files under %s are served at %s", cwd, *flListenAddr)

	http.ListenAndServe(*flListenAddr, nil)
}
