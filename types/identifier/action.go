package identifier

import "fmt"

// NewActionID returns a validated action ID from an int32. If
// the validation fails, the returned value signals an unsupported
// action and an error is returned along with it.
func NewActionID(id int32) (ActionID, error) {
	actionID := ActionID(id)
	if err := actionID.Validate(); err != nil {
		return ACTION_UNSUPPORTED, err
	}

	return actionID, nil
}

// Validate returns an error if the ID is not valid.
func (id ActionID) Validate() error {
	if id == ACTION_UNSUPPORTED {
		return fmt.Errorf("action id is not supported: %s", id.String())
	}
	if _, found := ActionID_name[int32(id)]; !found {
		return fmt.Errorf("action id is unknown: %d", int32(id))
	}

	return nil
}
