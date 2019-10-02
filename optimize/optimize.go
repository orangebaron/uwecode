package optimize

import "../core"
import "fmt"
import "sync"

type Optimization struct {
	ConversionFunc func(core.Obj, func(core.Obj) string, core.GlobalState) string
	Import         string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj, globalState core.GlobalState) string {
	select {
	case <-globalState.Stop:
		return fmt.Sprintf("%#v", obj)
	default:
	}
	mutex := &sync.Mutex{}
	newState := globalState
	newState.Stop = make(chan struct{})
	returnVal := ""
	newState.WaitGroup.Add(len(opts))
	for _, opt := range opts {
		go func() {
			str := opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObjHelper(opts, optsUsed, a, newState) }, newState)
			if str == "" {
				newState.WaitGroup.Done()
				return
			}
			mutex.Lock()
			if returnVal == "" {
				returnVal = str
			}
			select {
			case <-newState.Stop:
			default:
				close(newState.Stop)
			}
			mutex.Unlock()
			newState.WaitGroup.Done()
		}()
	}
	select {
	case <-globalState.Stop:
		mutex.Lock()
		close(newState.Stop)
		mutex.Unlock()
		return fmt.Sprintf("%#v", obj)
	case <-newState.Stop:
		mutex.Lock()
		val := returnVal
		defer mutex.Unlock()
		return val
	}
}

var DefaultOpt = Optimization{
	func(obj core.Obj, simp func(core.Obj) string, globalState core.GlobalState) string {
		// TODO: time it and wait that time before u do the thing
		var val string
		switch v := obj.(type) {
		case core.Function:
			val = fmt.Sprintf("core.Function{%d,%s}", v.N, simp(v.X))
		case core.Called:
			val = fmt.Sprintf("core.Called{%s,%s}", simp(v.X), simp(v.Y))
		case core.CalledChurchNum:
			val = fmt.Sprintf("core.CalledChurchNum{%d,%s}", v.Num, simp(v.X))
		default:
			val = fmt.Sprintf("%#v", obj)
		}
		defer func() { recover() }()
		close(globalState.Stop)
		return val
	},
	"",
}

func OptimizeObj(opts []*Optimization, obj core.Obj) (string, string) {
	optsUsed := make(map[*Optimization]bool)
	state := core.GlobalState{&sync.WaitGroup{}, make(chan struct{})}
	mainString := OptimizeObjHelper(append(opts, &DefaultOpt), optsUsed, obj, state)
	headerString := ""
	importsUsed := make(map[string]bool)
	for opt, wasUsed := range optsUsed {
		if wasUsed {
			importsUsed[opt.Import] = true
		}
	}
	for imp, _ := range importsUsed {
		if imp != "" {
			headerString += fmt.Sprintf("import \"%s\"\n", imp)
		}
	}
	state.WaitGroup.Wait()
	return headerString, mainString
}
