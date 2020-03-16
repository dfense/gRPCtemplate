package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Pallinder/go-randomdata"
)

type Person struct {
	fname  string
	lname  string
	age    int
	ctx    context.Context
	cancel context.CancelFunc
}

// main entry point
func main() {

	fmt.Println("hi")
	log.Println("hi")
	con, cancel := context.WithCancel(context.Background())
	p := &Person{fname: "jfive", age: 60}
	p.ctx = con
	p.cancel = cancel

	for i := 0; i < 2; i++ {
		p.changeName()
	}
	cancel()

}

func (p *Person) changeName() error {
	p.lname = randomdata.SillyName()
	fmt.Println(p)
	p.longOperationOnPerson()

	return nil
}

func (p *Person) longOperationOnPerson() {

	fmt.Println("nothing here")

	for {

		time.Sleep(time.Second * 3)
		fmt.Println("unsleep %s\n", p.lname)
		select {
		case <-p.ctx.Done():

		default:

		}

	}

	fmt.Println("Exiting %s\n", p.lname)

}
