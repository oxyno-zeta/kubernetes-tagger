package rules

import (
	"errors"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
)

// ErrRuleEmptyWhenCondition Error when "when condition" is empty
var ErrRuleEmptyWhenCondition = errors.New("when condition mustn't be empty")

// ErrRuleConditionOperatorNotSupported Rule condition operator not supported
var ErrRuleConditionOperatorNotSupported = errors.New("condition operator not supported")

// ErrRuleActionNotSupported Rule action not supported.
var ErrRuleActionNotSupported = errors.New("rule action not supported")

// New Create rules array from ruleConfig with validation
func New(ruleConfigs []*config.RuleConfig) ([]*Rule, error) {
	rules := make([]*Rule, 0)
	for i := 0; i < len(ruleConfigs); i++ {
		ruleConfig := ruleConfigs[i]
		rule, err := newFromRuleConfig(ruleConfig)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func newFromRuleConfig(ruleConfig *config.RuleConfig) (*Rule, error) {
	conditions := make([]*Condition, 0)
	for i := 0; i < len(ruleConfig.When); i++ {
		conditionConfig := ruleConfig.When[i]
		if conditionConfig.Condition == "" {
			return nil, ErrRuleEmptyWhenCondition
		}

		condition := &Condition{
			Condition: conditionConfig.Condition,
			Value:     conditionConfig.Value,
		}

		switch conditionConfig.Operator {
		case string(ConditionOperatorEqual):
			condition.Operator = ConditionOperatorEqual
		case string(ConditionOperatorNotEqual):
			condition.Operator = ConditionOperatorNotEqual
		default:
			return nil, ErrRuleConditionOperatorNotSupported
		}
		conditions = append(conditions, condition)
	}
	rule := &Rule{
		Tag:   ruleConfig.Tag,
		Query: ruleConfig.Query,
		Value: ruleConfig.Value,
		When:  conditions,
	}
	switch ruleConfig.Action {
	case string(RuleActionAdd):
		rule.Action = RuleActionAdd
	case string(RuleActionDelete):
		rule.Action = RuleActionDelete
	default:
		return nil, ErrRuleActionNotSupported
	}
	return rule, nil
}
