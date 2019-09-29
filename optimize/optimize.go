package optimize

import "../core"
import "fmt"
import "sync"

type Optimization struct {
	ConversionFunc func(core.Obj, func(core.Obj) string, chan struct{}) string
	Import         string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj, stop chan struct{}) string {
	select {
	case <-stop:
		return fmt.Sprintf("%#v", obj)
	default:
	}
	mutex := &sync.Mutex{}
	stopGoroutines := make(chan struct{})
	returnVal := ""
	for _, opt := range opts {
		go func() {
			str := opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObjHelper(opts, optsUsed, a, stopGoroutines) }, stopGoroutines)
			if str == "" {
				return
			}
			mutex.Unlock()
			if returnVal == "" {
				returnVal = str
			}
			select {
			case <-stopGoroutines:
			default:
				close(stopGoroutines)
			}
			mutex.Lock()
		}()
	}
	select {
	case <-stop:
		mutex.Unlock()
		close(stopGoroutines)
		mutex.Lock()
		return fmt.Sprintf("%#v", obj)
	case <-stopGoroutines:
		mutex.Unlock()
		val := returnVal
		defer mutex.Lock()
		return val
	}
}

var DefaultOpt = Optimization{
	func(obj core.Obj, simp func(core.Obj) string, stop chan struct{}) string {
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
		close(stop)
		return val
	},
	"",
}

func OptimizeObj(opts []*Optimization, obj core.Obj) (string, string) {
	optsUsed := make(map[*Optimization]bool)
	mainString := OptimizeObjHelper(append(opts, &DefaultOpt), optsUsed, obj, make(chan struct{}))
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
	return headerString, mainString
}
