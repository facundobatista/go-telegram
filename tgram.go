package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "bufio"
    "io"
    "strings"
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
const prefix = "\x1b[K"


type Telegram struct {
    tg_cli_path, tg_pub_path string
    ch_stdout chan string
    stdin *bufio.Writer
    issued_command string
}

func (t *Telegram) Init() error {
    // start the command
    cmd := exec.Command(t.tg_cli_path, "-C", "-k", t.tg_pub_path)

    // handle stdout
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }
    out := bufio.NewReader(stdout)
    t.ch_stdout = make(chan string)
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

    // this will populate backend telegram with contacts, needed to send messages
    t.ListContacts()
    return nil
}

func (t *Telegram) readlines(src *bufio.Reader, dst chan string) {
    for 1 == 1 {
        line, err := src.ReadString('\r')
        if err != nil && err != io.EOF {
            log.Fatal(err)
        }
        dst <- line
    }
}

func (t *Telegram) read_response() []string {
    useful := []string{}
    fmt.Printf("read response after command: %q\n", t.issued_command)
    for 1 == 1 {
        received := <-t.ch_stdout
        fmt.Printf("received: %q\n", received)
        if string(received) == raw_prompt {
            // if has something useful already, this is the end of it; otherwise
            // it's just garbage before getting any result
            if len(useful) > 0 {
                break
            } else {
                continue
            }
        }

        // remove the control prefix, and a prompt if was needed
        received = strings.TrimPrefix(received, prefix)
        received = strings.TrimPrefix(received, "> ")

        // split the string in lines, add those that are not a prompt
        lines := strings.Split(received, "\n")
        for _, v := range lines {
            if v != "> \r" {
                useful = append(useful, v)

                // if it matches the issued command it means that so far it
                // was telegram echoing us
                if v == t.issued_command {
                    useful = []string{}
                }
            }
        }
    }
    fmt.Printf("useful: %q\n", useful)
    return useful
}

func (t *Telegram) execute(order string) []string {
    fmt.Printf("Sending command: %q\n", order)
    t.issued_command = order
    t.stdin.WriteString(order + "\n")
    t.stdin.Flush()
    resp := t.read_response()
    return resp
}

func (t *Telegram) ListContacts() []string {
    fmt.Printf("Listing contacts\n")
    contacts := []string{}
    resp := t.execute("contact_list")
    for _, v := range resp {
        if strings.HasPrefix(v, "User") && strings.HasSuffix(v, "updated photo") {
            continue
        }
        contacts = append(contacts, v)
    }
    return contacts
}

func (t *Telegram) SendMessage(dest, message string) {
    fmt.Printf("Sending message to %q: %q\n", dest, message)
    dest = strings.Replace(dest, " ", "_", -1)
    parts := []string{"msg", dest, message}
    resp := t.execute(strings.Join(parts, " "))
    // FIXME: needs to consume better what comes from telegram backend...
    fmt.Printf("===== send resp: %q\n", resp)
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

    // start Telegram backend
    fmt.Printf("Hello! Starting backend...\n")
    telegram := &Telegram{tg_cli_path: tg_cli_path, tg_pub_path: tg_pub_path}
    err := telegram.Init()
    if err != nil {
        log.Fatal(err)
    }

    // start dialog with user
    reader := bufio.NewReader(os.Stdin)
    fmt.Printf("Done! Allowed: quit, send, list-contacts\n")

    // main user interface loop
    should_quit := false
    for 1 == 1 {
        fmt.Printf(">> ")
        text, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            } else {
                log.Fatal(err)
            }
        }
        text = strings.TrimSpace(text)
        tokens := strings.Split(text, " ")
        fmt.Printf("=== user: %q\n", tokens)
        switch tokens[0] {
            case "quit":
                should_quit = true
            case "list-contacts":
                contacts := telegram.ListContacts()
                for _, v := range contacts {
                    fmt.Print(v + "\n")
                }
            case "send":
                dest := tokens[1]
                msg := strings.Join(tokens[2:], " ")
                telegram.SendMessage(dest, msg)
        }
        if should_quit {
            break
        }
    }

    // clean up and die
    fmt.Printf("Quitting\n")
    telegram.Quit()
}
