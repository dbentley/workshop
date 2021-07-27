package state

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/tilt-dev/workshop/wslocal/apihelpers"
)

type State struct {
	StateName         string
	StateFriendlyName string
	Description       string
	Substeps          []Substep
	Buttons           []Button
	Unknown           map[string]interface{} `json:"-"`
	StepNum           int
	TotalSteps        int
	WorkshopStarted   bool
	WorkshopDone      bool
}

type Substep struct {
	Desc        string
	Instruction string
	Output      string
	Expected    string
	Done        bool
}

func NewSubstep(desc string, instruction string, done bool) Substep {
	return Substep{Desc: desc, Instruction: instruction, Done: done}
}

type Button struct {
	Name string
	Text string
}

func NewButton(name string, text string) Button {
	return Button{Name: name, Text: text}
}

const filename = "workshop_state.json"

const initState = "init"

func CopyState(st State) State {
	return State{
		StateName:         st.StateName,
		StateFriendlyName: st.StateFriendlyName,
		Description:       st.Description,
		Substeps:          append([]Substep(nil), st.Substeps...),
		Buttons:           append([]Button(nil), st.Buttons...),
		Unknown:           st.Unknown,
		StepNum:           st.StepNum,
		TotalSteps:        st.TotalSteps,
		WorkshopStarted:   st.WorkshopStarted,
		WorkshopDone:      st.WorkshopDone,
	}
}

func AdvanceIfSubstepsComplete(st State, newStateName string) State {
	for _, substep := range st.Substeps {
		if !substep.Done {
			return st
		}
	}

	st.StateName = newStateName
	return st
}

func LoadState() (State, error) {
	var result State
	f, err := os.Open(filename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			if false {
				log.Printf("Initializing Workshop state")
			}
			return State{Unknown: make(map[string]interface{})}, nil
		}
		return result, err
	}

	defer f.Close()

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bs, &result)
	if err != nil {
		return State{}, err
	}
	return result, nil
}

func SaveState(ctx context.Context, api *apihelpers.APIHelper, pre, post State, first bool) (bool, error) {
	preBytes, err := json.MarshalIndent(pre, "", "  ")
	if err != nil {
		return false, err
	}

	postBytes, err := json.MarshalIndent(post, "", "  ")
	if err != nil {
		return false, err
	}

	if bytes.Compare(preBytes, postBytes) == 0 && !first {
		// nothing changed
		return false, nil
	}

	if post.StateName != pre.StateName {
		for _, button := range pre.Buttons {
			if err := api.DeleteButton(ctx, button.Name); err != nil {
				return false, err
			}
		}
	}
	for _, button := range post.Buttons {
		// TODO(dbentley): only create buttons if different than before  (so we don't overwrite clicks)
		if err := api.CreateButton(ctx, button.Name, button.Text); err != nil {
			return false, err
		}
	}

	return true, os.WriteFile(filename, postBytes, 0666)
}
