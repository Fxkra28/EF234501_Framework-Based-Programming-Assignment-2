package main

import (
	"html/template" 
	"net/http"
	"os"
	"regexp"
	"log"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt" //Store in data
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt" //Load from data
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

//Regex to find [PageName] links
var linkRegex = regexp.MustCompile(`\[([a-zA-Z0-9]+)\]`)

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	//Process body for [PageName] links
	//Replaces [PageName] with <a href="/view/PageName">PageName</a>
	processedBody := linkRegex.ReplaceAll(p.Body, []byte(`<a href="/view/$1">$1</a>`))

	//Create a struct to pass data to the view template.
	//i use template.HTML to tell the template engine that
	viewData := struct {
		Title string
		Body  template.HTML
	}{
		Title: p.Title,
		Body:  template.HTML(processedBody),
	}

	renderTemplate(w, "view", viewData)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	//Pass the raw *Page struct to the edit template
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//Parse templates from the tmpl/ directory
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

//renderTemplate now accepts interface{} to be more flexible
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

//New handler for the root URL
func rootHandler(w http.ResponseWriter, r *http.Request) {
	//If the path is not exactly "/", show a 404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	//Redirect "/" to the FrontPage
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", rootHandler) //Add the new root handler

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}