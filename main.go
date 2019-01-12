package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"git.darknebu.la/GalaxySimulator/structs"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	infostring := `Galaxy Simulator Manager 

API: 
	/calcallforces/{treeindex}`
	_, _ = fmt.Fprintf(w, infostring)
}

func calcallforcesHandler(w http.ResponseWriter, r *http.Request) {
	// get the tree index
	vars := mux.Vars(r)
	treeindex, _ := strconv.ParseInt(vars["treeindex"], 10, 0)

	_, _ = fmt.Fprintf(w, "Got the treeindex, it's %d\n", treeindex)
	fmt.Printf("treeindex: %d\n", treeindex)

	// ################################################################################
	// Calculate the center of mass and total mass of each tree
	// ################################################################################

	// db.updatestufff()

	// ################################################################################
	// Get a list of all stars in the tree
	// ################################################################################
	// make the following post request:
	// db.docker.localhost/getallstars/{treeindex}
	requesturl := fmt.Sprintf("http://db.docker.localhost/starlist/%d", treeindex)
	_, _ = fmt.Fprintf(w, "Got the requesturl, it's %s\n", requesturl)
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
	_, _ = fmt.Fprintf(w, "Got the body, it's %v\n", string(body))
	fmt.Printf("body: %v\n", string(body))

	// unmarshal the json
	listofstars := &[]structs.Star2D{}
	jsonUnmarschalErr := json.Unmarshal(body, listofstars)
	if jsonUnmarschalErr != nil {
		panic(jsonUnmarschalErr)
	}

	fmt.Printf("listofstars: %v\n", listofstars)
	_, _ = fmt.Fprintf(w, "Got the listofstars, it's %v\n", listofstars)
	fmt.Printf("listofstars: %v\n", listofstars)

	// ################################################################################
	// Send the star to the simulator in order to simulate it's new position
	// ################################################################################

	// make a post request to the simulator with every star
	for _, star := range *listofstars {
		fmt.Printf("star: %v", star)
		_, _ = fmt.Fprintf(w, "One of the stars is %v\n", star)

		requesturlSimu := fmt.Sprintf("http://simulator.docker.localhost/calcallforces/%d", treeindex)

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

func printhiHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hi there!")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/calcallforces/{treeindex}", calcallforcesHandler).Methods("GET")
	router.HandleFunc("/printhi", printhiHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8031", router))
}
