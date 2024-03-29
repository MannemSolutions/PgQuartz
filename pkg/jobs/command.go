package jobs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Commands []*Command

func (cs Commands) Verify(stepName string, conns Connections) (errs []error) {
	for _, command := range cs {
		errs = append(errs, command.Verify(stepName, conns)...)
	}
	return errs
}

func (cs *Commands) Run(conns Connections, args InstanceArguments) (err error) {
	for _, command := range *cs {
		if err = command.Run(conns, args); err != nil {
			return err
		}
	}
	return nil
}

func (cs Commands) Clone() (clone Commands) {
	for _, c := range cs {
		clone = append(clone, c.Clone())
	}
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
		stdOut = append(stdOut, command.stdOut...)
	}
	return stdOut
}

func (cs Commands) StdErr() (stdErr Result) {
	for _, command := range cs {
		stdErr = append(stdErr, command.stdErr...)
	}
	return stdErr
}

type Command struct {
	// Home (~) is not resolved
	File      string `yaml:"file,omitempty"`
	Name      string `yaml:"name"`
	Role      string `yaml:"role"`
	Type      string `yaml:"type"`
	Inline    string `yaml:"inline,omitempty"`
	BatchMode bool   `yaml:"batchMode"`
	stdOut    Result `yaml:"-"`
	stdErr    Result `yaml:"-"`
	Rc        int    `yaml:"-"`
	tmpFile   string
}

func (c Command) Clone() *Command {
	return &Command{
		Name:      c.Name,
		Role:      c.Role,
		Type:      c.Type,
		Inline:    c.Inline,
		File:      c.File,
		BatchMode: c.BatchMode,
	}
}

func (c Command) String() string {
	var cmd string
	if c.Inline != "" {
		cmd = fmt.Sprintf("inline='%s'", strings.Replace(
			strings.Replace(c.Inline, "\n", "\\n", -1), "'", "''", -1))
	} else {
		cmd = fmt.Sprintf("file=%s", c.File)
	}
	return fmt.Sprintf("name='%s', type=%s, %s", strings.Replace(c.Name, "'", "''", -1), c.Type, cmd)
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
// Or this could be an executable temporary file with Command.Inline as contents.
// This is wat is used to run shell commands
func (c *Command) ScriptFile() (scriptFile string) {
	var err error
	var tmpFile *os.File
	if c.Inline != "" {
		if tmpFile, err = os.CreateTemp("", "pgQuartsInlineCommand"); err != nil {
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
func (c Command) ScriptBody() (string, error) {
	if c.Inline != "" {
		return c.Inline, nil
	}
	scriptBodyBytes, err := os.ReadFile(c.File)
	if err != nil {
		return "", nil
	}
	return string(scriptBodyBytes), nil
}

func (c *Command) Run(conns Connections, args InstanceArguments) (err error) {
	log.Infof("Running command: %s, args: %s", c.String(), args.String())
	if c.Type == "" || c.Type == "shell" {
		return c.RunOsCommand(args)
	}
	if body, err := c.ScriptBody(); err != nil {
		return err
	} else if c.stdOut, err = conns.Execute(c.Type, c.Role, body, c.BatchMode, args); err != nil {
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
	c.stdOut = NewResultFromString(stdOut.String())
	c.stdErr = NewResultFromString(stdErr.String())
	log.Debugf("command %s successfully executed", c.String())
	return nil
}
