package main

import "../reader"
import "../core"
import "os"
import "io/ioutil"
import "fmt"
import "os/exec"
import "errors"
import "strings"

func readFile(filename string) (reader.DeclaredDict, error) {
	dict := reader.NewDeclaredDict()
	f, err := os.Open(filename + ".uwe")
	if err != nil {
		return dict, err
	}
	err = reader.ReadCode(f, dict)
	if err != nil {
		return dict, err
	}
	return dict, nil
}

func makeGoFile(filename string, importStrings []string) error {
	dict, err := readFile(filename)
	if err != nil {
		return err
	}
	mainObj := dict.GetObj("main") // TODO: this can panic
	head, optimizedObj, err := optimizeObj(mainObj, filename)
	if err != nil {
		return err
	}
	goifiedImportString := "" //TODO: can create duplicates with optimizer
	for _, v := range importStrings {
		goifiedImportString += "import\"" + v + "\"\n"
	}
	projNameSplit := strings.Split(importStrings[0], "/")
	str := fmt.Sprintf("package main\n\nimport \"./core\"\n%s\nvar mainObj core.Obj = %s\n\n%s\n\nfunc main() {\n\tcore.RunProcess(mainObj, %s.NewController(), make(chan struct{}))\n}\n", goifiedImportString, head, optimizedObj, projNameSplit[len(projNameSplit)-1]) // TODO: ./core -> github link
	err = ioutil.WriteFile(filename+".go", []byte(str), 0644)
	if err != nil {
		return err
	}
	return nil
}

func runCmdNicely(cmd *exec.Cmd) (string, error) {
	s, err := cmd.Output()
	if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
		return string(s), errors.New(string(exitErr.Stderr) + "\n" + err.Error())
	}
	return string(s), err
}

func makeOptimizeGoOpts() string {
	defaultString := "var opts = []*optimize.Optimization{}"
	files, err := ioutil.ReadDir("./.uwe/opts")
	if err != nil {
		return defaultString
	}
	importsStr := ""
	optsStr := ""
	for _, f := range files {
		name := f.Name()
		importsStr += fmt.Sprintf("import \"./.uwe/opts/%s\"\n", name)
		if optsStr != "" {
			optsStr += " + "
		}
		optsStr += name + ".OptsList"
	}
	if optsStr == "" {
		return defaultString
	} else {
		return importsStr + "\nvar opts = " + optsStr
	}
}

func makeGoOptimizeFile(obj core.Obj, filename string) error {
	optsStr := makeOptimizeGoOpts()
	str := fmt.Sprintf("package main\n\nimport \"os\"\nimport \"fmt\"\nimport \"./core\"\nimport \"./optimize\"\n%s\nvar mainObj core.Obj = %#v\n\nfunc main() {\n\ta, b := optimize.OptimizeObj(opts, mainObj)\n\tfmt.Println(a)\n\tfmt.Fprintln(os.Stderr, b)\n}\n", optsStr, obj)
	return ioutil.WriteFile(filename+".go", []byte(str), 0644)
}

func optimizeObj(obj core.Obj, filename string) (string, string, error) {
	err := makeGoOptimizeFile(obj, filename)
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command("go", "run", filename+".go")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	err = cmd.Start()
	if err != nil {
		return "", "", err
	}
	head, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", "", err
	}
	optimizedObj, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", "", err
	}
	return string(head), string(optimizedObj), err
}

func buildFile(filename string, importStrings []string) error {
	err := makeGoFile(filename+".go", importStrings)
	if err != nil {
		return err
	}
	_, err = runCmdNicely(exec.Command("go", "build", filename+".go"))
	if err != nil {
		return err
	}
	os.Remove(filename + ".go")
	return nil
}

func runFile(filename string, importStrings []string) error {
	err := makeGoFile(filename, importStrings)
	if err != nil {
		return err
	}
	_, err = runCmdNicely(exec.Command("go", "build", filename+".go"))
	if err != nil {
		return err
	}
	cmd := exec.Cmd{Path: "./" + filename, Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
	err = cmd.Run()
	os.Remove(filename)
	if err != nil {
		return err
	}
	os.Remove(filename + ".go")
	return nil
}

func command_build() error {
	controller, err := ioutil.ReadFile(".uwe/controller")
	if err != nil {
		return err
	}
	return buildFile(os.Args[2], []string{string(controller)})
}

func command_run() error {
	controller, err := ioutil.ReadFile(".uwe/controller")
	if err != nil {
		return err
	}
	return runFile(os.Args[2], []string{string(controller)})
}

func command_project() error {
	return runFile("project", []string{"./std/std_project"})
}

func main() {
	var err error
	switch os.Args[1] {
	case "build":
		err = command_build()
	case "run":
		err = command_run()
	case "project":
		err = command_project()
	default:
		err = errors.New("Unrecognized command")
	}
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
	}
}
