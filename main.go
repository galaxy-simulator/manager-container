package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

type Page struct {
	NewTree bool
	X, Y, W float64
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(
		"tmpl/overview.html",
		"tmpl/header.html",
		"tmpl/nav.html",
		"tmpl/javascript.html",
	))
	templateExecuteError := t.ExecuteTemplate(w, "overview", &Page{})
	if templateExecuteError != nil {
		log.Println(templateExecuteError)
	}
}

func newtreeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("the new tree handler was accessed")
	// if the method used to get the page is GET, simply respond with the tree-creation form
	if r.Method == "GET" {
		log.Println("using a GET method")
		t := template.Must(template.ParseFiles(
			"tmpl/newtree.html",
			"tmpl/header.html",
			"tmpl/nav.html",
			"tmpl/javascript.html",
		))
		templateExecuteError := t.ExecuteTemplate(w, "newtree", &Page{})
		if templateExecuteError != nil {
			log.Println(templateExecuteError)
		}
	} else {
		log.Println("using a POST method")

		// get the values from the form
		err := r.ParseForm()
		log.Println(err)
		log.Println("parsed the form")

		x, _ := strconv.ParseFloat(r.Form["x"][0], 64)     // x position of the midpoint of the tree
		y, _ := strconv.ParseFloat(r.Form["y"][0], 64)     // y position of the midpoint of the tree
		width, _ := strconv.ParseFloat(r.Form["w"][0], 64) // width the the tree
		dbip := r.Form["db-ip"][0]                         // ip address of the database

		log.Println("parsed the form arguments")

		// parse the template files
		t := template.Must(template.ParseFiles(
			"tmpl/newtree.html",
			"tmpl/header.html",
			"tmpl/nav.html",
			"tmpl/javascript.html",
		))
		log.Println("parsed the template file")

		// execute the template
		templateExecuteError := t.ExecuteTemplate(w, "newtree", &Page{NewTree: true, X: x, Y: y, W: width})

		log.Println("executed the template")

		// handle potential errors
		if templateExecuteError != nil {
			log.Println(templateExecuteError)
		}

		log.Println("handled potential errors")

		// create the new tree by accessing the api endpoint of the database
		postUrl := fmt.Sprintf("http://%s/new", dbip)

		log.Printf("the postURL: %s", postUrl)
		log.Println("created the url for the post request")

		// define the data that should be posted
		formData := url.Values{
			"x": r.Form["x"],
			"y": r.Form["y"],
			"w": r.Form["w"],
		}

		log.Println("created the form to be sent in the post request")

		// make the http Post request
		resp, err := http.PostForm(postUrl, formData)

		log.Println("sent the post request")

		// some debug messages
		log.Printf("[   ] POST response: %v", resp)
		log.Printf("[   ] POST err: %v", err)
	}
}

func addstarsHandler(w http.ResponseWriter, r *http.Request) {
	// if the method used to get the page is GET, simply respond with the add-stars form
	if r.Method == "GET" {
		log.Println("/addstars was accessed using a GET method")
		t := template.Must(template.ParseFiles(
			"tmpl/addstars.html",
			"tmpl/header.html",
			"tmpl/nav.html",
			"tmpl/javascript.html",
		))
		templateExecuteError := t.ExecuteTemplate(w, "addstars", &Page{})
		if templateExecuteError != nil {
			log.Println(templateExecuteError)
		}
	} else {
		log.Println("/addstars was accessed using a POST method")

		// get the values from the form
		_ = r.ParseForm()

		log.Println("parsed the form")

		// extract some information from the form
		dbip := r.Form["dbip"][0]                                   // ip address of the database
		treeindex, _ := strconv.ParseInt(r.Form["tree"][0], 10, 64) // index of the tree
		log.Println(dbip)
		log.Println(treeindex)

		log.Println("parsed the form arguments")

		t := template.Must(template.ParseFiles(
			"tmpl/addstars.html",
			"tmpl/header.html",
			"tmpl/nav.html",
			"tmpl/javascript.html",
		))

		log.Println("parsed the template file")

		templateExecuteError := t.ExecuteTemplate(w, "addstars", &Page{})
		if templateExecuteError != nil {
			log.Println(templateExecuteError)
		}

		log.Println("executed the template")

		// create the new tree by accessing the api endpoint of the database
		postUrl := fmt.Sprintf("http://%s/insert/%d", dbip, treeindex)

		log.Printf("the postURL: %s", postUrl)
		log.Println("created the url for the post request")

		// define the data that should be posted
		formData := url.Values{
			"x":  r.Form["x"],
			"y":  r.Form["y"],
			"vx": r.Form["vx"],
			"vy": r.Form["vy"],
			"m":  r.Form["mass"],
		}

		log.Println("created the form to be sent in the post request")

		// make the http Post request
		resp, err := http.PostForm(postUrl, formData)

		log.Println("sent the post request")

		// some debug messages
		log.Printf("[   ] POST response: %v", resp)
		log.Printf("[   ] POST err: %v", err)
	}
}

func drawtreeHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", overviewHandler).Methods("GET")
	router.HandleFunc("/newtree", newtreeHandler).Methods("GET", "POST")
	router.HandleFunc("/addstars", addstarsHandler).Methods("GET", "POST")
	router.HandleFunc("/drawtree", drawtreeHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8429", router))
}
