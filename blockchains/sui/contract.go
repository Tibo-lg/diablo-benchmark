package sui

import (
	"bufio"
	"diablo-benchmark/core"
	"diablo-benchmark/util"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type application struct {
	logger  core.Logger
	text    []byte            // application binary
	entries map[string][]byte // function name -> hash
	parser  *util.ServiceProcess
	scanner *bufio.Scanner
}

type moveCompiler struct {
	logger core.Logger
	base   string // contracts directory
}

func newMoveCompiler(logger core.Logger, base string) *moveCompiler {
	return &moveCompiler{
		logger: logger,
		base:   base,
	}
}


func (this *moveCompiler) compile(name string) (*application, error) {
	var entries map[string][]byte  // empty
	
	sourceDir := this.base + "/" + name

	stdout, err := exec.Command("sui", "move", "build", "-p", sourceDir).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("sui move build failed.\nstdout: %s\nstderr: %v", stdout, err)
	}

	buildFile := sourceDir + "build/" + name + "/bytecode_modules/" + name + ".mv"  // name::name
	text, err := ioutil.ReadFile(buildFile)
	if (err != nil) {
		return nil, err
	}

	return &application{logger: this.logger, text: text, entries: entries, parser: ,scanner: bufio.NewScanner(parser)}
}
