import (
	"os"
	"path/filepath"
	"golang.org/x/sys/unix"
)

// Create a temporary directory for the container's filesystem
func createContainerRoot() (string, error) {
	dir, err := os.MkdirTemp("", "container-root")
	if err != nil {
		return "", err
	}
	return dir, nil
}

// RunContainer runs a command in a new namespace
func RunContainer(command string, args []string) error {
	// Create a temporary root filesystem for the container
	rootDir, err := createContainerRoot()
	if err != nil {
		return fmt.Errorf("failed to create container root: %w", err)
	}
	defer os.RemoveAll(rootDir) // Clean up

	// Mount the new root filesystem
	if err := unix.Chroot(rootDir); err != nil {
		return fmt.Errorf("failed to chroot: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Create a new process in a new namespace
	attr := &syscall.ProcAttr{
		Env: os.Environ(),
		Sys: &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER,
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
