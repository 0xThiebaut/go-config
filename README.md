# Go Config
A simple configuration module without overhead.

**Why?**
Because all other libraries either did not restrict available configuration options or had an unneeded complexity.

```go
package main

import (
	"fmt"
	"github.com/0xThiebaut/go-config"
)

type Config struct {
	My            string
	Exotic        map[string]Config
	Configuration bool
}

func main() {
	demo := &Config{
		My: "Demo",
	}
	c := config.New(&demo)
	if s, err := c.ReadString("my"); err == nil {
		fmt.Println(s)
		// Output: Demo
	}
	if err := c.Write("my", "Hello World!"); err == nil {
		fmt.Println(demo.My)
		// Output: Hello World!
	}
	if err := c.Write("exotic.exotic.exotic.exotic.my", "Success!"); err == nil {
		fmt.Println(demo.Exotic["exotic"].Exotic["exotic"].My)
		// Output: Success!
	}
}
```