package optimize

import "../core"
import "fmt"
import "sync"

type Optimization struct {
	ConversionFunc func(core.Obj, func(core.Obj, core.GlobalState) string, core.GlobalState) string
	Import         string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj, state core.GlobalState) string {
	select {
	case <-state.Stop:
		return fmt.Sprintf("%#v", obj)
	default:
	}
	mutex := &sync.Mutex{}
	newState := state
	newState.Stop = make(chan struct{})
	returnVal := ""
	newState.WaitGroup.Add(len(opts))
	for _, opt := range opts {
		go func() {
			str := opt.ConversionFunc(obj, func(a core.Obj, _ core.GlobalState) string { return OptimizeObjHelper(opts, optsUsed, a, newState) }, newState)
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
	case <-state.Stop:
		defer func() {
			defer func() { recover() }()
			mutex.Lock()
			close(newState.Stop)
			mutex.Unlock()
		}()
		return fmt.Sprintf("%#v", obj)
	case <-newState.Stop:
		mutex.Lock()
		val := returnVal
		mutex.Unlock()
		return val
	}
}

var DefaultOpt = Optimization{
	func(obj core.Obj, simp func(core.Obj, core.GlobalState) string, state core.GlobalState) string {
		// TODO: time it and wait that time before u do the thing
		var val string
		switch v := obj.(type) {
		case core.Function:
			val = fmt.Sprintf("core.Function{%d,%s}", v.N, simp(v.X, state))
		case core.Called:
			val = fmt.Sprintf("core.Called{%s,%s}", simp(v.X, state), simp(v.Y, state))
		case core.CalledChurchNum:
			val = fmt.Sprintf("core.CalledChurchNum{%d,%s}", v.Num, simp(v.X, state))
		default:
			val = fmt.Sprintf("%#v", obj)
		}
		defer func() { recover() }()
		close(state.Stop)
		return val
	},
	"",
}

func OptimizeObj(opts []*Optimization, obj core.Obj) (string, string) {
	optsUsed := make(map[*Optimization]bool)
	state := core.MakeGlobalState()
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
