# gb2312

## 使用

```go
package main

import (
  "net/http"
  "github.com/Dmaxzj/gb2312"
  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "gb2312测试")
  })

  n := negroni.New()
  n.Use(negroni.NewGB2312())
  n.UseHandler(mux)

  http.ListenAndServe(":3003", n)
}
```