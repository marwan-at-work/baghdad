# Dockerfile Parser

If you simply want to input a docker file and read its commands.

`go get github.com/marwan-at-work/dfparser`

### Usage

```go
import (
  "github.com/marwan-at-work/dfparser"
  "fmt"
)

func main() {
  f, err := os.Open("./Dockerfile")
  if err != nil {
    panic(err)
  }
  defer f.Close()

  d, err := dfparser.Parse(f) // handleErr
  if err != nil {
    panic(err)
  }

  fmt.Println("from:", d.From, "workdir:", d.Workdir)
}
```
