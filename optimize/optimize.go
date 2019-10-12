package optimize

import "../core"
import "fmt"
import "sync"

type Optimization struct {
	ConversionFunc func(core.Obj, func(core.Obj) string, core.SimplifyState) string
	Import         string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj, state core.SimplifyState) string {
	select {
	case <-state.GlobalState.Stop:
		return fmt.Sprintf("%#v", obj)
	default:
	}
	mutex := &sync.Mutex{}
	newState := state
	newState.GlobalState.Stop = make(chan struct{})
	returnVal := ""
	newState.GlobalState.WaitGroup.Add(len(opts))
	for _, opt := range opts {
		go func() {
			str := opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObjHelper(opts, optsUsed, a, newState) }, newState)
			if str == "" {
				newState.GlobalState.WaitGroup.Done()
				return
			}
			mutex.Lock()
			if returnVal == "" {
				returnVal = str
			}
			select {
			case <-newState.GlobalState.Stop:
			default:
				close(newState.GlobalState.Stop)
			}
			mutex.Unlock()
			newState.GlobalState.WaitGroup.Done()
		}()
	}
	select {
	case <-state.GlobalState.Stop:
		defer func() {
			defer func() { recover() }()
			mutex.Lock()
			close(newState.GlobalState.Stop)
			mutex.Unlock()
		}()
		return fmt.Sprintf("%#v", obj)
	case <-newState.GlobalState.Stop:
		mutex.Lock()
		val := returnVal
		mutex.Unlock()
		return val
	}
}

var DefaultOpt = Optimization{
	func(obj core.Obj, simp func(core.Obj) string, state core.SimplifyState) string {
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
		close(state.GlobalState.Stop)
		return val
	},
	"",
}

func OptimizeObj(opts []*Optimization, obj core.Obj) (string, string) {
	optsUsed := make(map[*Optimization]bool)
	state := core.MakeSimplifyState()
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
	state.GlobalState.WaitGroup.Wait()
	return headerString, mainString
}
