package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// RunContainer runs a command in a new namespace
func RunContainer(command string, args []string) error {
	// Create a new process in a new namespace
	attr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Sys: &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID,
			Unshareflags: syscall.CLONE_NEWUSER,
		},
	}

	// Start the command
	pid, err := syscall.ForkExec(command, args, attr)
	if err != nil {
		return fmt.Errorf("failed to fork exec: %w", err)
	}

	// Wait for the command to finish
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for process: %w", err)
	}

	if ws.Exited() {
		fmt.Printf("Process exited with status: %d\n", ws.ExitStatus())
	} else {
		fmt.Println("Process did not exit normally")
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rawcontainers <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	err := RunContainer(command, args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
