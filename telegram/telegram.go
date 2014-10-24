/*
Copyright 2014 Facundo Batista

This program is free software: you can redistribute it and/or modify it
under the terms of the GNU General Public License version 3, as published
by the Free Software Foundation.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranties of
MERCHANTABILITY, SATISFACTORY QUALITY, or FITNESS FOR A PARTICULAR
PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package telegram

import (
	"../logging"
	"bufio"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

const prompt = "> \r"
const control_prefix = "\x1b[K"

var logger = logging.New(logging.LevelError)

type callback func(string, string)

type telegram struct {
	ch_stdout                chan string
	stdin                    *bufio.Writer
	issued_command           string
	incoming_callback        callback
	command_mode             bool
}

func New(tg_cli_path string, tg_pub_path string, incoming_callback callback) (*telegram, error) {
    t := new(telegram)
    t.incoming_callback = incoming_callback

	// start the command
	cmd := exec.Command(tg_cli_path, "-C", "-k", tg_pub_path)

	// handle stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	out := bufio.NewReader(stdout)
	t.ch_stdout = make(chan string, 3)
	go t.readlines(out, t.ch_stdout)

	// handle stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	t.stdin = bufio.NewWriter(stdin)

	// actually start the process
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	// this will consume the initial header with telegram version, etc
	t.command_mode = true
	t.read_response()

	// this will populate backend telegram with contacts, needed to send messages
	t.ListContacts()
	t.command_mode = false
	return t, nil
}

func (t *telegram) readlines(src *bufio.Reader, dst chan string) {
	// the regex to match incoming messages
	re_incoming := regexp.MustCompile("\\[\\d\\d:\\d\\d\\] (.*) >>> (.*)")

	for 1 == 1 {
		line, err := src.ReadString('\r')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		logger.Debug("raw: %q\n", line)

		// remove the control prefix, and a prompt if was needed
		line = strings.TrimPrefix(line, control_prefix)

		// discard the notifications (we may want to execute a callback
		// with these in a future)
		if strings.HasPrefix(line, "User ") {
			logger.Debug("Discarding notification: %q\n", line)
			continue
		}

		// execute a callback with incoming messages (these are not part
		// of the command responses!)
		match := re_incoming.FindStringSubmatch(line)
		if match != nil {
			t.incoming_callback(match[1], match[2]) // user, message
			continue
		}

		if t.command_mode {
			logger.Debug("Sending: %q\n", line)
			dst <- line
		}
	}
}

func (t *telegram) read_response() []string {
	useful := []string{}
	logger.Info("read response after command: %q\n", t.issued_command)
	for 1 == 1 {
		received := <-t.ch_stdout
		logger.Info("received: %q\n", received)
		if received == prompt {
			// if has something useful already, this is the end of it; otherwise
			// it's just garbage before getting any result
			if len(useful) > 0 {
				break
			} else {
				continue
			}
		}

		received = strings.TrimPrefix(received, "> ")

		// split the string in lines, add those that are not a prompt
		lines := strings.Split(received, "\n")
		for _, v := range lines {
			if v == prompt {
				continue
			}

			useful = append(useful, v)

			// if it matches the issued command it means that so far it
			// was telegram echoing us
			if v == t.issued_command {
				useful = []string{}
			}
		}
	}
	t.command_mode = false
	logger.Info("useful: %q\n", useful)
	return useful
}

func (t *telegram) execute(order string) []string {
	logger.Info("Sending command: %q\n", order)
	t.issued_command = order
	t.command_mode = true
	t.stdin.WriteString(order + "\n")
	t.stdin.Flush()
	resp := t.read_response()
	return resp
}

func (t *telegram) ListContacts() []string {
	logger.Info("Listing contacts\n")
	return t.execute("contact_list")
}

func (t *telegram) SendMessage(dest, message string) {
	logger.Info("Sending message to %q: %q\n", dest, message)
	dest = strings.Replace(dest, " ", "_", -1)
	parts := []string{"msg", dest, message}
	resp := t.execute(strings.Join(parts, " "))
	logger.Debug("Response to send message: %q\n", resp)
}

func (t *telegram) Quit() {
	t.stdin.WriteString("quit\n")
}
