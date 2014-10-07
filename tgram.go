package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
//    "time"
    "bufio"
    "io"
)

/*
TODO:
- allow it to register a callback
- put a flag to know if what it's getting from stdout is a response to a
  command or something produced by the system, to send to the callback
- find how to do select on a socket, or something, to avoid that 
  nasty polling of stdout
*/

const raw_prompt = "\x1b[K> \r"


type Telegram struct {
    tg_cli_path, tg_pub_path string
    ch_stdout chan []byte
    stdin *bufio.Writer
}

func (t *Telegram) Init() error {
    // start the command
    cmd := exec.Command(t.tg_cli_path, "-k", t.tg_pub_path)

    // handle stdout
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }
    out := bufio.NewReader(stdout)
    t.ch_stdout = make(chan []byte)
    go t.readlines(out, t.ch_stdout)

    // handle stdin
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return err
    }
    t.stdin = bufio.NewWriter(stdin)

    // actually start the process
    err = cmd.Start()
    if err != nil {
        return err
    }

    // this will consume the initial header with telegram version, etc
    t.read_response()
    return nil
}

func (t *Telegram) readlines(src *bufio.Reader, dst chan []byte) {
    for 1 == 1 {
        line, err := src.ReadBytes('\r')
        if err != nil && err != io.EOF {
            log.Fatal(err)
        }
        fmt.Printf("raw: %q\n", line)
        dst <- line
        //fmt.Printf("waiting stdout\n")
        //time.Sleep(500 * time.Millisecond)
    }
}

func (t *Telegram) read_response() string {
    /*
    TODO:
    - separate received stuff by \n
    - remove "> \r" in a clean line
    - put all lines in a list and return that
    */
    for 1 == 1 {
        received := <-t.ch_stdout
        fmt.Printf("received: %q\n", received)
        if string(received) == raw_prompt {
            break
        }
    }
    return "FIXME"
}

func (t *Telegram) Execute(order string) string {
    // FIXME: this is not working!!
    fmt.Printf("Sending command: %q\n", order)
    t.stdin.WriteString(order)
    resp := t.read_response()
    return resp
}

func (t *Telegram) Quit() {
    t.stdin.WriteString("quit\n")
}



func main() {
    // check parameters
    if len(os.Args) < 3 {
        log.Fatal("Usage: tgram <path-to-telegram-cli> <path-to-server.pub>")
    }
    tg_cli_path := os.Args[1]
    tg_pub_path := os.Args[2]

    telegram := &Telegram{tg_cli_path: tg_cli_path, tg_pub_path: tg_pub_path}
    err := telegram.Init()
    if err != nil {
        log.Fatal(err)
    }

    resp := telegram.Execute("contact_list")
    fmt.Printf("Resp: %q\n", resp)

    telegram.Quit()
}
