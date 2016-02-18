package main

import "fmt"

// START OMIT
type FooPrinter interface {
	PrintFoo()
}

type Metadata struct {
	Foo string
}

func (m Metadata) PrintFoo() {
	fmt.Println(m.Foo)
}

type MyAccount struct {
	Metadata
}

func main() {
	MyAccount{Metadata: Metadata{"Food bar"}}.PrintFoo()
}
// END OMIT

