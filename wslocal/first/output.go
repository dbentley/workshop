package first

import (
	"context"
	"fmt"

	"github.com/tilt-dev/workshop/wslocal/state"
)

func (m *Machine) PrintState(ctx context.Context, st state.State) error {
	fmt.Printf("\n\n")

	fmt.Printf("───────────┤ Step %d: %v ├───────────\n\n",
		st.StepNum, st.StateFriendlyName)

	fmt.Printf("%v\n\n", st.Description)

	fmt.Printf("To progress to step %d of %d:\n", st.StepNum + 1, st.TotalSteps)
	for _, substep := range st.Substeps {
		if err := m.PrintSubstep(ctx, substep); err != nil {
			return err
		}
	}

	return nil
}

func (m *Machine) PrintSubstep(ctx context.Context, substep state.Substep) error {
	doneStr := "☐"
	if substep.Done {
		doneStr = "✔︎"
	}

	if substep.Instruction == "" {
		fmt.Printf("  ╎ %v %-24s\n", doneStr, substep.Desc)
		return nil
	}
	if substep.Output == "" {
		fmt.Printf("  ╎ %v %-24s (%20s)\n", doneStr, substep.Desc, substep.Instruction)
		return nil
	}

	fmt.Printf("  ╎ %v %-24s (%20s) -> %q\n", doneStr, substep.Desc, substep.Instruction, substep.Output)
	return nil
}
