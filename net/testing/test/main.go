package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	resp, err := http.Get("https://api.github.com/repos/btcsuite/btcd/commits?client_id=80386779008eea5dab41&client_secret=f2086aebf790729026fb209b803010029821d8a3")

	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		// handle error
	}
	var f interface{}
	json.Unmarshal(body, &f)

	for _, n := range f.([]interface{}) {
		for k, v := range n.(map[string]interface{}) {
			if k == "commit" {
				value := v.(map[string]interface{})["message"].(string)
				fmt.Println(value)
				if strings.Contains(value, "gx publish ") {
					url := v.(map[string]interface{})["url"].(string)
					fmt.Println(v.(map[string]interface{})["message"])
					temps := strings.Split(url, "/")
					fmt.Println(temps[len(temps)-1])

					str := "github.com/agl/ed25519"
					temps = strings.Split(str, "ghub.com")
					fmt.Println(len(temps))
					fmt.Println(temps)
					return
				}
			}
		}
	}

}
