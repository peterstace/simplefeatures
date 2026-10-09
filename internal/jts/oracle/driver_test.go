package oracle

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// driver is the JVM side of the oracle: tools/oracle/Driver.java running
// JTS 1.20.0.
type driver struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out *bufio.Reader
}

// cacheDir is where tools/setup.sh puts JTS and the compiled oracle driver.
const cacheDir = "../tools/.cache"

// startDriver compiles (if needed) and starts the Java driver, or skips the
// test if Java or the JTS cache is unavailable.
func startDriver(t *testing.T) *driver {
	t.Helper()
	jtsClasses := filepath.Join(cacheDir, "jts-classes")
	if _, err := os.Stat(jtsClasses); err != nil {
		t.Skip("JTS classes not found; run internal/jts/tools/setup.sh")
	}
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java not found")
	}
	classes := filepath.Join(cacheDir, "oracle-classes")
	src := "../tools/oracle/Driver.java"
	if stale(filepath.Join(classes, "Driver.class"), src) {
		out, err := exec.Command("javac", "-nowarn", "-cp", jtsClasses, "-d", classes, src).CombinedOutput()
		if err != nil {
			t.Fatalf("compiling driver: %v\n%s", err, out)
		}
	}
	cmd := exec.Command("java",
		// The runtime's default locale is en-US.
		"-Duser.language=en", "-Duser.country=US",
		"-XX:+UnlockDiagnosticVMOptions", "-XX:-InlineMathNatives",
		"-XX:+UnlockExperimentalVMOptions", "-XX:hashCode=3",
		"-Xss64m",
		"-cp", jtsClasses+string(os.PathListSeparator)+classes,
		"Driver", jtsClasses)
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	d := &driver{cmd: cmd, in: in, out: bufio.NewReaderSize(out, 1<<20)}
	t.Cleanup(func() {
		d.in.Close()
		_ = d.cmd.Wait()
	})
	return d
}

func stale(target, src string) bool {
	ti, err := os.Stat(target)
	if err != nil {
		return true
	}
	si, err := os.Stat(src)
	return err != nil || si.ModTime().After(ti.ModTime())
}

// call sends one request line and returns the reply line.
func (d *driver) call(fields ...string) string {
	if _, err := fmt.Fprintln(d.in, strings.Join(fields, " ")); err != nil {
		panic(err)
	}
	line, err := d.out.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSuffix(line, "\n")
}
