package jobs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Commands []Command

func (cs Commands) Verify(stepName string, conns Connections) (errs []error) {
	for _, command := range cs {
		errs = append(errs, command.Verify(stepName, conns)...)
	}
	return errs
}

func (cs *Commands) Initialize() {
	for _, command := range *cs {
		command.Initialize()
	}
}

func (cs Commands) Run(conns Connections, args InstanceArguments) (err error) {
	for _, command := range cs {
		if err = command.Run(conns, args); err != nil {
			return err
		}
	}
	return nil
}

func (cs Commands) Clone() (clone Commands) {
	for _, c := range cs {
		clone = append(clone, Command{
			Type:   c.Type,
			Inline: c.Inline,
			File:   c.File,
			Matrix: c.Matrix,
		})
	}
	clone.Initialize()
	return clone
}

func (cs Commands) Rc() (rc int) {
	for _, command := range cs {
		rc += command.Rc
	}
	return rc
}

func (cs Commands) StdOut() (stdOut Result) {
	for _, command := range cs {
		for args := range command.stdOut {
			stdOut = append(stdOut, command.stdOut[args]...)
		}
	}
	return stdOut
}

func (cs Commands) StdErr() (stdErr Result) {
	for _, command := range cs {
		for args := range command.stdOut {
			stdErr = append(stdErr, command.stdErr[args]...)
		}
	}
	return stdErr
}

type Command struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Inline string `yaml:"inline,omitempty"`
	// Home (~) is not resolved
	File    string            `yaml:"file,omitempty"`
	stdOut  InstanceResult    `yaml:"-"`
	stdErr  InstanceResult    `yaml:"-"`
	Rc      int               `yaml:"-"`
	Test    string            `yaml:"test,omitempty"`
	Matrix  map[string]string `yaml:"matrix,omitempty"`
	tmpFile string
}

func (c Command) GetCommands() string {
	if c.Name != "" {
		return c.Name
	}
	if c.Inline != "" {
		return c.Inline
	}
	return c.File
}

func (c Command) VerifyScriptFile() (err error) {
	if c.Inline != "" {
		return nil
	}
	if c.Type != "shell" {
		return nil
	}
	// Check file exists
	if info, err := os.Stat(c.File); err != nil {
		return err
	} else {
		// Check file is executable by me
		// Requires a fix for Windows...
		mode := info.Mode()
		stat := info.Sys().(*syscall.Stat_t)
		if mode&0001 != 0 {
			return nil
		} else if mode&0100 != 0 && int(stat.Uid) == os.Getuid() {
			return nil
		} else if mode&0010 != 0 && int(stat.Gid) == os.Getgid() {
			return nil
		}
		return fmt.Errorf("script file %s is not executable by me (uid: %d, gid: %d)", c.File, os.Getuid(),
			os.Getgid())
	}
}

func (c Command) Verify(stepName string, conns Connections) (errs []error) {
	if c.Type == "" {
		if len(conns) == 1 {
			// This is fine. When only one, we use that.
		} else {
			errs = append(errs, fmt.Errorf(
				"please reference a specific Type for step command %s.%s, or just define only one Connection",
				stepName, c.Name))
		}
	} else if c.Type == "shell" {
		// Special type shell for running shell commands instead of db connection
	} else if _, exists := conns[c.Type]; !exists {
		errs = append(errs, fmt.Errorf("step command %s.%s references an unknown Type %s", stepName,
			c.Type, c.Name))
	}
	if err := c.VerifyScriptFile(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (c *Command) Initialize() {
	c.stdOut = make(InstanceResult)
	c.stdErr = make(InstanceResult)
}

// ScriptFile returns a path to the script holding the command.
// This could be the symlink evaluated version of Command.File.
// Or this could be an executable temporary file with Command.Inline as contents.
// This is wat is used to run shell commands
func (c *Command) ScriptFile() (scriptFile string) {
	var err error
	var tmpFile *os.File
	if c.Inline != "" {
		if tmpFile, err = ioutil.TempFile("", "pgQuartsInlineCommand"); err != nil {
			log.Panicf("error creating tempfile: %e", err)
		}
		c.tmpFile = tmpFile.Name()
		if _, err = tmpFile.WriteString(c.Inline); err != nil {
			log.Panicf("error writing inline command to tempfile: %e", err)
		} else if err = tmpFile.Close(); err != nil {
			log.Panicf("error closing the tmpfile: %e", err)
			// os.Chmod should also work on Windows
		} else if err = os.Chmod(tmpFile.Name(), 0600); err != nil {
			log.Panicf("error making inline tempfile script executable: %e", err)
		}
		return c.tmpFile
	}
	if err = c.VerifyScriptFile(); err != nil {
		log.Panicf("Cannot run script %s", c.File)
	}
	if scriptFile, err = filepath.EvalSymlinks(c.File); err != nil {
		log.Panicf("error while evaluating SymLinks: %e", err)
	}
	return scriptFile
}

// ScriptBody does the exact opposite of ScriptFile.
// For Command.File it reads the contents.
// In other situations it just returns Command.Inline.
// This is wat is used to run commands on database connections.
func (c Command) ScriptBody() (scriptBody string) {
	if c.Inline != "" {
		return c.Inline
	}
	scriptBodyBytes, err := os.ReadFile(c.File)
	if err != nil {
		log.Panicf("error while reading %s: %e", c.File, err)
	}
	return string(scriptBodyBytes)
}

func (c *Command) Run(conns Connections, args InstanceArguments) (err error) {
	log.Debugf("Running the following command: %s", c.GetCommands())
	if c.Type == "" || c.Type == "shell" {
		return c.RunOsCommand(args)
	}
	if c.stdOut[args.String()], err = conns.Execute(c.Type, c.ScriptBody(), args); err != nil {
		c.Rc = 1
		return err
	}
	return nil
}

func (c *Command) CleanTempFile() {
	if c.tmpFile != "" {
		log.Debugf("removing tmp file %s", c.tmpFile)
		if err := os.Remove(c.tmpFile); err != nil {
			log.Errorf("error removing file %s: %e", c.tmpFile, err)
		}
		c.tmpFile = ""
	}
}

func (c *Command) RunOsCommand(args InstanceArguments) (err error) {
	exCommand := exec.Command("/bin/bash", c.ScriptFile()) // #nosec
	exCommand.Env = args.AsEnv()
	var stdOut, stdErr bytes.Buffer
	exCommand.Stdout = io.MultiWriter(&stdOut)
	exCommand.Stderr = io.MultiWriter(&stdErr)
	defer c.CleanTempFile()
	if err = exCommand.Run(); err != nil {
		switch typedErr := err.(type) {
		case *exec.ExitError:
			c.Rc = typedErr.ExitCode()
		default:
			c.Rc = 1
		}
		return err
	}
	c.stdOut[args.String()] = NewResultFromString(stdOut.String())
	c.stdErr[args.String()] = NewResultFromString(stdErr.String())
	log.Debugf("command %s successfully executed", c.GetCommands())
	return nil
}
