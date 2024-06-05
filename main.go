package main

import (
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type program struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

func newProgram(args []string, stdout, stderr io.Writer) *program {
	return &program{args: args, stdout: stdout, stderr: stderr}
}

func main() {
	program := newProgram(os.Args, os.Stdout, os.Stderr)
	status := program.main()
	os.Exit(status)
}

func (p *program) main() int {
	uncmd, cmd, status := p.parse()
	if status > 0 {
		return status
	}
	path, status := p.unpath(uncmd)
	if status > 0 {
		return status
	}
	return p.run(cmd, path)
}

func (p *program) parse() (string, []string, int) {
	if len(p.args) < 3 {
		err := `Usage: {PROGRAM} UNCMD CMD

unpath runs CMD with a modified PATH that does not contain UNCMD.

Arguments:
  UNCMD the command to hide from PATH
  CMD   the command to run with the modified PATH

Examples:
  unpath cat ./script script-arg

  unpath cat CMD subcmd-arg

  unpath cat unpath env CMD`
		err = strings.ReplaceAll(err, "{PROGRAM}", p.args[0])
		fmt.Fprintf(p.stderr, fmt.Sprintln(err))
		return "", nil, 1
	}
	return p.args[1], p.args[2:], 0
}

type result struct {
	dir    string
	status int
}

func (p *program) unpath(cmd string) (string, int) {
	path, _ := os.LookupEnv("PATH")
	dirs := strings.Split(path, ":")
	newDirs := make([]result, len(dirs))
	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		go func(i int, dir string) {
			entries, _ := os.ReadDir(dir) // ignore errors caused by empty dirs in PATH
			n, found := slices.BinarySearchFunc(entries, cmd, func(a fs.DirEntry, b string) int {
				return cmp.Compare(a.Name(), b)
			})
			if found {
				dir, status := p.unpathEntry(dir, entries, n)
				newDirs[i] = result{dir, status}
			} else {
				newDirs[i] = result{dir, 0}
			}
			wg.Done()
		}(i, dir)
	}
	wg.Wait()
	for i, result := range newDirs {
		if result.status > 0 {
			return "", result.status
		}
		dirs[i] = result.dir
	}
	return strings.Join(dirs, ":"), 0
}

func (p *program) unpathEntry(dir string, entries []fs.DirEntry, entriesIndex int) (string, int) {
	tmpdir, err := os.MkdirTemp("", filepath.Base(dir))
	if err != nil {
		fmt.Fprintln(p.stderr, err)
		return "", 1
	}

	for i, entry := range entries {
		if i == entriesIndex {
			continue
		}
		err := os.Symlink(filepath.Join(dir, entry.Name()), filepath.Join(tmpdir, entry.Name()))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return "", 1
		}
	}

	return tmpdir, 0
}

func (p *program) run(cmd []string, path string) int {
	arg := append([]string{"-P", path}, cmd...)
	subcmd := exec.Command("env", arg...)
	subcmd.Env = append(subcmd.Environ(), fmt.Sprintf("PATH=%s", path))
	subcmd.Stdout = p.stdout
	subcmd.Stderr = p.stderr
	if subcmd.Run() == nil {
		return 0
	} else {
		return 1
	}
}
