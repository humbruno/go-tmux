package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"unicode/utf8"

	"github.com/charmbracelet/huh"
)

var (
	sessionName string
	numWindows  string
	attach      bool
)

func sessionNameExists(str string) bool {
	// -t= checks for exact match
	arg := fmt.Sprintf("-t=%s", str)
	cmd := exec.Command("tmux", "has-session", arg)
	err := cmd.Run()
	return err == nil
}

func isRunningInTmux() bool {
	_, exists := os.LookupEnv("TMUX")
	return exists
}

func main() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Session Name").
				Value(&sessionName).
				Placeholder("tmux-session").
				Validate(func(str string) error {
					if utf8.RuneCountInString(str) == 0 {
						return errors.New("Please enter at least 1 character.")
					}
					if sessionNameExists(str) {
						return errors.New("Session name already exists.")
					}
					return nil
				}),
			huh.NewInput().
				Title("Number of windows").
				Value(&numWindows).
				Placeholder("1").
				Validate(func(str string) error {
					num, err := strconv.Atoi(str)
					if err != nil {
						return errors.New("Please enter a valid (whole) number.")
					}
					if num < 1 {
						return errors.New("Please enter a number above zero.")
					}
					return nil
				}),
			huh.NewConfirm().
				Title("Attach to session automatically?").
				Inline(true).
				Value(&attach),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("> Creating session named [%s] with [%s] windows... \n", sessionName, numWindows)
	defer fmt.Println("All done!")

	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to create session, exiting program.")
	}

	n, _ := strconv.Atoi(numWindows)
	for i := 1; i < n; i++ {
		cmd := exec.Command("tmux", "new-window", "-d", "-t", sessionName)
		if err := cmd.Run(); err != nil {
			fmt.Println("Failed to create window in session.")
		}
	}

	if !attach {
		return
	}

	fmt.Println("> Attaching to new session...")
	cmd = exec.Command("tmux", "switch-client", "-t", sessionName)

	if !isRunningInTmux() {
		cmd = exec.Command("tmux", "attach", "-t", sessionName)
		// hands over control of the terminal to the tmux process
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("! Failed to attach to session.")
	}
}
