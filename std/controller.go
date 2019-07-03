package std

import "../core"
import "fmt"
import "sync"

type ExampleController struct {
	n  *byte
	wg *sync.WaitGroup
}

func (c ExampleController) GiveInput(pid string) core.Obj {
	return core.ChurchNum{uint(pid[0])}
}
func (c ExampleController) TakeOutput(pid string, otp core.Obj) {
	fmt.Println(uint(pid[0]))
	fmt.Println(otp)
}
func (c ExampleController) InitNewProcess() string {
	*c.n++
	c.wg.Add(1)
	return string([]byte{*c.n})
}
func (c ExampleController) DoneWithProcess(pid string) {
	c.wg.Done()
	if pid[0] == 1 {
		c.wg.Wait()
	}
}

func NewController() ExampleController {
	b := byte(0)
	wg := sync.WaitGroup{}
	return ExampleController{&b, &wg}
}
