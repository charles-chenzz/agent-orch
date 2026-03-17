package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

func TestListTmuxSessions_ParsesSessions(t *testing.T) {
	output := "agent-orch-abc\t1\t1700000000\t1700000100\t/Users/test/a\n" +
		"agent-orch-def\t0\t1700000001\t1700000200\t/Users/test/b\n" +
		"other\t1\t1700000002\t1700000300\t/ignore"

	reset := stubExecCommand(output, 0)
	defer reset()

	m := &Manager{hasTmux: true, tmuxPath: "tmux"}
	infos, err := m.listTmuxSessions()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(infos))
	}

	byID := map[string]SessionInfo{}
	for _, info := range infos {
		byID[info.ID] = info
	}

	infoABC, ok := byID["abc"]
	if !ok {
		t.Fatalf("expected session abc")
	}
	if infoABC.State != string(StateRunning) {
		t.Fatalf("expected abc state %q, got %q", StateRunning, infoABC.State)
	}
	if infoABC.CWD != "/Users/test/a" {
		t.Fatalf("expected abc cwd /Users/test/a, got %q", infoABC.CWD)
	}
	if !infoABC.CreatedAt.Equal(time.Unix(1700000000, 0)) {
		t.Fatalf("unexpected abc createdAt: %v", infoABC.CreatedAt)
	}
	if !infoABC.LastActive.Equal(time.Unix(1700000100, 0)) {
		t.Fatalf("unexpected abc lastActive: %v", infoABC.LastActive)
	}

	infoDEF, ok := byID["def"]
	if !ok {
		t.Fatalf("expected session def")
	}
	if infoDEF.State != string(StateDetached) {
		t.Fatalf("expected def state %q, got %q", StateDetached, infoDEF.State)
	}
	if infoDEF.CWD != "/Users/test/b" {
		t.Fatalf("expected def cwd /Users/test/b, got %q", infoDEF.CWD)
	}
	if !infoDEF.CreatedAt.Equal(time.Unix(1700000001, 0)) {
		t.Fatalf("unexpected def createdAt: %v", infoDEF.CreatedAt)
	}
	if !infoDEF.LastActive.Equal(time.Unix(1700000200, 0)) {
		t.Fatalf("unexpected def lastActive: %v", infoDEF.LastActive)
	}
}

func TestListTmuxSessions_NoServer(t *testing.T) {
	reset := stubExecCommand("no server running on /tmp/tmux-1000/default", 1)
	defer reset()

	m := &Manager{hasTmux: true, tmuxPath: "tmux"}
	infos, err := m.listTmuxSessions()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(infos) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(infos))
	}
}

func stubExecCommand(output string, exitCode int) func() {
	execCommand = func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", name)
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"TMUX_OUTPUT="+output,
			"TMUX_EXIT="+strconv.Itoa(exitCode),
		)
		return cmd
	}
	return func() {
		execCommand = exec.Command
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	output := os.Getenv("TMUX_OUTPUT")
	if output != "" {
		fmt.Fprint(os.Stdout, output)
	}

	exitCode := 0
	if value := os.Getenv("TMUX_EXIT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			exitCode = parsed
		}
	}
	os.Exit(exitCode)
}
