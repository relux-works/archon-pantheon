package kimi

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/skill-agents-management/internal/nativeargs"
	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

const promptPlaceholder = "<prompt>"

const (
	modelFlag       = "--model"
	promptFlag      = "--prompt"
	yoloFlag        = "--yolo"
	autoFlag        = "--auto"
	shortYoloFlag   = "-y"
	yesYoloFlag     = "--yes"
	autoApproveFlag = "--auto-approve"
)

// ErrPromptInvalid reports a prompt that the CLI cannot accept as text.
var ErrPromptInvalid = errors.New("kimi: prompt is empty or cannot be represented by the CLI")

// Args is the single Kimi CLI argument construction site.
func Args(req agentic.LaunchRequest, mode agentic.LaunchMode) ([]string, error) {
	model := req.Model.ID
	if strings.TrimSpace(model) == "" {
		return nil, agentic.ErrModelMissing
	}

	switch mode {
	case agentic.LaunchModeExec, agentic.LaunchModeDryRun:
		if !req.PermissionMode.IsZero() {
			return nil, fmt.Errorf("kimi: %w: permission mode is only available for interactive launches", agentic.ErrPermissionModeNotInteractive)
		}
		if len(req.NativeArgs) != 0 {
			return nil, fmt.Errorf("kimi: %w: native arguments are only forwarded for interactive launches", agentic.ErrNativeArgsNotInteractive)
		}
		if prompt, err := promptArgument(req, mode); err != nil {
			return nil, err
		} else {
			return []string{modelFlag, model, promptFlag, prompt}, nil
		}
	case agentic.LaunchModeInteractive:
		if len(req.Prompt) != 0 || strings.TrimSpace(req.PromptPath) != "" || req.Goal != nil || req.Budget != nil || strings.TrimSpace(req.ServiceTier) != "" || strings.TrimSpace(req.Effort) != "" {
			return nil, fmt.Errorf("kimi: %w: interactive mode carries no prompt, goal, budget, service tier, or effort", agentic.ErrParameterNotInteractive)
		}
		var permission agentic.PermissionMode
		if permissionValue, err := req.PermissionMode.Resolve(); err != nil {
			return nil, fmt.Errorf("kimi: %w", err)
		} else {
			permission = permissionValue
		}
		if err := validateInteractiveArgs(req.NativeArgs, permission); err != nil {
			return nil, err
		}
		args := []string{modelFlag, model}
		if permission == agentic.PermissionModeYolo {
			var mapping agentic.PermissionMapping
			if mappingValue, err := permissionMapping(req.ToolRelease, permission); err != nil {
				return nil, err
			} else {
				mapping = mappingValue
			}
			args = append(args, mapping.Flag)
		}
		return append(args, req.NativeArgs...), nil
	default:
		return nil, fmt.Errorf("kimi: unsupported launch mode %s", mode)
	}
}

func promptArgument(req agentic.LaunchRequest, mode agentic.LaunchMode) (string, error) {
	if mode == agentic.LaunchModeDryRun {
		return promptPlaceholder, nil
	}
	var prompt []byte
	if path := strings.TrimSpace(req.PromptPath); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("kimi: reading the prompt file: %w", err)
		}
		prompt = data
	} else {
		prompt = req.Prompt
	}
	if len(prompt) == 0 || !utf8.Valid(prompt) || strings.ContainsRune(string(prompt), '\x00') || strings.TrimSpace(string(prompt)) == "" {
		return "", ErrPromptInvalid
	}
	return string(prompt), nil
}

func validateInteractiveArgs(args []string, permission agentic.PermissionMode) error {
	var sessionSelector bool
	var continueSelector bool
	for _, index := range nativeargs.FlagIndexes(args) {
		token := args[index]
		name, _, _ := nativeargs.SplitFlagValue(token)
		for _, shortName := range shortOptionNames(token) {
			if shortName == shortYoloFlag || shortName == "-m" || shortName == "-p" {
				name = shortName
				break
			}
		}
		shortSession := (strings.HasPrefix(token, "-S") && !strings.HasPrefix(token, "--") && token != "-S") ||
			(strings.HasPrefix(token, "-r") && !strings.HasPrefix(token, "--") && token != "-r")
		if shortSession {
			if sessionSelector {
				return resumeInvalid("conflicting session selectors")
			}
			sessionSelector = true
			if strings.TrimSpace(token[2:]) == "" {
				return resumeInvalid("session selector requires an id")
			}
			continue
		}
		switch name {
		case yoloFlag, shortYoloFlag, yesYoloFlag, autoApproveFlag:
			if permission == agentic.PermissionModeYolo {
				return fmt.Errorf("kimi: %w: native arguments already carry a yolo flag", agentic.ErrPermissionModeDuplicate)
			}
			return fmt.Errorf("kimi: %w: select yolo through PermissionMode", agentic.ErrNativePolicyConflict)
		case autoFlag:
			return fmt.Errorf("kimi: %w: --auto is not represented by the shared permission-mode API", agentic.ErrPermissionModeUnsupported)
		case promptFlag, "-p":
			return fmt.Errorf("kimi: %w: use exec mode for --prompt", agentic.ErrNativePolicyConflict)
		case modelFlag, "-m":
			return fmt.Errorf("kimi: %w: select the model through the launch request", agentic.ErrNativePolicyConflict)
		case "--session", "--resume", "-S", "-r":
			if sessionSelector {
				return resumeInvalid("conflicting session selectors")
			}
			sessionSelector = true
			if id, err := optionalSessionID(args, index, name); err != nil {
				return err
			} else if strings.TrimSpace(id) == "" {
				return resumeInvalid("session selector requires an id")
			}
		case "--continue", "-c", "-C":
			if continueSelector {
				return resumeInvalid("conflicting session selectors")
			}
			continueSelector = true
			if strings.Contains(token, "=") {
				return resumeInvalid("continue selector does not take an id")
			}
		}
	}
	if sessionSelector && continueSelector {
		return resumeInvalid("cannot combine continue and explicit session selectors")
	}
	return nil
}

// shortOptionNames expands the source-pinned Commander short-option grammar:
// boolean options may cluster, and the argument-taking -S, -r, -m, and -p
// consume the remainder of a cluster as their value.
func shortOptionNames(token string) []string {
	if len(token) < 3 || token[0] != '-' || token[1] == '-' {
		return nil
	}
	cluster, _, _ := strings.Cut(token[1:], "=")
	options := make([]string, 0, len(cluster))
	for i := 0; i < len(cluster); i++ {
		option := cluster[i]
		options = append(options, "-"+string(option))
		if shortOptionTakesValue(option) {
			break
		}
	}
	return options
}

func shortOptionTakesValue(option byte) bool {
	switch option {
	case 'S', 'r', 'm', 'p':
		return true
	default:
		return false
	}
}

func optionalSessionID(args []string, index int, name string) (string, error) {
	token := args[index]
	_, value, attached := nativeargs.SplitFlagValue(token)
	if attached {
		return value, nil
	}
	if (name == "-S" || name == "-r") && strings.HasPrefix(token, name) && token != name {
		return strings.TrimPrefix(token, name), nil
	}
	if index+1 >= len(args) || args[index+1] == "--" || strings.HasPrefix(args[index+1], "-") {
		return "", resumeInvalid("session selector requires an id")
	}
	return args[index+1], nil
}

func resumeInvalid(reason string) error {
	return &agentic.ResumeInvalidError{Reason: reason}
}
