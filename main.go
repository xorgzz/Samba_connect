package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	curses "github.com/gbin/goncurses"
)

type Server struct {
	Name     string
	User     string
	Password string
	Host     string
	Server   string
}

func pretty_exit(message string) {
	fmt.Printf("%s\n", message)
	os.Exit(0)
}

func shell_mode(args []string) {
	var exit = func() {
		fmt.Println("Proper use:\n$ executable user:password host_mount:server")
		os.Exit(0)
	}
	var user string
	var password string
	var host string
	var server string

	if len(args) == 2 {
		if len(strings.Split(args[0], ":")) == 2 {
			user, password = strings.Split(args[0], ":")[0], strings.Split(args[0], ":")[1]
		} else {
			exit()
		}
		if len(strings.Split(args[1], ":")) == 2 {
			host, server = strings.Split(args[1], ":")[0], strings.Split(args[1], ":")[1]
		} else {
			exit()
		}
		if !is_mounted(host) {
			mount_samba(user, password, host, server, false)
		}
	} else if len(args) == 1 {
		host = args[0]
		dismount_samba(host, false)
	} else {
		exit()
	}
}

func exec_path() string {
	exePath, err := os.Executable()
	if err != nil {
		pretty_exit("Can't find exec path")
	}
	exeDir := filepath.Dir(exePath)

	return exeDir
}

func print_banner(stdscr *curses.Window) int {
	read_banner_file := func() string {
		content, err := os.ReadFile(fmt.Sprintf("%s/banner.txt", exec_path()))
		if err != nil {
			log.Fatal(err)
		}
		return string(content)
	}

	banner := strings.Split(read_banner_file(), "\n")
	y_position := 0
	actual_banner := 6
	for i, line := range banner {
		if i < actual_banner {
			stdscr.AttrOn(curses.ColorPair(1))
		}
		stdscr.MovePrint(i+1, 2, line)
		y_position = i + 1
		if i < actual_banner {
			stdscr.AttrOff(curses.ColorPair(1))
		}
	}

	return y_position + 1
}

func print_menu(stdscr *curses.Window, highlight int, servers []Server) int {
	y_position := print_banner(stdscr)
	y_position += 1
	for i, s := range servers {
		if i == highlight {
			stdscr.AttrOn(curses.ColorPair(2))
		}
		stdscr.MovePrint(y_position, 5, fmt.Sprintf(" %s ", s.Name))
		if i == highlight {
			stdscr.AttrOff(curses.ColorPair(2))
		}
		y_position += 1
	}

	return y_position
}

