package main

import (
	"fmt"
	"io"
	"os"
)

var fopenModes = map[string]int{
	"r":  os.O_RDONLY,
	"w":  os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
	"a":  os.O_WRONLY | os.O_CREATE | os.O_APPEND,
	"r+": os.O_RDWR,
	"w+": os.O_RDWR | os.O_CREATE | os.O_TRUNC,
	"a+": os.O_RDWR | os.O_CREATE | os.O_APPEND,
}

func init() {
	// fopen: P M -> S — open file at path P with mode M
	register("fopen", func(m *Machine) {
		m.NeedStack(2, "fopen")
		mode := m.Pop()
		path := m.Pop()
		if path.Typ != TypeString || mode.Typ != TypeString {
			joyErr("fopen: two strings expected")
		}
		flags, ok := fopenModes[mode.Str]
		if !ok {
			joyErr("fopen: invalid mode %q", mode.Str)
		}
		f, err := os.OpenFile(path.Str, flags, 0644)
		if err != nil {
			m.Push(FileVal(nil, path.Str))
			return
		}
		m.Push(FileVal(f, path.Str))
	})

	// fclose: S -> — close file
	register("fclose", func(m *Machine) {
		m.NeedStack(1, "fclose")
		a := m.Pop()
		if a.Typ != TypeFile {
			joyErr("fclose: file expected")
		}
		if a.File != nil {
			a.File.Close()
		}
	})

	// feof: S -> S B — check if at end of file
	register("feof", func(m *Machine) {
		m.NeedStack(1, "feof")
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("feof: open file expected")
		}
		cur, _ := a.File.Seek(0, io.SeekCurrent)
		end, _ := a.File.Seek(0, io.SeekEnd)
		a.File.Seek(cur, io.SeekStart)
		m.Push(BoolVal(cur >= end))
	})

	// ferror: S -> S B — check for error (always false in simple impl)
	register("ferror", func(m *Machine) {
		m.NeedStack(1, "ferror")
		a := m.Peek()
		if a.Typ != TypeFile {
			joyErr("ferror: file expected")
		}
		m.Push(BoolVal(false))
	})

	// fflush: S -> S — flush file
	register("fflush", func(m *Machine) {
		m.NeedStack(1, "fflush")
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fflush: open file expected")
		}
		a.File.Sync()
	})

	// fgets: S -> S L — read line as list of characters
	register("fgets", func(m *Machine) {
		m.NeedStack(1, "fgets")
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fgets: open file expected")
		}
		var chars []Value
		buf := make([]byte, 1)
		for {
			n, err := a.File.Read(buf)
			if n > 0 {
				chars = append(chars, CharVal(int64(buf[0])))
				if buf[0] == '\n' {
					break
				}
			}
			if err != nil {
				break
			}
		}
		if chars == nil {
			chars = []Value{}
		}
		m.Push(ListVal(chars))
	})

	// fgetch: S -> S C — read single character; push -1 on EOF
	register("fgetch", func(m *Machine) {
		m.NeedStack(1, "fgetch")
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fgetch: open file expected")
		}
		buf := make([]byte, 1)
		n, _ := a.File.Read(buf)
		if n == 0 {
			m.Push(IntVal(-1))
		} else {
			m.Push(CharVal(int64(buf[0])))
		}
	})

	// fread: S I -> S L — read I bytes as list of integers
	register("fread", func(m *Machine) {
		m.NeedStack(2, "fread")
		count := m.Pop()
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fread: open file expected")
		}
		buf := make([]byte, count.Int)
		n, _ := a.File.Read(buf)
		chars := make([]Value, n)
		for i := 0; i < n; i++ {
			chars[i] = IntVal(int64(buf[i]))
		}
		m.Push(ListVal(chars))
	})

	// fwrite: S L -> S — write list of integers as bytes
	register("fwrite", func(m *Machine) {
		m.NeedStack(2, "fwrite")
		data := m.Pop()
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fwrite: open file expected")
		}
		if data.Typ != TypeList {
			joyErr("fwrite: list expected")
		}
		buf := make([]byte, len(data.List))
		for i, v := range data.List {
			buf[i] = byte(v.Int)
		}
		a.File.Write(buf)
	})

	// fput: S X -> S — write value string representation to file
	register("fput", func(m *Machine) {
		m.NeedStack(2, "fput")
		x := m.Pop()
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fput: open file expected")
		}
		fmt.Fprint(a.File, x.String())
	})

	// fputch: S C -> S — write single character to file
	register("fputch", func(m *Machine) {
		m.NeedStack(2, "fputch")
		ch := m.Pop()
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fputch: open file expected")
		}
		if ch.Typ == TypeChar || ch.Typ == TypeInteger {
			fmt.Fprint(a.File, string(rune(ch.Int)))
		} else {
			fmt.Fprint(a.File, ch.String())
		}
	})

	// fputchars: S Str -> S — write string without quotes to file
	register("fputchars", func(m *Machine) {
		m.NeedStack(2, "fputchars")
		s := m.Pop()
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fputchars: open file expected")
		}
		if s.Typ == TypeString {
			fmt.Fprint(a.File, s.Str)
		} else {
			fmt.Fprint(a.File, s.String())
		}
	})

	registerAlias("fputstring", "fputchars")

	// fseek: S P W -> S — seek in file (W: 0=start, 1=current, 2=end)
	register("fseek", func(m *Machine) {
		m.NeedStack(3, "fseek")
		whence := m.Pop()
		pos := m.Pop()
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("fseek: open file expected")
		}
		a.File.Seek(pos.Int, int(whence.Int))
	})

	// ftell: S -> S I — get current file position
	register("ftell", func(m *Machine) {
		m.NeedStack(1, "ftell")
		a := m.Peek()
		if a.Typ != TypeFile || a.File == nil {
			joyErr("ftell: open file expected")
		}
		pos, _ := a.File.Seek(0, io.SeekCurrent)
		m.Push(IntVal(pos))
	})

	// fremove: P -> B — remove file at path P
	register("fremove", func(m *Machine) {
		m.NeedStack(1, "fremove")
		path := m.Pop()
		if path.Typ != TypeString {
			joyErr("fremove: string expected")
		}
		err := os.Remove(path.Str)
		m.Push(BoolVal(err == nil))
	})

	// frename: P1 P2 -> B — rename file P1 to P2
	register("frename", func(m *Machine) {
		m.NeedStack(2, "frename")
		newPath := m.Pop()
		oldPath := m.Pop()
		if oldPath.Typ != TypeString || newPath.Typ != TypeString {
			joyErr("frename: two strings expected")
		}
		err := os.Rename(oldPath.Str, newPath.Str)
		m.Push(BoolVal(err == nil))
	})

	// stdin: -> S — push stdin file
	register("stdin", func(m *Machine) {
		m.Push(FileVal(os.Stdin, "stdin"))
	})

	// stdout: -> S — push stdout file
	register("stdout", func(m *Machine) {
		m.Push(FileVal(os.Stdout, "stdout"))
	})

	// stderr: -> S — push stderr file
	register("stderr", func(m *Machine) {
		m.Push(FileVal(os.Stderr, "stderr"))
	})
}
