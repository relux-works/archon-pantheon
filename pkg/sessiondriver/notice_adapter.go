package sessiondriver

// NoticeSelection is opaque selection metadata for notice-based adapters.
type NoticeSelection struct {
	Vendor    string `json:"vendor"`
	Model     string `json:"model"`
	Reasoning string `json:"reasoning"`
}

// Notice contains a reference, never mailbox content or an arbitrary prompt.
type Notice struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
}

// InputState is a conservative observation of adapter input readiness.
type InputState string

const (
	InputIdle         InputState = "idle"
	InputBusy         InputState = "busy"
	InputDraft        InputState = "draft"
	InputApproval     InputState = "approval"
	InputUnknown      InputState = "unknown"
	InputSecurityStop InputState = "security-stop"
)

// NoticeAdapter encodes notice deliveries and checks output acknowledgements.
type NoticeAdapter interface {
	Name() string
	InputState() InputState
	Encode(Notice, NoticeSelection) ([]byte, error)
	Applied([]byte, Notice, NoticeSelection) (bool, error)
}
