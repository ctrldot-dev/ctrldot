package policy

// PolicySet represents a parsed policy set
type PolicySet struct {
	PolicySet string `yaml:"policy_set"`
	Version   string `yaml:"version"`
	AppliesTo struct {
		Namespace string `yaml:"namespace"`
	} `yaml:"applies_to"`
	Rules []Rule `yaml:"rules"`
}

// Rule represents a policy rule
type Rule struct {
	ID      string   `yaml:"id"`
	When    When     `yaml:"when"`
	Require []Require `yaml:"require"`
	Effect  Effect   `yaml:"effect"`
}

// When represents rule matching conditions
type When struct {
	Op       []string `yaml:"op"`
	LinkType string   `yaml:"link_type"`
}

// Require represents a predicate requirement
type Require struct {
	Predicate string                 `yaml:"predicate"`
	Args      map[string]interface{} `yaml:"args"`
}

// Effect represents the effect of a rule violation
type Effect struct {
	Type    string `yaml:"deny"` // Can be "deny", "warn", or "info"
	Message string `yaml:"deny"` // The message for deny, warn, or info
}

// UnmarshalYAML custom unmarshaler for Effect to handle deny/warn/info
func (e *Effect) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var m map[string]string
	if err := unmarshal(&m); err != nil {
		return err
	}

	if deny, ok := m["deny"]; ok {
		e.Type = "deny"
		e.Message = deny
	} else if warn, ok := m["warn"]; ok {
		e.Type = "warn"
		e.Message = warn
	} else if info, ok := m["info"]; ok {
		e.Type = "info"
		e.Message = info
	}

	return nil
}
