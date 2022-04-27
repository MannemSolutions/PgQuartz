package jobs

import (
	"fmt"
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

func (cs Commands) Run(conns Connections) (err error) {
	for _, command := range cs {
		if err = command.Run(conns); err != nil {
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
	return clone
}

type Command struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Inline string `yaml:"inline"`
	// Home (~) is not resolved
	File    string            `yaml:"file"`
	Result  string            `yaml:"result"`
	Matrix  map[string]string `yaml:"matrix"`
	tmpFile *os.File
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

// ScriptFile returns a path to the script holding the command.
// This could be the symlink evaluated version of Command.File.
// Or this could be an executable tempfile with Command.Inline as contents.
// This is wat is used to run shell commands
func (c Command) ScriptFile() (scriptFile string) {
	var err error
	if c.Inline != "" {
		if c.tmpFile, err = ioutil.TempFile("", "pgQuartsInlineCommand"); err != nil {
			log.Panicf("error creating tempfile: %e", err)
		}
		if _, err = c.tmpFile.WriteString(c.Inline); err != nil {
			log.Panicf("error writing inline command to tempfile: %e", err)
		}
		// os.Chmod should also work on Windows
		if err = os.Chmod(c.tmpFile.Name(), 0700); err != nil {
			log.Panicf("error making inline tempfile script executable: %e", err)
		}
		return c.tmpFile.Name()
	}
	if err := c.VerifyScriptFile(); err != nil {
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

func (c Command) Run(conns Connections) (err error) {
	log.Infof("Running the following command: %s", c.Name)
	if c.Type == "" || c.Type == "shell" {
		return c.RunOsCommand()
	}
	return conns.Execute(c.Type, c.ScriptBody(), c.Result)
}

func (c Command) RunOsCommand() (err error) {
	exCommand := exec.Command("/bin/bash", c.ScriptFile())
	exCommand.Stdout = os.Stdout
	if err = exCommand.Run(); err != nil {
		return err
	}
	log.Infof("command %s successfully executed", c.Name)
	return nil
}
