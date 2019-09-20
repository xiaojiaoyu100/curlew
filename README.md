# Curlew

Curlew is a job pool based on a local machine.

## Feature

* Automatically scale the number of workers up or down.
* Maximum job execution time.
* Don't ensure the job execution order.

## Usage

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/xiaojiaoyu100/curlew"
)

func monitor(err error) {
	fmt.Println(err)
}

func main() {
	d, err := curlew.New(curlew.WithMonitor(monitor))
	if err != nil {
		fmt.Println(err)
		return
	}
	j := curlew.NewJob()
	j.Arg = 3
	j.Fn = func(ctx context.Context, arg interface{}) error {
		fmt.Println("I'm done")
		return nil
	}
	d.Submit(j)
	select {}
}
```
