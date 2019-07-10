package std_project

import "../../core"
import "io/ioutil"
import "os/exec"
import "os"
import "strings"
import "errors"

func runCmdNicely(cmd *exec.Cmd) error {
	_, err := cmd.Output()
	if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
		return errors.New(string(exitErr.Stderr) + "\n" + err.Error())
	} else {
		return err
	}
}

type ProjectController struct {
	FirstThread     *bool
	AlreadySaidController *bool
}

func (c ProjectController) GiveInput(pid string) core.Obj {
	panic("Don't ask for input in a project.uwe")
}
func (c ProjectController) TakeOutput(pid string, otp core.Obj) {
	isController, strObj := core.ObjToEither(otp)
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
		split := strings.Split(str, "/")
		afterLastSlash := split[len(split) - 1]
		beforePeriod := afterLastSlash[:len(afterLastSlash) - 4]
		err := runCmdNicely(exec.Command("git", "clone", str, ".uwe/" + beforePeriod))
		if err != nil {
			panic(err)
		}
	}
}
func (c ProjectController) InitNewProcess() string {
	if !*c.FirstThread {
		panic("Don't make multiple threads in a project.uwe")
	}
	*c.FirstThread = false
	os.Remove(".uwe")
	os.Mkdir(".uwe", os.ModePerm)
	return ""
}
func (c ProjectController) DoneWithProcess(pid string) {
}

func NewController() ProjectController {
	b1, b2 := true, false
	return ProjectController{&b1, &b2}
}
