package core

import "fmt"

type Controller interface {
	GiveInput(string) Obj
	TakeOutput(string, Obj)
	InitNewProcess() string
	DoneWithProcess(string)
}

type IO interface {
	run(Controller, string) (Obj, bool)
}

type InputIO struct {
	Obj
}

func (io InputIO) run(controller Controller, pid string) (Obj, bool) {
	return io.Obj.Call(controller.GiveInput(pid)), true
}

type OutputIO struct {
	otp     Obj
	nextObj Obj
}

func (io OutputIO) run(controller Controller, pid string) (Obj, bool) {
	controller.TakeOutput(pid, io.otp)
	return io.nextObj, true
}

type ForkIO struct {
	objA Obj
	objB Obj
}

func (io ForkIO) run(controller Controller, _ string) (Obj, bool) {
	c := make(chan struct{})
	go RunProcess(io.objB, controller, c)
	<-c
	return io.objA, true
}

type NullIO struct{}

func (_ NullIO) run(_ Controller, _ string) (Obj, bool) {
	return ReturnVal{0}, false
}

func RunProcess(f Obj, controller Controller, closeWhenInitted chan struct{}) {
	fmt.Println(f)
	pid := controller.InitNewProcess()
	close(closeWhenInitted)
	for keepGoing := true; keepGoing; f, keepGoing = ObjToIO(f).run(controller, pid) {
	}
	controller.DoneWithProcess(pid)
}
