package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/github_flavored_markdown/gfmstyle"
	"github.com/vrischmann/hutil"
)

type paths struct {
	cwd   string
	paths []string
}

func (p *paths) getRealPath(reqPath string) string {
	for _, el := range p.paths {
		if !strings.HasSuffix(el, reqPath) {
			continue
		}

		return el
	}
	return ""
}

func (p *paths) getRelativePaths() ([]string, error) {
	paths := make([]string, len(p.paths))
	for i, el := range p.paths {
		rel, err := filepath.Rel(p.cwd, el)
		if err != nil {
			return nil, err
		}
		paths[i] = rel
	}
	return paths, nil
}

func renderList(w http.ResponseWriter, p paths) {
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

	items, err := p.getRelativePaths()
	if err != nil {
		hutil.WriteError(w, err)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, items); err != nil {
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

	p := paths{cwd: cwd}
	err = filepath.Walk(cwd, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(fi.Name(), ".md") {
			p.paths = append(p.paths, path)
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
			renderList(w, p)
			return
		}

		path := p.getRealPath(req.URL.Path)
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

	fmt.Printf("all markdown files under %s are served at %s\n", cwd, *flListenAddr)

	addr := strings.Split(*flListenAddr, ":")
	local := fmt.Sprintf("http://127.0.0.1:%s", addr[1])
	if err := exec.Command("open", local).Run(); err != nil {
		log.Fatal(err)
	}

	if err := http.ListenAndServe(*flListenAddr, nil); err != nil {
		fmt.Printf("unable to listen. err=%v\n", err)
		os.Exit(1)
	}

}
