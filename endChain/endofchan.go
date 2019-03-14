package main

import (
	"net/http"
	"io"
	"fmt"
	"log"
)

func helloHandler(w http.ResponseWriter, r * http.Request) {
	svc := "the lastpoint header x-request-id:"+ r.Header.Get("x-request-id") + "\n"
	svc2 := "the lastpoint header x-b3-traceid:"+r.Header.Get("x-b3-traceid") + "\n"
	svc3 := "the lastpoint header x-b3-spanid:"+r.Header.Get("x-b3-spanid") + "\n"
	svc4 := "the lastpoint header x-b3-parentspanid:"+r.Header.Get("x-b3-parentspanid") + "\n"
	v1 := fmt.Sprintf("##############\nthe lastpoint resp\n############%s%s%s%s",svc,svc2,svc3,svc4)
	io.WriteString(w, v1 )


}
func main() {
	ht := http.HandlerFunc(helloHandler)
	if ht != nil {
		http.Handle("/echo", ht)
	}

	err := http.ListenAndServe(":8091", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

