package integration

import (
	"bytes"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"testing"
	"time"
)

const (
	DefaultMaxWaitRetry = 300
)

type Step interface {
	Execute(prog tea.Program)
	Wait() bool
}

type Steps []Step

type InputStep struct {
	Msg tea.Msg
}

func (i InputStep) Execute(prog tea.Program) {
	prog.Send(i.Msg)
}

func (i InputStep) Wait() bool {
	return true
}

type WaitStep struct {
	Check func() bool
}

func (w WaitStep) Execute(prog tea.Program) {}

func (w WaitStep) Wait() bool {
	return w.Check()
}

func runProgramWithSteps(t *testing.T, program tea.Model, steps Steps) tea.Model {
	prog := tea.NewProgram(program, tea.WithInput(nil))
	done := make(chan struct{})
	finalModel := tea.Model(nil)

	go func() {
		var err error
		finalModel, err = prog.Run()
		if err != nil {
			t.Errorf("Program run failed: %v", err)
			return
		}
		close(done)
	}()

	for _, step := range steps {
		if waitStep, ok := step.(WaitStep); ok {
			retryCount := 0
			for {
				if waitStep.Wait() {
					break
				}

				if retryCount >= DefaultMaxWaitRetry {
					t.Errorf("Max retries reached while waiting for condition in WaitStep")
					return nil
				}

				retryCount++
				time.Sleep(100 * time.Millisecond)
			}
			time.Sleep(3 * time.Second)
		}

		step.Execute(*prog)
		time.Sleep(100 * time.Millisecond)
	}

	<-done
	return finalModel
}

func clearTestDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		panic(fmt.Sprintf("failed to remove test directory: %v", err))
	}
}

func compareFiles(t *testing.T, expectedPath, actualPath string) {
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	actualContent, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("Failed to read actual file: %v", err)
	}

	if !bytes.Equal(expectedContent, actualContent) {
		t.Errorf("Files do not match:\nExpected (%s):\n%s\n\nActual (%s):\n%s\n",
			expectedPath, expectedContent, actualPath, actualContent)
	}
}
