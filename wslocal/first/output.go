package first

import (
	"context"
	"fmt"

	"github.com/tilt-dev/workshop/wslocal/state"
)

func (m *Machine) PrintState(ctx context.Context, st state.State) error {
	fmt.Printf("\n\n")

	fmt.Printf(".... %v (Step %d of %d) ....\n",
		st.StateFriendlyName, st.StepNum, st.TotalSteps)

	fmt.Printf("%v\n", st.Description)

	fmt.Printf("Complete the following steps to advance:\n")
	for _, substep := range st.Substeps {
		if err := m.PrintSubstep(ctx, substep); err != nil {
			return err
		}
	}

	fmt.Printf("^^^^ ^^^^\n")

	return nil
}

func (m *Machine) PrintSubstep(ctx context.Context, substep state.Substep) error {
	doneStr := "X  "
	if substep.Done {
		doneStr = "Ch "
	}

	if substep.Instruction == "" {
		fmt.Printf("  %v - %-24s\n", doneStr, substep.Desc)
		return nil
	}
	if substep.Output == "" {
		fmt.Printf("  %v - %-24s (%20s)\n", doneStr, substep.Desc, substep.Instruction)
		return nil
	}

	fmt.Printf("  %v - %-24s (%20s) -> %q\n", doneStr, substep.Desc, substep.Instruction, substep.Output)
	return nil
}
