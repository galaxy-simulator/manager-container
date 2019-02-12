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
	metrics           map[string]float64
	mutex                   = &sync.Mutex{}
	starBufferChannel       = make(chan structs.Stargalaxy, 1000000)
	currentStarBuffer int64 = 0
)

// struct bundling the star and the galaxy index it comes from
// type stargalaxy struct {
// 	star  structs.Star2D
// 	index int64
// }

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
	log.Println("[   ] The getallstars Handler was accessed")

	// get the tree index
	vars := mux.Vars(r)
	treeindex, _ := strconv.ParseInt(vars["treeindex"], 10, 0)

	// get a pointer to the list of all stars
	listOfStars := listofstars(treeindex)

	_, _ = fmt.Fprintf(w, "Got the listOfStars, it's %v\n", listOfStars)
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

	// define and make the request
	requesturl := fmt.Sprintf("http://db/starlist/%d", int(treeindex))
	resp, err := http.Get(requesturl)
	if err != nil {
		fmt.Println("PANIC")
		panic(err)
	}
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

	return listofstars
}

// newHandler makes a request to the database container creating a new tree creating a new tree
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

	// if the request is a GET request, return the metrics map element wise
	if r.Method == "GET" {
		for key, value := range metrics {
			_, _ = fmt.Fprintf(w, "manager_%s %v\n", key, value)
		}

		// if the request is a POST request, insert the incoming metric into the map
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

// getstarHandler returns a single star (in json) on which the forces still need to be calculated
// it returns a json object containing the star and the galaxy index the star is from:
//
// {{{x, y},{vx, vy}, m}, index}
//
// This handler is used by the simulators to obtains stars from the manager and enables them to work all the time
// without needing to wait for the manager to provide star to process
func getstarHandler(w http.ResponseWriter, r *http.Request) {
	// get a stargalaxy (struct containing a star and it's galaxy index) from the starBufferChannel
	star := <-starBufferChannel

	_, _ = fmt.Fprintf(w, "%v", star)
}

// fillStarBufferChannel fills the starBufferChannel with all the stars from the next timestep
func fillStarBufferChannel() {
	fmt.Println("Filling the star buffer channel")

	// get the list of stars using the currentStarBuffer value
	// the currentStarBuffer value is a counter keeping track of which galaxy is going to
	// be inserted into the StarBufferChannel next
	listofstars := listofstars(currentStarBuffer)

	// insert all the stars from the list of stars into the starBufferChannel in the form of
	// a star galaxy, a type combining the star and the galaxy it belongs into into a single type
	for _, star := range *listofstars {
		starBufferChannel <- structs.Stargalaxy{star, currentStarBuffer}
	}

	// increase the currentStarBuffer counter
	currentStarBuffer += 1
}

// providestarsHandler returns a single star for usage by the simulator container
func providestarsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Providing stars from the StarBufferHandler")

	// if there are no stars in the star buffer channel, fill the star buffer channel with stars
	if len(starBufferChannel) == 0 {
		fillStarBufferChannel()
	}

	// get a single star from the starBufferChannel and marshal it to json
	stargalaxy := <-starBufferChannel
	marshaledStargalaxy, _ := json.Marshal(stargalaxy)

	// return the star
	_, _ = fmt.Fprintf(w, "%v", string(marshaledStargalaxy))

	fmt.Println("Done providing stars from the StarBufferHandler")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/calcallforces/{treeindex}", calcallforcesHandler).Methods("GET")
	router.HandleFunc("/getallstars/{treeindex}", getallstarsHandler).Methods("GET")
	router.HandleFunc("/new/{w}", newHandler).Methods("GET")
	router.HandleFunc("/metrics", metricHandler).Methods("GET", "POST")
	router.HandleFunc("/getstar", getstarHandler).Methods("GET")

	router.HandleFunc("/providestars/{treeindex}", providestarsHandler).Methods("GET")

	fmt.Println("Manager Container up")
	log.Fatal(http.ListenAndServe(":80", router))
}
