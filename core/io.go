package core

type Controller interface {
	giveInput([]byte) Obj
	takeOutput([]byte, Obj)
	initNewProcess() []byte
}

type IO interface {
	run(Controller, []byte) Obj
}

type InputIO struct {
	Obj
}

func (io InputIO) run(controller Controller, pid []byte) Obj {
	return io.Obj.Call(controller.giveInput(pid))
}

type OutputIO struct {
	otp     Obj
	nextObj Obj
}

func (io OutputIO) run(controller Controller, pid []byte) Obj {
	controller.takeOutput(pid, io.otp)
	return io.nextObj
}

type ForkIO struct {
	objA Obj
	objB Obj
}

func (io ForkIO) run(controller Controller, _ []byte) Obj {
	go runProcess(io.objB, controller)
	return io.objA
}

func runProcess(f Obj, controller Controller) {
	pid := controller.initNewProcess()
	for {
		f = ObjToIO(f).run(controller, pid)
	}
}
