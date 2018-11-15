package main

import (
	"fmt"
	"net/http"

	"github.com/Dmaxzj/gb2312"
	"github.com/urfave/negroni"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%s\n", req.Form)
		fmt.Printf("%s\n", req.PostForm)
		fmt.Fprintf(w, "测试")
	})
	n := negroni.Classic()
	n.Use(&gb2312.Gb2312Encode{})
	n.UseHandler(mux)
	n.Run(":3000")
	// reader := transform.NewReader(bytes.NewReader([]byte("你好")), simplifiedchinese.HZGB2312.NewEncoder())
	// b, _ := ioutil.ReadAll(reader)
	// fmt.Println(string(b))
}
