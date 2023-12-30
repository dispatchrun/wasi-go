package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	wasiclient "github.com/dev-wasm/dev-wasm-go/http/client"
)

func printResponse(r *http.Response) {
	fmt.Printf("Status: %d\n", r.StatusCode)
	for k, v := range r.Header {
		fmt.Printf("%s: %s\n", k, v[0])
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(4)
	}
	fmt.Printf("Body: \n%s\n", body)
}

func main() {
	server := os.Getenv("SERVER")
	client := http.Client{
		Transport: wasiclient.WasiRoundTripper{},
	}
	res, err := client.Get("http://" + server + "/get?some=arg&goes=here")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	printResponse(res)
	res.Body.Close()

	res, err = client.Post("http://"+server+"/post", "application/json", wasiclient.BodyReaderCloser([]byte("{\"foo\": \"bar\"}")))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	printResponse(res)
	res.Body.Close()

	res, err = wasiclient.Put(&client, "http://"+server+"/put", "application/json", wasiclient.BodyReaderCloser([]byte("{\"baz\": \"blah\"}")))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}
	printResponse(res)
	res.Body.Close()

	os.Exit(0)
}
