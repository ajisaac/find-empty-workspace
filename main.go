package main

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"strconv"
)

type Workspace struct {
	Num     int
	Name    string
	Visible bool
	Focused bool
	Output  string
	Urgent  bool
}

/*
When we hit a button we take the currently selected program and move it
to the next available empty workspace.

do nothing if
	all the workspaces are taken

*/
func main() {

	wses, err := getWorkspaces()
	if err != nil {
		log.Fatal(err.Error())
	}

	fWS, err := getfocusedWorkspace(wses)
	if err != nil {
		log.Fatal(err.Error())
	}

	naWS, err := getNextAvailableWorkspace(wses, fWS)
	if err != nil {
		log.Fatal(err.Error())
	}

	moveWindowToWorkspace(naWS)

	moveToWorkspace(naWS)

}

func moveToWorkspace(ws int) {
	cmd := exec.Command("i3-msg", "workspace", strconv.Itoa(ws))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err.Error())
	}

	// run and decode response
	if err := cmd.Start(); err != nil {
		log.Fatal(err.Error())
	}

	p := make([]byte, 1000)
	stdout.Read(p)
	log.Printf("%s", p)
}

// Get an array of the workspace objects
func getWorkspaces() ([]Workspace, error) {
	// prepare command
	cmd := exec.Command("i3-msg", "-t", "get_workspaces")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err.Error())
	}

	// run and decode response
	if err := cmd.Start(); err != nil {
		log.Fatal(err.Error())
	}

	workspaces := make([]Workspace, 0)

	if err := json.NewDecoder(stdout).Decode(&workspaces); err != nil {
		log.Fatal(err)
	}

	if len(workspaces) == 0 {
		return workspaces, errors.New("no workspaces available")
	}
	return workspaces, nil
}

// take the currently focused container and move it to ws
func moveWindowToWorkspace(ws int) {
	cmd := exec.Command("i3-msg", "move", "container", "to", "workspace", strconv.Itoa(ws))
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err.Error())
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err.Error())
	}

	p := make([]byte, 1000)
	stdout.Read(p)
	log.Printf("%s", p)
}

// returns the workspace we're focused on or error if there is no workspace
func getfocusedWorkspace(wses []Workspace) (Workspace, error) {
	// the number of the focused workspace
	var focused Workspace
	foundWs := false
	for _, w := range wses {
		if w.Focused {
			focused = w
			foundWs = true
			break
		}
	}
	if foundWs {
		return focused, nil
	}
	return focused, errors.New("found no focused workspace")

}

// for the workspaces, return the number of the next available
func getNextAvailableWorkspace(wses []Workspace, focused Workspace) (int, error) {

	// these workspaces are taken
	unavailable := make([]int, 0, 10)
	for _, ws := range wses {
		unavailable = append(unavailable, ws.Num)
	}

	// find first and second choices of WS
	nextLowestAvailable := -1  // the next lowest num available after focused WS
	nextHighestAvailable := -1 // the next highest num below focused WS
	// for 1 - 10
outer:
	for i := 1; i <= 10; i++ {
		//is this number available
		for _, uv := range unavailable {
			if i == uv {
				//i is unavailable
				continue outer
			}
		}
		if i == focused.Num {
			continue
		}

		// assuming we made it here
		if i < focused.Num {
			nextLowestAvailable = i
		} else if i > focused.Num {
			nextHighestAvailable = i
			break outer
		}

	}

	if nextHighestAvailable > -1 {
		return nextHighestAvailable, nil
	}
	if nextLowestAvailable > -1 {
		return nextLowestAvailable, nil
	}
	return -1, errors.New("no workspaces available")

}
