package main

import (
	"log"
	"net/http"
	"html/template"
	"io/ioutil"
	"regexp"
)

type ArticlePage struct {
	Title string
	Body []byte
}

type IndexPage struct {
	Title string
	FilesList []string
}

var templates = template.Must(template.ParseGlob("./templates/*"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// Write text to a file in the current directory
func (p *ArticlePage) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// Read an existing file in current directory
func loadPage(title string) (*ArticlePage, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return &ArticlePage{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *ArticlePage) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Load existing page in viewer or create a new one in editor
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Load existing page or create new one in editor
	p, err := loadPage(title)
	if err != nil {
		p = &ArticlePage{Title: title}
	}

	// Render edit page from template
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Save form data
	body := r.FormValue("body")
	p := &ArticlePage{Title: title, Body: []byte(body)}
	err := p.save()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to new wiki page
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

	// 
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check title with regex
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}

		// Call function passed to makeHandler()
		fn(w, r, m[2])
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Read files in data directory
	files, err := ioutil.ReadDir("./data/")
	if err != nil {
		log.Fatal(err)
	}

	// Store filenames minus the extension in string slice
	var filenames []string

	for _, file := range files {
		filenames = append(filenames, file.Name()[:len(file.Name())-4] )
	}
	
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	
	// Compile template with IndexPage struct
	err = t.Execute(w, IndexPage{"Wiki index", filenames})
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
