package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
)

type Handler struct {
	// preHook []reflect.Value // 前置钩子
	f      reflect.Value
	input  reflect.Type
	output reflect.Type
}

var registry = map[string]Handler{}

func init() {
	registry["/echo"] = Handler{
		f:      reflect.ValueOf(echo),
		input:  reflect.TypeOf(EchoReq{}),
		output: reflect.TypeOf(EchoResp{}),
	}
}

type EchoReq struct {
	Msg string
}

type EchoResp struct {
	Msg string
}

func echo(req *EchoReq, rv *EchoResp) {
	rv.Msg = req.Msg

	log.Printf("recv:%+v", req)
}

func Send(w io.Writer, buf []byte) {
	log.Printf("send: %+v", string(buf))
	w.Write(buf)
}

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id := r.URL.Path
	h, ok := registry[id]
	if !ok {
		Send(w, []byte(fmt.Sprintf("unsupported url:%v", id)))
		return
	}

	var buf []byte
	var err error

	switch r.Method {
	case "GET":
		buf = []byte(r.FormValue("json"))

	case "POST":
		buf, err = ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			Send(w, []byte(err.Error()))
			return
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		Send(w, []byte("unsupported method"))
		return
	}

	// v for first argument, v2 as output for second argument
	v, v2 := reflect.New(h.input), reflect.New(h.output)

	// decoding input
	if err := json.Unmarshal(buf, v.Interface()); err != nil {
		Send(w, []byte(err.Error()))
		return
	}

	h.f.Call([]reflect.Value{v, v2})
	rv, _ := json.Marshal(v2.Interface())
	Send(w, rv)

}

func main() {
	log.Println("start server...")
	http.ListenAndServe(":8080", &Server{})
}
