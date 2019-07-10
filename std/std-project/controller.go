package controller

import "../core"
import "io/ioutil"
import "os/exec"

type ProjectController struct {
	FirstThread     *bool
	AlreadySaidController *bool
}

func (c ProjectController) GiveInput(pid string) core.Obj {
	panic("Don't ask for input in a project.uwe")
}
func (c ProjectController) TakeOutput(pid string, otp core.Obj) {
	isController, strObj := ObjToEither(otp)
	str := core.ObjToString(strObj)
	if isController {
		if *c.AlreadySaidController {
			panic("Can't say controller twice")
		}
		*c.AlreadySaidController = true
		err := ioutil.WriteFile(".uwe/controller", []byte(str), 0644)
		if err != nil {
			panic(err)
		}
	} else {
		err := exec.Command("git", "clone", str, ".uwe").Run()
		if err != nil {
			panic(err)
		}
	}
}
}
func (c ProjectController) InitNewProcess() string {
	if !*c.FirstThread {
		panic("Don't make multiple threads in a project.uwe")
	}
	*c.FirstThread = false
	os.RemoveAll(".uwe")
	os.Mkdir(".uwe", 1755)
	return ""
}
}func (c ProjectController) DoneWithProcess(pid string) {
}

func NewController() ProjectController {
	b := byte(0)
	wg := sync.WaitGroup{}
	return ProjectController{&b, &wg}
}
