package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnpathsNonExistingCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := newProgram([]string{"unpath", "not-cat", "cat", "main_test.go"}, &stdout, &stderr).main()
	if status != 0 {
		t.Fatal(stderr)
	}
}

func TestUnpathsCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := newProgram([]string{"unpath", "cat", "cat", "main_test.go"}, &stdout, &stderr).main()
	if status == 0 {
		t.Errorf("got: %d; want: %d", status, 0)
	}
	message := "env: cat: No such file or directory"
	if !strings.Contains(stderr.String(), message) {
		t.Errorf("got: %s; want: %s", stderr.String(), message)
	}
}

func TestUnpathsNonExistingCommandThroughScript(t *testing.T) {
	script := createScript("#!/usr/bin/env bash\ncat $1", t.Fatal)
	var stdout, stderr bytes.Buffer
	status := newProgram([]string{"unpath", "not-cat", script.Name(), "main_test.go"}, &stdout, &stderr).main()
	if status != 0 {
		t.Fatal(stderr.String())
	}
}

func TestUnpathsCommandThroughScript(t *testing.T) {
	script := createScript("#!/usr/bin/env bash\ncat $1", t.Fatal)
	var stdout, stderr bytes.Buffer
	status := newProgram([]string{"unpath", "cat", script.Name(), "main_test.go"}, &stdout, &stderr).main()
	if status == 0 {
		t.Fatal(stderr)
	}
	message := "cat: command not found"
	if !strings.Contains(stderr.String(), message) {
		t.Errorf("got: %s; want: %s", stderr.String(), message)
	}
}

func Test_e2e_UnpathsSiblingCommand(t *testing.T) {
	dir := createDir(t.Fatal)
	command := createScriptIn(dir, "#!/usr/bin/env bash\ncat $1", t.Fatal)
	command_ := filepath.Base(command.Name())
	sibling := createScriptIn(dir, "#!/usr/bin/env bash\ncat $1", t.Fatal)
	sibling_ := filepath.Base(sibling.Name())
	script := createScriptIn(dir, fmt.Sprintf("#!/usr/bin/env bash\n%s $1", command_), t.Fatal)
	script_ := filepath.Base(script.Name())

	path, _ := os.LookupEnv("PATH")
	path = fmt.Sprintf("%s:%s", dir, path)

	arg := []string{"go", "run", "main.go", sibling_, script_, "main_test.go"}
	arg = append([]string{"-P", path}, arg...)
	cmd := exec.Command("env", arg...)
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PATH=%s", path))
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_e2e_UnpathsCommand(t *testing.T) {
	dir := createDir(t.Fatal)
	command := createScriptIn(dir, "#!/usr/bin/env bash\ncat $1", t.Fatal)
	command_ := filepath.Base(command.Name())
	script := createScriptIn(dir, fmt.Sprintf("#!/usr/bin/env bash\n%s $1", command_), t.Fatal)
	script_ := filepath.Base(script.Name())

	path, _ := os.LookupEnv("PATH")
	path = fmt.Sprintf("%s:%s", dir, path)

	arg := []string{"go", "run", "main.go", command_, script_, "main_test.go"}
	arg = append([]string{"-P", path}, arg...)
	cmd := exec.Command("env", arg...)
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PATH=%s", path))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal(err)
	}
	message := fmt.Sprintf("%s: command not found", command_)
	if !strings.Contains(stderr.String(), message) {
		t.Errorf("got: %s; want: %s", stderr.String(), message)
	}
}

func Test_e2e_UnpathsCommandsRecursively(t *testing.T) {
	script := createScript("#!/usr/bin/env bash\ncat $1", t.Fatal)
	cmd := exec.Command("go", "run", "main.go", "non-cat", "go", "run", "main.go", "cat", script.Name(), "main_test.go")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal(err)
	}
	message := fmt.Sprintf("cat: command not found")
	if !strings.Contains(stderr.String(), message) {
		t.Errorf("got: %s; want: %s", stderr.String(), message)
	}
}

func createScript(content string, fatal func(args ...any)) *os.File {
	dir := createDir(fatal)
	return createScriptIn(dir, content, fatal)
}

func createDir(fatal func(args ...any)) string {
	dir, err := os.MkdirTemp("", "bin")
	if err != nil {
		fatal(err)
	}
	return dir
}

func createScriptIn(dir, content string, fatal func(args ...any)) *os.File {
	file, err := os.CreateTemp(dir, "script")
	if err != nil {
		fatal(err)
	}
	err = os.Chmod(file.Name(), 0777)
	if err != nil {
		fatal(err)
	}
	fmt.Fprintf(file, content)

	return file
}
