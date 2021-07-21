package apihelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

type Update interface{}

type UIButtonUpdate struct {
	Text string
}

type APIHelper struct {
}

func NewAPIHelper() (*APIHelper, error) {
	return &APIHelper{}, nil
}

func (h *APIHelper) GetUsername(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}

type JSONHelperUIButton struct {
	Status JSONHelperStatus
}

type JSONHelperStatus struct {
	LastClickedAt time.Time
}

func (h *APIHelper) HasBeenClicked(ctx context.Context, name string) (bool, error) {
	cmd := exec.Command("tilt", "get", "-o", "json", "uibutton", name)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	var r JSONHelperUIButton
	if err := json.Unmarshal(output, &r); err != nil {
		return false, err
	}

	return !r.Status.LastClickedAt.IsZero(), nil
}

type UIResourceStatus struct {
	RuntimeStatus string
	UpdateStatus  string
}

type JSONHelperUIResource struct {
	Status UIResourceStatus
}

func (h *APIHelper) GetUIResource(ctx context.Context, name string) (UIResourceStatus, error) {
	cmd := exec.Command("tilt", "get", "-o", "json", "uiresource", name)
	output, err := cmd.Output()
	if err != nil {
		return UIResourceStatus{}, err
	}

	var r JSONHelperUIResource
	if err := json.Unmarshal(output, &r); err != nil {
		return UIResourceStatus{}, err
	}

	return r.Status, nil
}

func (h *APIHelper) CreateButton(ctx context.Context, name string, text string) error {
	cmd := exec.Command("tilt", "apply", "-f", "/dev/stdin")

	yaml := fmt.Sprintf(buttonYAMLF, name, text)
	// set stdin
	cmd.Stdin = bytes.NewBufferString(yaml)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if false {
			log.Printf("error running command: %s", output)
		}
	}
	return err
}

func (h *APIHelper) DeleteButton(ctx context.Context, name string) error {
	cmd := exec.Command("tilt", "delete", "uibutton", name)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if false {
			log.Printf("error running command: %s", output)
		}
	}
	return err
}

const buttonYAMLF = `
apiVersion: tilt.dev/v1alpha1
kind: UIButton
metadata:
  name: %s
spec:
  location:
    componentID: "workshop"
    componentType: "Resource"
  text: "%s"
`
