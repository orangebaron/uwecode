package core

type Controller interface {
	GiveInput(string, SimplifyState) Obj
	TakeOutput(string, Obj, SimplifyState)
	InitNewProcess() string
	DoneWithProcess(string)
}

type IO interface {
	Run(Controller, string, SimplifyState) (Obj, bool)
}

type InputIO struct {
	Obj
}

func (io InputIO) Run(controller Controller, pid string, state SimplifyState) (Obj, bool) {
	return io.Obj.Call(controller.GiveInput(pid, state)), true
}

type OutputIO struct {
	Otp     Obj
	NextObj Obj
}

func (io OutputIO) Run(controller Controller, pid string, state SimplifyState) (Obj, bool) {
	controller.TakeOutput(pid, io.Otp, state)
	return io.NextObj, true
}

type ForkIO struct {
	objA Obj
	objB Obj
}

func (io ForkIO) Run(controller Controller, _ string, state SimplifyState) (Obj, bool) {
	c := make(chan struct{})
	go RunProcess(io.objB, controller, c, state)
	<-c
	return io.objA, true
}

type NullIO struct{}

func (_ NullIO) Run(_ Controller, _ string, _ SimplifyState) (Obj, bool) {
	return nil, false
}

func RunProcess(f Obj, controller Controller, closeWhenInitted chan struct{}, state SimplifyState) {
	pid := controller.InitNewProcess()
	close(closeWhenInitted)
	for keepGoing := true; keepGoing; f, keepGoing = ObjToIO(f, state).Run(controller, pid, state) {
	}
	controller.DoneWithProcess(pid)
}
