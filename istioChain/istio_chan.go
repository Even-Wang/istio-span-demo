package main

import (
	"net/http"
	"os"
	"io"
	"fmt"
	"log"
	"io/ioutil"
)




//test1
//with headspan

func NextHandler(w http.ResponseWriter, r * http.Request) {
	var NEXTURL,LOCATION,VERSION string
	NEXTURL = os.Getenv("NEXT_URL")
	LOCATION = os.Getenv("LOCATION")
	VERSION = os.Getenv("VERSION")
	if NEXTURL ==""{
		finfo := "unfound the "+LOCATION+"th NEXTURL"
		io.WriteString(w, finfo)
		fmt.Println(finfo)
	}else {
		v1 := buidResInfo(r, NEXTURL, LOCATION, VERSION)
		ten, err := nextChanl(NEXTURL, r, true)
		if err != nil {
			err = fmt.Errorf("err when get the %sth withspan's nextchan：%s", LOCATION,err)
			fmt.Println(err)
		}
		io.WriteString(w, v1+ten)
	}

}

func NextHandlerWithoutSpan(w http.ResponseWriter, r * http.Request) {
	var NEXTURL,LOCATION,VERSION string
	NEXTURL = os.Getenv("NEXT_URL")
	LOCATION = os.Getenv("LOCATION")
	VERSION = os.Getenv("VERSION")
	if NEXTURL ==""{
		finfo := "unfound the NEXTURL"
		io.WriteString(w, finfo)
		fmt.Println(finfo)
	}else {
		v1 := buidResInfo(r, NEXTURL, LOCATION, VERSION)
		ten, err := nextChanl(NEXTURL, r, false)
		if err != nil {
			err = fmt.Errorf("err when get the %sth withoutspan's nextchan：%s", LOCATION,err)
			fmt.Println(err)
		}
		io.WriteString(w, v1+ten)
	}
}

func buidResInfo (r * http.Request,NEXTURL,LOCATION,VERSION string) string{
	svc := "the x-request-id:"+ r.Header.Get("x-request-id") + "\n"
	svc2 := "the x-b3-traceid:"+r.Header.Get("x-b3-traceid") + "\n"
	svc3 := "the x-b3-spanid:"+r.Header.Get("x-b3-spanid") + "\n"
	svc4 := "the x-b3-parentspanid:"+r.Header.Get("x-b3-parentspanid") + "\n"
	base1 := "NEXTURL:"+ NEXTURL + "\n"
	//base2 := "LOCATION:"+ LOCATION + "\n"
		base3 := "VERSION:"+ VERSION + "\n"
	v1 := fmt.Sprintf("######\nthe %sth resp \n######\n" +
		"%s%s##############\n%s%s%s%s",LOCATION,base3,base1,svc,svc2,svc3,svc4)
	return v1
}

func nextChanl(url string,breq *http.Request,bar bool) (string,error){
	client := &http.Client{}
	req, err := http.NewRequest("", url, nil)
	if err!=nil{
		err = fmt.Errorf("next chan err：%s", err)
		fmt.Println(err)
	}

	if bar {
		for key,val :=range getheaderl(breq){
			req.Header.Add(key,val)
		}

	}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("get http err：%s", err)
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	val := fmt.Sprintf("%s", body)
	return val, nil
}

func getheaderl(request *http.Request)  map[string]string{
	//var headtest map[string]string
	headtest := make(map[string]string)
	//if 'user' in session:
	//headers['end-user'] = session['user']

	incoming_headers := []string{
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-ot-span-context"}
	for _,bval :=range incoming_headers{
		val := request.Header.Get(bval)
		if val != ""{
			headtest[bval] = val
		}

	}

	return headtest
}

func main() {
	htt := http.HandlerFunc(NextHandler)
	if htt != nil {
		http.Handle("/echo", htt)
	}
	htt2 := http.HandlerFunc(NextHandlerWithoutSpan)
	if htt2 != nil {
		http.Handle("/woecho", htt2)
	}

	err := http.ListenAndServe(":8096", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

