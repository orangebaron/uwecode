package core

type Controller interface {
	giveInput([]byte) Obj
	takeOutput([]byte, Obj)
	initNewProcess() []byte
}

type IO interface {
	run(Controller, []byte) (Obj, bool)
}

type InputIO struct {
	Obj
}

func (io InputIO) run(controller Controller, pid []byte) (Obj, bool) {
	return io.Obj.Call(controller.giveInput(pid)), true
}

type OutputIO struct {
	otp     Obj
	nextObj Obj
}

func (io OutputIO) run(controller Controller, pid []byte) (Obj, bool) {
	controller.takeOutput(pid, io.otp)
	return io.nextObj, true
}

type ForkIO struct {
	objA Obj
	objB Obj
}

func (io ForkIO) run(controller Controller, _ []byte) (Obj, bool) {
	go runProcess(io.objB, controller)
	return io.objA, true
}

type NullIO struct{}

func (_ NullIO) run(_ Controller, _ []byte) (Obj, bool) {
	return ReturnVal{0}, false
}

func runProcess(f Obj, controller Controller) {
	pid := controller.initNewProcess()
	for keepGoing := true; keepGoing; f, keepGoing = ObjToIO(f).run(controller, pid) {
	}
}
