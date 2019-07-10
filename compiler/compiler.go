package main

import "../reader"
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
	goifiedImportString := ""
	for _, v := range importStrings {
		goifiedImportString += "import\"" + v + "\"\n"
	}
	projNameSplit := strings.Split(importStrings[0], "/")
	str := fmt.Sprintf("package main\n\nimport \"./core\"\n%s\nvar mainObj core.Obj = %#v\n\nfunc main() {\n\tcore.RunProcess(mainObj, %s.NewController(), make(chan struct{}))\n}", goifiedImportString, mainObj, projNameSplit[len(projNameSplit)-1]) // TODO: ./core -> github link
	err = ioutil.WriteFile(filename+".go", []byte(str), 0644)
	if err != nil {
		return err
	}
	return nil
}

func runCmdNicely(cmd *exec.Cmd) error {
	_, err := cmd.Output()
	if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
		return errors.New(string(exitErr.Stderr) + "\n" + err.Error())
	} else {
		return err
	}
}

func buildFile(filename string, importStrings []string) error {
	err := makeGoFile(filename, importStrings)
	if err != nil {
		return err
	}
	err = runCmdNicely(exec.Command("go", "build", filename+".go"))
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
	err = runCmdNicely(exec.Command("go", "run", filename+".go"))
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
