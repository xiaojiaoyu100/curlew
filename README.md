# Curlew

Curlew is a job queue based on single machine memory.

## Feature

* Automatically scale the number of workers up or down.
* Don't ensure the job execution order.

## Usage

```go
d, err := curlew.New()
if err != nil {
    return
}
j := curlew.NewJob()
j.Arg = 3
j.Fn = func(ctx context.Context, arg interface{}) error {
    fmt.Println("I'm done")
    return nil
}
d.SubmitAsync(j)
```
