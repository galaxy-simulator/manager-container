package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/gorilla/mux"

	"git.darknebu.la/GalaxySimulator/structs"
)

var (
	metrics map[string]float64
	mutex   = &sync.Mutex{}
)

type metric struct {
	key   string
	value float64
}

// indexHandler handles incomming requests on the / endpoint
func indexHandler(w http.ResponseWriter, r *http.Request) {
	infostring := `Galaxy Simulator Manager 

API: 
    /calcallforces/{treeindex}
    /getallstars/{treeindex}
    /new/{treewidth}`
	_, _ = fmt.Fprintf(w, infostring)
}

// calcallforcesHandler handles requests on the /calcallforces/{treeindex} handler
func calcallforcesHandler(w http.ResponseWriter, r *http.Request) {
	// get the tree index
	vars := mux.Vars(r)
	treeindex, _ := strconv.ParseInt(vars["treeindex"], 10, 0)

	_, _ = fmt.Fprintf(w, "Got the treeindex, it's %d\n", treeindex)
	fmt.Printf("treeindex: %d\n", treeindex)

	dbUpdateTotalMass(treeindex)
	dbUpdateCenterOfMass(treeindex)

	listofstars := listofstars(treeindex)

	// make a post request to the simulator with every star
	for _, star := range *listofstars {
		fmt.Printf("star: %v", star)
		_, _ = fmt.Fprintf(w, "One of the stars is %v\n", star)

		requesturlSimu := fmt.Sprintf("http://simulator/calcallforces/%d", treeindex)

		response, err := http.PostForm(requesturlSimu, url.Values{
			"x":  {fmt.Sprintf("%f", star.C.X)},
			"y":  {fmt.Sprintf("%f", star.C.Y)},
			"vx": {fmt.Sprintf("%f", star.V.X)},
			"vy": {fmt.Sprintf("%f", star.V.Y)},
			"m":  {fmt.Sprintf("%f", star.M)},
		})
		if err != nil {
			panic(err)
		}

		respBody, readAllErr := ioutil.ReadAll(response.Body)
		if readAllErr != nil {
			panic(readAllErr)
		}
		_, _ = fmt.Fprintf(w, "%s\n", string(respBody))
	}
	_, _ = fmt.Fprintf(w, "DONE!\n")
	fmt.Printf("DONE!\n")
}

// getallstarsHandler handles requests on the /getallstars/{treeindex} endpoint
func getallstarsHandler(w http.ResponseWriter, r *http.Request) {
	// get the tree index
	vars := mux.Vars(r)
	treeindex, _ := strconv.ParseInt(vars["treeindex"], 10, 0)

	listofstars := listofstars(treeindex)

	_, _ = fmt.Fprintf(w, "Got the listofstars, it's %v\n", listofstars)
}

// dbUpdateTotalMass updates the Total mass of each node by recursively iterating over all subnodes
func dbUpdateTotalMass(treeindex int64) {
	fmt.Println("Updating the total mass")
	requestURL := fmt.Sprintf("http://db/updatetotalmass/%d", treeindex)
	fmt.Println(requestURL)
	resp, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("err! %v", err)
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("Done!")
}

// dbUpdateTotalMass updates the Center of mass of each node by recursively iterating over all subnodes
func dbUpdateCenterOfMass(treeindex int64) {
	requestURL := fmt.Sprintf("http://db/updatecenterofmass/%d", treeindex)
	resp, err := http.Get(requestURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

// listofstars returns a pointer to a list of all stars
func listofstars(treeindex int64) *[]structs.Star2D {
	requesturl := fmt.Sprintf("http://db/starlist/%d", int(treeindex))
	fmt.Printf("requesturl: %s\n", requesturl)
	resp, err := http.Get(requesturl)
	fmt.Println("After http GET, but before the error handling")
	if err != nil {
		fmt.Println("PANIC")
		panic(err)
	}
	fmt.Println("After http GET and after error handling")
	defer resp.Body.Close()

	// read the response containing a list of all stars in json format
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("body: %v\n", string(body))

	// unmarshal the json
	listofstars := &[]structs.Star2D{}
	jsonUnmarschalErr := json.Unmarshal(body, listofstars)
	if jsonUnmarschalErr != nil {
		panic(jsonUnmarschalErr)
	}

	fmt.Printf("listofstars: %v\n", *listofstars)
	listofstar := *listofstars

	for i := 0; i < len(*listofstars); i++ {
		fmt.Println(listofstar[i])
	}

	return listofstars
}

// newHandler makes a request to the database container creating a new tree
func newHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	treewidth, _ := strconv.ParseFloat(vars["w"], 64)

	_, _ = fmt.Fprintf(w, "Creating a new tree with a width of %f.\n", treewidth)

	// define where the request should go
	requestURL := "http://db/new"

	// define the request
	response, err := http.PostForm(requestURL, url.Values{
		"w": {fmt.Sprintf("%f", treewidth)},
	})
	if err != nil {
		panic(err)
	}

	// close the response body at the end of the function
	defer response.Body.Close()

	// read the response
	respBody, readAllErr := ioutil.ReadAll(response.Body)
	if readAllErr != nil {
		panic(readAllErr)
	}

	// print the response to the resonse writer
	_, _ = fmt.Fprintf(w, "Response from the Database: %s", string(respBody))
}

// metricHandler handles publishing the metrics and inserting metrics incoming via a POST request into the metrics map
func metricHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("The MetricHandler was accessed")
	mutex.Lock()

	if r.Method == "GET" {
		for key, value := range metrics {
			_, _ = fmt.Fprintf(w, "manager_%s %v\n", key, value)
		}

	} else if r.Method == "POST" {
		log.Println("Incomming POST")

		// parse the POST form
		errParseForm := r.ParseForm()
		if errParseForm != nil {
			panic(errParseForm)
		}

		// parse the parameters
		key := r.Form.Get("key")
		value, _ := strconv.ParseFloat(r.Form.Get("value"), 64)

		if metrics == nil {
			metrics = map[string]float64{}
		}
		metrics[key] = value

		_, _ = fmt.Fprintf(w, "added the metrics")
	}
	mutex.Unlock()
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/calcallforces/{treeindex}", calcallforcesHandler).Methods("GET")
	router.HandleFunc("/getallstars/{treeindex}", getallstarsHandler).Methods("GET")
	router.HandleFunc("/new/{w}", newHandler).Methods("GET")
	router.HandleFunc("/metrics", metricHandler).Methods("GET", "POST")

	fmt.Println("Manager Container up")
	log.Fatal(http.ListenAndServe(":80", router))
}
