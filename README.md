# gomonit

Package gomonit consumes and parses Monit status and event notifications. It disguises as M/Monit collector server.

## Example

```go
// create channel and pass it to the collector
channel := make(chan *gomonit.Monit)
collector := gomonit.NewCollector(channel)
http.Handle("/collector", collector)

go func() {
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("http.ListenAndServe: ", err)
	}
}()

// consume notifications
for monit := range channel {
	fmt.Println(monit.Server.Uptime)
}
```
# License

ISC
