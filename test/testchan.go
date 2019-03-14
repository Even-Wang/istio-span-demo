package main

import
(
	"net/http"
	"io"
	"fmt"
	"log"

	"io/ioutil"
)

func getheader(request *http.Request)  map[string]string{
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



var iCnt int = 0;


//test1
//with headspan

func helloNextHandler(w http.ResponseWriter, r * http.Request) {

	url := "http://localhost:8096/nextwispan2"
	svc := "the firpoint header x-request-id:"+ r.Header.Get("x-request-id")+ "\n"
	svc2 := "the firpoint header x-b3-traceid:"+r.Header.Get("x-b3-traceid")+ "\n"
	svc3 := "the firpoint header x-b3-spanid:"+r.Header.Get("x-b3-spanid")+ "\n"
	svc4 := "the firpoint header x-b3-parentspanid:"+r.Header.Get("x-b3-parentspanid")+ "\n"
	v1 := fmt.Sprintf("the first withspan resp\n##########################\n%s%s%s%s",svc,svc2,svc3,svc4)
	ten,err:= nextChan(url,r,true)
	if err != nil{
		err = fmt.Errorf("err when get the first withspan's nextchan：%s", err)
		fmt.Println(err)
	}
	io.WriteString(w, v1 + ten)

}
func helloNextHandler2(w http.ResponseWriter, r * http.Request) {

	url := "http://localhost:8091/hello"
	svc := "the sepoint header x-request-id:" + r.Header.Get("x-request-id") + "\n"
	svc2 := "the sepoint header x-b3-traceid:" + r.Header.Get("x-b3-traceid") + "\n"
	svc3 := "the sepoint header x-b3-spanid:" + r.Header.Get("x-b3-spanid") + "\n"
	svc4 := "the sepoint header x-b3-parentspanid:" + r.Header.Get("x-b3-parentspanid") + "\n"

	v1 := fmt.Sprintf("the second withspan resp\n##########################\n%s%s%s%s",svc,svc2,svc3,svc4)
	ten,err:= nextChan(url,r,true)
	if err != nil{
		err = fmt.Errorf("err when get the second withspan's nextchan：%s", err)
		fmt.Println(err)
	}
	io.WriteString(w, v1+ten)

}


func helloNextHandlerWithoutSpan(w http.ResponseWriter, r * http.Request) {
	url := "http://localhost:8096/nextwospan2"
	//fmt.Printf("the fir point header of the chan without span:%s",r.Header)
	svc := "the firpoint header x-request-id:"+ r.Header.Get("x-request-id") + "\n"
	svc2 := "the firpoint header x-b3-traceid:"+r.Header.Get("x-b3-traceid") + "\n"
	svc3 := "the firpoint header x-b3-spanid:"+r.Header.Get("x-b3-spanid") + "\n"
	svc4 := "the firpoint header x-b3-parentspanid:"+r.Header.Get("x-b3-parentspanid") + "\n"

	v1 := fmt.Sprintf("the first withoutspan resp\n##########################\n%s%s%s%s",svc,svc2,svc3,svc4)
	ten,err:= nextChan(url,r,true)
	if err != nil{
		err = fmt.Errorf("err when get the first withoutspan's nextchan：%s", err)
		fmt.Println(err)
	}
	io.WriteString(w, v1+ten)

}

func helloNextHandlerWithoutSpan3(w http.ResponseWriter, r * http.Request) {

	url := "http://localhost:8091/hello"

	svc := "the se point header x-request-id:"+ r.Header.Get("x-request-id") + "\n"
	svc2 := "the se point header x-b3-traceid:"+r.Header.Get("x-b3-traceid") + "\n"
	svc3 := "the se point header x-b3-spanid:"+r.Header.Get("x-b3-spanid") + "\n"
	svc4 := "the se point header x-b3-parentspanid:"+r.Header.Get("x-b3-parentspanid") + "\n"

	v1 := fmt.Sprintf("the second withoutspan resp\n##########################\n%s%s%s%s",svc,svc2,svc3,svc4)
	ten,err:= nextChan(url,r,true)
	if err != nil{
		err = fmt.Errorf("err when get the second withoutspan's nextchan：%s", err)
		fmt.Println(err)
	}
	io.WriteString(w, v1+ten)

}

func nextChan(url string,breq *http.Request,bar bool) (string,error){
	client := &http.Client{}
	req, err := http.NewRequest("", url, nil)
	if err!=nil{
		err = fmt.Errorf("next chan err：%s", err)
		fmt.Println(err)
		}

	if bar {
		for key,val :=range getheader(breq){
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

func main() {
	ht2 := http.HandlerFunc(helloNextHandler)
	if ht2 != nil {
		http.Handle("/nextwispan", ht2)
	}
	ht22 := http.HandlerFunc(helloNextHandler2)
	if ht22 != nil {
		http.Handle("/nextwispan2", ht22)
	}
	ht3 := http.HandlerFunc(helloNextHandlerWithoutSpan)
	if ht3 != nil {
		http.Handle("/nextwospan", ht3)
	}
	ht33 := http.HandlerFunc(helloNextHandlerWithoutSpan3)
	if ht33 != nil {
		http.Handle("/nextwospan2", ht33)
	}
	err := http.ListenAndServe(":8096", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

