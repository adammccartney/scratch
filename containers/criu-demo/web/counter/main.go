package main

import (
	"fmt"
	"log"
	"net/http"
)

type Counter struct {
	val int
}

func (c *Counter) NewCounter() *Counter {
	return &Counter{
		val: 0,
	}
}

func (c *Counter) Increment() {
	c.val += 1
}

func countHandler(c *Counter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Increment()
		fmt.Fprintf(w, "Current count: %d\n", c.val)
	}
}

func main() {
	var c *Counter
	c = c.NewCounter()

	http.HandleFunc("/", countHandler(c))
	log.Fatal(http.ListenAndServe(":9993", nil))
}
