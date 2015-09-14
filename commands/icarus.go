// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golangchallenge/gc6/mazelib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Defining the icarus command.
// This will be called as 'laybrinth icarus'
var icarusCmd = &cobra.Command{
	Use:     "icarus",
	Aliases: []string{"client"},
	Short:   "Start the laybrinth solver",
	Long: `Icarus wakes up to find himself in the middle of a Labyrinth.
  Due to the darkness of the Labyrinth he can only see his immediate cell and if
  there is a wall or not to the top, right, bottom and left. He takes one step
  and then can discover if his new cell has walls on each of the four sides.

  Icarus can connect to a Daedalus and solve many laybrinths at a time.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunIcarus()
	},
}

func init() {
	RootCmd.AddCommand(icarusCmd)
}

func RunIcarus() {
	// Run the solver as many times as the user desires.
	fmt.Println("Solving", viper.GetInt("times"), "times")
	for x := 0; x < viper.GetInt("times"); x++ {

		solveMaze()
	}

	// Once we have solved the maze the required times, tell daedalus we are done
	makeRequest("http://127.0.0.1:" + viper.GetString("port") + "/done")
}

// Make a call to the laybrinth server (daedalus) that icarus is ready to wake up
func awake() mazelib.Survey {
	contents, err := makeRequest("http://127.0.0.1:" + viper.GetString("port") + "/awake")
	if err != nil {
		fmt.Println(err)
	}
	r := ToReply(contents)
	return r.Survey
}

// Make a call to the laybrinth server (daedalus)
// to move Icarus a given direction
// Will be used heavily by solveMaze
func Move(direction string) (mazelib.Survey, error) {
	if direction == "left" || direction == "right" || direction == "up" || direction == "down" {

		contents, err := makeRequest("http://127.0.0.1:" + viper.GetString("port") + "/move/" + direction)
		if err != nil {
			return mazelib.Survey{}, err
		}

		rep := ToReply(contents)
		if rep.Victory == true {
			fmt.Println(rep.Message)
			// os.Exit(1)
			return rep.Survey, mazelib.ErrVictory
		} else {
			//don't return an error if no error occured
			return rep.Survey, nil // errors.New(rep.Message)
		}
	}

	return mazelib.Survey{}, errors.New("invalid direction")
}

// utility function to wrap making requests to the daedalus server
func makeRequest(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// Handling a JSON response and unmarshalling it into a reply struct
func ToReply(in []byte) mazelib.Reply {
	res := &mazelib.Reply{}
	json.Unmarshal(in, &res)
	return *res
}

func solveMaze() {
	survey := awake()

	breadcrumbs := mazelib.NewStack(100)
	Iterate(survey, breadcrumbs)
}

func Iterate(survey mazelib.Survey, breadcrumbs *mazelib.Stack) {
	var err error
	co := mazelib.Coordinate{0, 0}

	//data structure to keep track of already visited rooms
	visited := make(map[mazelib.Coordinate]bool)
	//start from the cnter (0,0) and set it to visited
	visited[co] = true
	for true {

		//If solution found
		if err == mazelib.ErrVictory {
			fmt.Printf("Solution with %d steps found\n", breadcrumbs.Len())
			return
		}
		if err != nil {
			fmt.Printf("An error happend: |%v|\n", err)
			return
		}
		move, moveErr := getValidMove(survey, visited, co)

		//a move forward is possible
		if moveErr == nil {
			survey, co, err = executeMove(move, breadcrumbs, visited, co)
		} else if undoObj, stackErr := breadcrumbs.Pop(); stackErr != mazelib.ErrEmptyStack {
			// back up one step and try again
			undo := ReverseMove(undoObj.(string))
			survey, err = Move(undo)
			co = co.TranformByMove(undo)
		} else {
			fmt.Println("No solution found.")
			return
		}

	}
}
//execute move take a possible move to execute,
//keep track of it in the breadcrumbs and update coordination
func executeMove(move string, breadcrumbs *mazelib.Stack, visited map[mazelib.Coordinate]bool, co mazelib.Coordinate) (mazelib.Survey, mazelib.Coordinate, error) {
	//add the reverse move to the breadcrumbs stack
	breadcrumbs.Push(move)
	//perform the move
	survey, err := Move(move)
	//mark the new room as visited
	co = co.TranformByMove(move)
	visited[co] = true
	return survey, co, err
}
//return a possible move toward not already visited room
func getValidMove(survey mazelib.Survey, visited map[mazelib.Coordinate]bool, co mazelib.Coordinate) (string, error) {
	if move := "down"; !survey.Bottom && !visited[co.TranformByMove(move)] {
		return move, nil
	}
	if move := "right"; !survey.Right && !visited[co.TranformByMove(move)] {
		return move, nil
	}
	if move := "up"; !survey.Top && !visited[co.TranformByMove(move)] {
		return move, nil
	}
	if move := "left"; !survey.Left && !visited[co.TranformByMove(move)] {
		return move, nil
	}
	return "", errors.New("No move is allowed")
}
//reverse a move to enable taking a step back
//when reverse walking the breadcrumbs 
func ReverseMove(move string) string {
	switch move {
	case "down":
		return "up"
	case "up":
		return "down"
	case "right":
		return "left"
	case "left":
		return "right"
	}
	return ""
}
