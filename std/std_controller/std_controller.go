package std_controller

import "../../core"
import "fmt"
import "sync"
import "bufio"
import "os"
import "os/exec"

type CommandType uint

const (
	GetPid CommandType = iota
	WriteStdout
	WriteStderr
	ReadStdin
	RunCommand
	ReadCommandStdout
	ReadCommandStderr
	WriteCommandStdin
	WaitForCommand
)

type StdController struct {
	PidNum     *byte
	CommandNum *uint
	WG         *sync.WaitGroup
	NextInput  map[string]core.Obj
	Commands   map[uint]*exec.Cmd
}

func (c StdController) GiveInput(pid string) core.Obj {
	return c.NextInput[pid]
}
func (c StdController) TakeOutput(pid string, otp core.Obj) {
	otpType, otpVal := core.ObjToTuple(otp)
	switch CommandType(core.ObjToInt(otpType)) {
	case GetPid:
		c.NextInput[pid] = core.ChurchTupleCharString{pid}
	case WriteStdout:
		fmt.Print(core.ObjToString(otpVal))
	case WriteStderr:
		os.Stderr.WriteString(core.ObjToString(otpVal))
	case ReadStdin:
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n') // TODO: what about just reading a single character
		if err != nil {
			panic(err)
		}
		c.NextInput[pid] = core.ChurchTupleCharString{text}
	case RunCommand:
		cmd := exec.Command(core.ObjToString(otpVal))
		*c.CommandNum++
		c.Commands[*c.CommandNum] = cmd
		cmd.Start()
		c.NextInput[pid] = core.ChurchNum{*c.CommandNum}
	case ReadCommandStdout:
		cmdNum, size := core.ObjToTuple(otpVal)
		str := make([]byte, core.ObjToInt(size))
		pipe, _ := c.Commands[core.ObjToInt(cmdNum)].StdoutPipe()
		pipe.Read(str)
		c.NextInput[pid] = core.ChurchTupleCharString{string(str)}
	case ReadCommandStderr:
		cmdNum, size := core.ObjToTuple(otpVal)
		str := make([]byte, core.ObjToInt(size))
		pipe, _ := c.Commands[core.ObjToInt(cmdNum)].StderrPipe()
		pipe.Read(str)
		c.NextInput[pid] = core.ChurchTupleCharString{string(str)}
	case WriteCommandStdin:
		cmdNum, str := core.ObjToTuple(otpVal)
		pipe, _ := c.Commands[core.ObjToInt(cmdNum)].StdinPipe()
		pipe.Write([]byte(core.ObjToString(str)))
	case WaitForCommand:
		c.Commands[core.ObjToInt(otpVal)].Wait()
	default:
		panic("Invalid command number")
	} // TODO: communication between threads
	// TODO: get os args
}
func (c StdController) InitNewProcess() string {
	*c.PidNum++
	c.WG.Add(1)
	return string([]byte{*c.PidNum})
}
func (c StdController) DoneWithProcess(pid string) {
	c.WG.Done()
	if pid[0] == 1 {
		c.WG.Wait()
	}
}

func NewController() StdController {
	b := byte(0)
	i := uint(0)
	wg := sync.WaitGroup{}
	return StdController{&b, &i, &wg, make(map[string]core.Obj), make(map[uint]*exec.Cmd)}
}
