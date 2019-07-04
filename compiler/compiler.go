package main

//import "../core"
import "../reader"
import "os"
import "io/ioutil"
import "fmt"
import "os/exec"

func main() {
	f, err := os.Open(os.Args[1] + ".uwe")
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}
	dict := reader.NewDeclaredDict()
	err = reader.ReadCode(f, dict)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}
	mainObj := dict.GetObj("main")
	str := "package main\n\nimport \"./core\"\nimport\"" + os.Args[2] + "\"\n\nvar mainObj core.Obj = " + fmt.Sprintf("%#v", mainObj) + "\n\nfunc main() {\n\tcore.RunProcess(mainObj, " + os.Args[3] + ", make(chan struct{}))\n}"
	ioutil.WriteFile(os.Args[1]+".go", []byte(str), 0644)
	exec.Command("go", "build", os.Args[1]+".go").Run()
	os.Remove(os.Args[1] + ".go")
}
