package cli

import (
	"errors"
	"testing"
)

func TestFlagSetLongFlags(t *testing.T) {
	fs := NewFlagSet("test")
	name := fs.String("name", "guest", "display name")
	useTLS := fs.Bool("tls", false, "enable tls")

	if _, err := fs.Parse([]string{"--name", "alice", "--tls"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if *name != "alice" {
		t.Fatalf("name = %q, want alice", *name)
	}
	if !*useTLS {
		t.Fatal("tls should be true")
	}
}

func TestFlagSetInlineValues(t *testing.T) {
	fs := NewFlagSet("test")
	addr := fs.String("addr", "localhost:9000", "server address")

	if _, err := fs.Parse([]string{"--addr=example.com:443"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if *addr != "example.com:443" {
		t.Fatalf("addr = %q", *addr)
	}
}

func TestFlagSetRejectsSingleDashLongFlag(t *testing.T) {
	fs := NewFlagSet("test")
	fs.String("listen", ":9000", "listen address")

	_, err := fs.Parse([]string{"-listen", ":8080"})
	if err == nil {
		t.Fatal("expected error for -listen")
	}
	if got := err.Error(); got != `unknown flag "-listen" (long flags use --listen)` {
		t.Fatalf("error = %q", got)
	}
}

func TestFlagSetHelp(t *testing.T) {
	fs := NewFlagSet("test")
	fs.Usage = func() {}

	_, err := fs.Parse([]string{"--help"})
	if !errors.Is(err, ErrHelp) {
		t.Fatalf("err = %v, want ErrHelp", err)
	}
}

func TestRunVersion(t *testing.T) {
	if code := Run([]string{"--version"}); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRunHelp(t *testing.T) {
	if code := Run([]string{"help", "server"}); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
}