func tui_mode() {

	exeDir := exec_path()

	content, err := os.ReadFile(fmt.Sprintf("%s/servers.json", exeDir))
	if err != nil {
		pretty_exit("Can't find/open servers.json")
	}

	var servers []Server
	err = json.Unmarshal(content, &servers)
	if err != nil {
		pretty_exit("Can't parse json file")
	}

	// Here Ncurses start
	stdscr, err := curses.Init()
	if err != nil {
		pretty_exit("Error regarding ncurses")
	}
	defer curses.End()
	curses.StartColor()
	curses.InitPair(1, curses.C_RED, curses.C_BLACK)
	curses.InitPair(2, curses.C_BLACK, curses.C_WHITE)
	curses.InitPair(3, curses.C_WHITE, curses.C_GREEN)
	curses.InitPair(4, curses.C_WHITE, curses.C_RED)
	height, width := stdscr.MaxYX()
	curses.CBreak(true)
	stdscr.Keypad(true)

	if width < 60 || height < 24 {
		pretty_exit("Your terminal is too small")
	}

	highlight := 0
	for {
		print_menu(stdscr, highlight, servers)

		stdscr.MovePrint(0, 0, " ")
		c := stdscr.GetChar()
		if c == 'q' {
			break
		} else if c == curses.KEY_DOWN {
			if highlight < len(servers)-1 {
				highlight += 1
			}
		} else if c == curses.KEY_UP {
			if highlight > 0 {
				highlight -= 1
			}
		}

		stdscr.Clear()
		stdscr.Refresh()

		if c == 'd' {
			success := dismount_samba(servers[highlight].Host, true)
			if success {
				stdscr.AttrOn(curses.ColorPair(3))
				stdscr.MovePrint(0, 2, "Successful dismount")
				stdscr.AttrOff(curses.ColorPair(3))
			} else {
				stdscr.AttrOn(curses.ColorPair(4))
				stdscr.MovePrint(0, 2, "Failed to dismount")
				stdscr.AttrOff(curses.ColorPair(4))
			}
		}

		if c == 'm' {
			if servers[highlight].User != "" && servers[highlight].Password != "" {
				success := mount_samba(servers[highlight].User, servers[highlight].Password, servers[highlight].Host, servers[highlight].Server, true)
				if success {
					stdscr.AttrOn(curses.ColorPair(3))
					stdscr.MovePrint(0, 2, "Successful mount")
					stdscr.AttrOff(curses.ColorPair(3))
				} else {
					stdscr.AttrOn(curses.ColorPair(4))
					stdscr.MovePrint(0, 2, "Failed to mount")
					stdscr.AttrOff(curses.ColorPair(4))
				}
			} else {
				user := servers[highlight].User
				password := servers[highlight].Password
				if user == "" {
					msg := "User: "
					for {
						y_position := print_menu(stdscr, highlight, servers)
						y_position += 1
						stdscr.MovePrint(y_position, 6, msg)
						stdscr.MovePrint(y_position, 6+len(msg), user)
						ch := stdscr.GetChar()
						if ch == '\n' {
							stdscr.Clear()
							stdscr.Refresh()
							break
						} else if ch == curses.KEY_BACKSPACE {
							if len(user) > 0 {
								user = user[:len(user)-1]
							}
						} else {
							user = fmt.Sprintf("%s%c", user, ch)
						}

						stdscr.Clear()
						stdscr.Refresh()
					}
				}
				if password == "" {
					msg := "Password: "
					for {
						y_position := print_menu(stdscr, highlight, servers)
						y_position += 1
						stdscr.MovePrint(y_position, 6, msg)
						// stdscr.MovePrint(y_position, 6+len(msg), password) // Don't display your password
						ch := stdscr.GetChar()
						if ch == '\n' {
							stdscr.Clear()
							stdscr.Refresh()
							break
						} else if ch == curses.KEY_BACKSPACE {
							if len(password) > 0 {
								password = password[:len(password)-1]
							}
						} else {
							password = fmt.Sprintf("%s%c", password, ch)
						}

						stdscr.Clear()
						stdscr.Refresh()
					}
				}
				success := mount_samba(user, password, servers[highlight].Host, servers[highlight].Server, true)
				if success {
					stdscr.AttrOn(curses.ColorPair(3))
					stdscr.MovePrint(0, 2, "Successful mount")
					stdscr.AttrOff(curses.ColorPair(3))
				} else {
					stdscr.AttrOn(curses.ColorPair(4))
					stdscr.MovePrint(0, 2, "Failed to mount")
					stdscr.AttrOff(curses.ColorPair(4))
				}
			}
		}
	}

}

func mount_samba(user string, password string, host string, server string, tui_mode bool) bool {
	shell := exec.Command("mount", "-t", "cifs", server, host, "-o", fmt.Sprintf("username=%s,password=%s,uid=1000,gid=1000,file_mode=0740,dir_mode=0700", user, password))
	err := shell.Run()
	if err != nil {
		if !tui_mode {
			pretty_exit("Can't mount directory, already mounted/no such directory/error with connection")
		}
		return false
	}
	if !tui_mode {
		fmt.Println("Succesful mount :)")
	}
	return true
}

func dismount_samba(host string, tui_mode bool) bool {
	shell := exec.Command("umount", host)
	err := shell.Run()
	if err != nil {
		if !tui_mode {
			pretty_exit("Can't dismount directory, wrong directory/not mounted")
		}
		return false
	}
	if !tui_mode {
		fmt.Println("Succesful dismount :)")
	}
	return true
}

func is_mounted(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		pretty_exit("Can't find directory")
	}

	if !fileInfo.IsDir() {
		pretty_exit("Mount location is not a directory")
	}

	stuff, err := os.ReadDir(path)

	if err != nil {
		pretty_exit("Directory can't be found")
	}

	return len(stuff) != 0
}

func is_root() bool {
	user, err := user.Current()
	if err != nil {
		pretty_exit("Run this command as sudo/root")
		return false
	}
	return user.Username == "root"
}

func main() {
	if !is_root() {
		fmt.Println("You need to run this app as a root user !!")
		os.Exit(0)
	}
	if len(os.Args) > 1 {
		args := os.Args[1:]
		shell_mode(args)
	} else {
		tui_mode()
	}
}
