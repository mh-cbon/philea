package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestOnePathOneCmd(t *testing.T) {
	args := []string{"run", "main.go", "-p", "main_test.go", "echo %s-%d-%dname"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	out, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	sOut := string(out)
	expected := `echo ./main_test.go-./-.: ./main_test.go-./-.
`

	if sOut != expected {
		t.Errorf("Expected=%q, got=%q\n", expected, sOut)
	}
}

func TestOnePathTwoCmds(t *testing.T) {
	args := []string{"run", "main.go", "-S", "-p", "main_test.go", "echo hello", "echo world"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	out, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	sOut := string(out)
	expected := `echo hello: hello
echo world: world
`

	if sOut != expected {
		t.Errorf("Expected=%q, got=%q\n", expected, sOut)
	}
}

func TestTwoPathsOneCmd(t *testing.T) {
	args := []string{"run", "main.go", "-S", "-p", "main*", "echo %s"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	out, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	sOut := string(out)
	expected := `echo ./main.go: ./main.go
echo ./main_test.go: ./main_test.go
`
	if sOut != expected {
		t.Errorf("Expected=%q, got=%q\n", expected, sOut)
	}
}

func TestExecuteInSeries(t *testing.T) {
	args := []string{"run", "main.go", "-S", "-p", "main.go", "sleep 1", "sleep 1", "sleep 1"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	elapsed := 0
	shouldStop := false
	go func() {
		for {
			select {
			case <-time.After(1 * time.Millisecond):
				if shouldStop {
					break
				} else {
					elapsed++
				}
			}
		}
	}()

	_, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	shouldStop = true
	if elapsed < 3*1000 {
		t.Errorf("Expected to take a long time\n")
	}
}

func TestExecuteInParallel(t *testing.T) {
	args := []string{"run", "main.go", "-p", "main.go", "sleep 1", "sleep 1", "sleep 1"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	elapsed := 0
	shouldStop := false
	go func() {
		for {
			select {
			case <-time.After(1 * time.Millisecond):
				if shouldStop {
					break
				} else {
					elapsed++
				}
			}
		}
	}()

	_, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	shouldStop = true
	if elapsed > 2*1000 {
		t.Errorf("Expected to take a short time\n")
	}
}

func TestReportAFailure(t *testing.T) {
	args := []string{"run", "main.go", "-p", "main.go", "nopnopnop"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("Expected err!=nil, got err=%s\n", err)
	}
	if cmd.ProcessState.Success() == true {
		t.Errorf("Expected success=false, got success=%t\n", true)
	}
	sOut := string(out)
	expected := "nopnopnop: \n\n-------------\nThere were 1 error(s)\nexec: \"nopnopnop\": executable file not found in $PATH\nexit status 1\n"
	if sOut != expected {
		t.Errorf("Expected=%q, got=%q\n", expected, sOut)
	}
}

func TestDry(t *testing.T) {
	args := []string{"run", "main.go", "-p", "main.go", "-d", "mkdir -p tests/%fname"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	out, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	sOut := string(out)
	expected := `mkdir -p tests/main
`
	if sOut != expected {
		t.Errorf("Expected=%q, got=%q\n", expected, sOut)
	}
	if _, err := os.Stat("tests/main"); !os.IsNotExist(err) {
		t.Errorf("Expected directory %s to not exist\n", "tests/main")
	}
}

func TestTokens(t *testing.T) {
	args := []string{"run", "main.go", "-p", "main.go", "-d", "echo %fname", "echo %f", "echo %s", "echo %d", "echo %dname"}
	bin := "go"
	cmd := exec.Command(bin, args...)

	out, err := cmd.CombinedOutput()
	mustSucceed(t, cmd, err)
	sOut := string(out)
	expected := `echo main
echo main.go
echo ./main.go
echo ./
echo .
`
	if sOut != expected {
		t.Errorf("Expected=%q, got=%q\n", expected, sOut)
	}
}

func mustSucceed(t *testing.T, cmd *exec.Cmd, err error) bool {
	if err != nil {
		fmt.Println(err)
		t.Errorf("Expected err=nil, got err=%s\n", err)
		return false
	}
	if cmd.ProcessState.Success() == false {
		t.Errorf("Expected success=true, got success=%t\n", true)
		return false
	}
	return true
}
