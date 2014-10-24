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

package main

import (
	"./telegram"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func show_incoming(origin, message string) {
	fmt.Printf("<---[%s] %q\n", origin, message)
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
	telegram, err := telegram.New(tg_cli_path, tg_pub_path, show_incoming)
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
