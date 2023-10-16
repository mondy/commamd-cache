package command

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/google/shlex"
	"github.com/meltycat/commamd-cache/collection"
	"github.com/meltycat/commamd-cache/constant"
	"github.com/spf13/cobra"
)

var rootCommand = cobra.Command{
	Use:     constant.Name + " commands cache",
	Short:   "Cache the results of command output",
	Version: constant.Version,
	Args:    cobra.ExactArgs(2),
	RunE:    rootRun,
}

func Execute() {
	rootCommand.Execute()
}

func rootRun(command *cobra.Command, arguments []string) error {
	commandsFile := arguments[0]
	cacheFile := arguments[1]

	commandsModTime, err := modTime(commandsFile)
	if err != nil {
		return err
	}
	commands, err := readCommands(commandsFile)
	if err != nil {
		return err
	}

	cacheModTime, err := modTime(cacheFile)
	if os.IsNotExist(err) || commandsModTime.After(cacheModTime) {
		return writeCommandExecutionOutput(commands, cacheFile)
	}

	if ok, err := isNewestCommands(commands, cacheModTime); err != nil {
		return err
	} else if ok {
		return writeCommandExecutionOutput(commands, cacheFile)
	}

	cache, err := os.Open(cacheFile)
	if err != nil {
		return err
	}
	defer cache.Close()

	_, err = io.Copy(os.Stdout, cache)
	return err
}

func modTime(name string) (time.Time, error) {
	info, err := os.Stat(name)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

func readCommands(name string) ([][]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands [][]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		arguments, err := shlex.Split(scanner.Text())
		if err != nil {
			return nil, err
		}
		commands = append(commands, arguments)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

func isNewestCommands(commands [][]string, time time.Time) (bool, error) {
	return collection.SomeWithError(commands, func(command []string, _ int) (bool, error) {
		path, err := exec.LookPath(command[0])
		if err != nil {
			return false, err
		}

		info, err := os.Stat(path)
		if err != nil {
			return false, err
		}

		return info.ModTime().After(time), nil
	})
}

func writeCommandExecutionOutput(commands [][]string, name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, arguments := range commands {
		if err := executeCommand(arguments, file, os.Stdout); err != nil {
			return err
		}
	}

	return nil
}

func executeCommand(arguments []string, writer1, writer2 io.Writer) error {
	command := exec.Command(arguments[0], arguments[1:]...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	if err := command.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(writer2, io.TeeReader(stdout, writer1)); err != nil {
		return err
	}
	if err := command.Wait(); err != nil {
		return err
	}

	return nil
}
