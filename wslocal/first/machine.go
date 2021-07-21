package first

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/tilt-dev/workshop/wslocal/apihelpers"
	"github.com/tilt-dev/workshop/wslocal/state"
)

type Machine struct {
	api *apihelpers.APIHelper
}

func NewMachine(api *apihelpers.APIHelper) (*Machine, error) {
	return &Machine{api: api}, nil
}

const initState = ""
const waitForReadyState = "WaitForReady"
const waitForUpdateState = "WaitForUpdate"
const doneState = "Done"
const deleteState = "Delete"

func (m *Machine) Advance(ctx context.Context, pre state.State) (state.State, error) {
	if pre.StateName == initState {
		return m.handleInitState(ctx, pre)
	}

	if pre.StateName == waitForReadyState {
		return m.handleWaitForReadyState(ctx, pre)
	}

	if pre.StateName == waitForUpdateState {
		return m.handleWaitForUpdateState(ctx, pre)
	}

	if pre.StateName == doneState {
		return m.handleDoneState(ctx, pre)
	}

	if pre.StateName == deleteState {
		return m.handleDeleteState(ctx, pre)
	}

	return state.State{}, fmt.Errorf("not yet implemented")
}

func (m *Machine) handleInitState(ctx context.Context, pre state.State) (state.State, error) {

	pre.StepNum = 0
	pre.TotalSteps = 3
	pre.Buttons = []state.Button{state.NewButton("workshop-init-advance", "Next Step")}
	pre.Description = `Welcome to the Tilt Workshop!

Press the "Press Here" button above and we'll get going.
`

	pressed, _ := m.api.HasBeenClicked(ctx, "workshop-init-advance") // ignore error cause it might not exist yet

	substeps := []state.Substep{
		state.NewSubstep("Click Next", "", pressed),
	}
	pre.Substeps = substeps

	pre = state.AdvanceIfSubstepsComplete(pre, waitForReadyState)

	return pre, nil
}

var resources = []string{"frontend", "muxer", "glitch", "red", "storage", "rectangler", "recompile-muxer"}

func (m *Machine) handleWaitForReadyState(ctx context.Context, pre state.State) (state.State, error) {
	pre.WorkshopStarted = true
	pre.StepNum = 1
	pre.Description = `Get Pixeltilt up and running.
(This should just work, but if you have issues talking to docker or k8s you could have issues)

`
	substeps := m.checkPixeltiltReady(ctx)

	pre.Buttons = []state.Button{state.NewButton("workshop-ready-advance", "Next step")}
	pressed, _ := m.api.HasBeenClicked(ctx, "workshop-ready-advance") // ignore error cause it might not exist yet

	substeps = append(substeps, state.NewSubstep("Click Next", "", pressed))
	pre.Substeps = substeps
	pre = state.AdvanceIfSubstepsComplete(pre, waitForUpdateState)

	return pre, nil
}

func (m *Machine) handleWaitForUpdateState(ctx context.Context, pre state.State) (state.State, error) {

	pre.Buttons = nil
	pre.Substeps = nil
	pre.StepNum = 2
	pre.Description = `Tilt handles updating the runnign servers for you.

To see this, change some Go code and see it running. Let's change the author tag on the muxer's html.

In the file muxer/main.go, change "Tilt Team" to "workshop".
`

	pre.Buttons = []state.Button{state.NewButton("workshop-first-update-advance", "Next step")}
	pressed, _ := m.api.HasBeenClicked(ctx, "workshop-first-update-advance") // ignore error cause it might not exist yet

	substeps := m.checkPixeltiltEdited(ctx)
	substeps = append(substeps, state.NewSubstep("Click Next", "", pressed))
	pre.Substeps = substeps
	pre = state.AdvanceIfSubstepsComplete(pre, doneState)

	return pre, nil
}

func (m *Machine) handleDoneState(ctx context.Context, pre state.State) (state.State, error) {
	pre.Buttons = []state.Button{state.NewButton("workshop-done-advance", "Exit Workshop")}
	pre.Substeps = nil
	pre.StepNum = 3
	pre.Description = `Congrats! You're done with the workshop for now.
`
	pressed, _ := m.api.HasBeenClicked(ctx, "workshop-done-advance")

	pre.Substeps = []state.Substep{state.NewSubstep("Click Exit Workshop", "", pressed)}
	pre = state.AdvanceIfSubstepsComplete(pre, deleteState)

	return pre, nil
}

func (m *Machine) handleDeleteState(ctx context.Context, pre state.State) (state.State, error) {
	pre.WorkshopDone = true
	return pre, nil
}

func (m *Machine) checkPixeltiltEdited(ctx context.Context) []state.Substep {
	var r []state.Substep

	for _, name := range resources {
		substep := m.checkResourceReady(ctx, name)
		if !substep.Done {
			// only mention things that aren't ready. Will this be too noisy?
			r = append(r, substep)
		}
	}

	r = append(r, m.checkPixeltiltAuthor(ctx))
	return r
}

func (m *Machine) checkPixeltiltAuthor(ctx context.Context) state.Substep {
	var r state.Substep
	r.Desc = `author set to "workshop"`
	r.Instruction = "curl http://localhost:8080 | grep author"
	resp, err := m.getMuxer(ctx)
	if err != nil {
		r.Output = fmt.Sprintf("Error: %v", err)
		return r
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.Output = fmt.Sprintf("Error reading http: %v", err)
		return r
	}

	match := regexp.MustCompile(`content=".*"`).Find(body)

	if string(match) == expectedAuthor {
		r.Done = true
		return r
	}

	r.Output = string(match)
	r.Expected = `content="workshop"`

	return r

}

const expectedAuthor = `content="workshop"`

func (m *Machine) checkPixeltiltReady(ctx context.Context) []state.Substep {
	var r []state.Substep

	for _, name := range resources {
		r = append(r, m.checkResourceReady(ctx, name))
	}

	var muxerSubstep state.Substep
	muxerSubstep.Desc = "muxer ready"
	muxerSubstep.Instruction = "curl http://localhost:8080"
	resp, err := m.getMuxer(ctx)
	if err != nil {
		muxerSubstep.Output = fmt.Sprintf("Error: %v", err)
	} else {
		muxerSubstep.Done = true
		resp.Body.Close()
	}

	r = append(r, muxerSubstep)
	return r
}

func (m *Machine) checkResourceReady(ctx context.Context, name string) state.Substep {
	var r state.Substep
	r.Desc = fmt.Sprintf("resource %s", name)
	r.Instruction = fmt.Sprintf("tilt get uiresource -o json %v", name)

	status, err := m.api.GetUIResource(ctx, name)
	if err != nil {
		r.Output = fmt.Sprintf("Resource %q threw an error getting info: %q", name, err)
		return r
	}

	if status.RuntimeStatus != "ok" && status.RuntimeStatus != "not_applicable" {
		r.Output = fmt.Sprintf("RuntimeStatus: %q (should be \"ok\" or \"not_applicable\")", status.RuntimeStatus)
		return r
	}
	if status.UpdateStatus != "ok" && status.UpdateStatus != "not_applicable" {
		r.Output = fmt.Sprintf("UpdateStatus: %q (should be \"ok\" or \"not_applicable\")", status.UpdateStatus)
		return r
	}

	r.Done = true
	return r
}

func (m *Machine) getMuxer(ctx context.Context) (*http.Response, error) {
	cl := http.Client{Timeout: 1 * time.Second}
	return cl.Get("http://localhost:8080")
}
