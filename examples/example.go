package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"

	"github.com/ppc64le-cloud/manageiq-client-go"
)

func main() {
	a := &manageiq.BasicAuthenticator{
		UserName: "admin",
		Password: "smartvm",
		BaseURL:  "https://127.0.0.1:8443/api",
		Insecure: true,
	}
	//g, err := manageiq.NewClient(a, manageiq.ClientParams{}).GetGroup("2")
	//if err != nil {
	//	log.Printf("errored getting services: %+v", err)
	//}
	//spew.Dump(g)
	s, err := manageiq.NewClient(a, manageiq.ClientParams{}).GetServiceCatalogs()
	if err != nil {
		log.Printf("errored getting services: %+v", err)
	}
	spew.Dump(s)
	//req, err := http.NewRequest("POST", "github.com", nil)
	//if err != nil {
	//	log.Printf("errored creating http request: %v", err)
	//}
	//err = a.Authenticate(req)
	//if err != nil {
	//	log.Printf("failed to auth: %v", err)
	//}
	//spew.Dump(req)
	//fmt.Println("hello")
}
