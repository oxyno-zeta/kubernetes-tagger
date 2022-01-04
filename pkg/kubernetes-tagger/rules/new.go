package rules

import (
	"errors"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
)

// ErrRuleEmptyWhenCondition Error when "when condition" is empty.
var ErrRuleEmptyWhenCondition = errors.New("when condition mustn't be empty")

// ErrRuleConditionOperatorNotSupported Rule condition operator not supported.
var ErrRuleConditionOperatorNotSupported = errors.New("condition operator not supported")

// ErrRuleActionNotSupported Rule action not supported.
var ErrRuleActionNotSupported = errors.New("rule action not supported")

// ErrRuleEmptyTag Err Rule Empty Tag.
var ErrRuleEmptyTag = errors.New("tag rule mustn't be empty")

// ErrRuleQueryAndValueEmptyForAddCase Err Rule Query and Value empty for Add case.
var ErrRuleQueryAndValueEmptyForAddCase = errors.New("query and value mustn't be empty for add case")

// ErrRuleQueryAndValuePopulatedForAddCase Err Rule query and value populated for Add case.
var ErrRuleQueryAndValuePopulatedForAddCase = errors.New("query and value cannot be populated at the same time in add case")

// New Create rules array from ruleConfig with validation.
func New(ruleConfigs []*config.RuleConfig) ([]*Rule, error) {
	rules := make([]*Rule, 0)
	if ruleConfigs == nil {
		return rules, nil
	}

	for _, ruleConfig := range ruleConfigs {
		rule, err := newFromRuleConfig(ruleConfig)
		// Check error
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func newFromRuleConfig(ruleConfig *config.RuleConfig) (*Rule, error) {
	if ruleConfig == nil {
		return nil, nil // nolint: nilnil // No need
	}

	// Create rule
	rule := &Rule{
		Tag:   ruleConfig.Tag,
		Query: ruleConfig.Query,
		Value: ruleConfig.Value,
	}

	switch ruleConfig.Action {
	case string(RuleActionAdd):
		rule.Action = RuleActionAdd
	case string(RuleActionDelete):
		rule.Action = RuleActionDelete
	default:
		return nil, ErrRuleActionNotSupported
	}
	// Check if rule is valid
	if rule.Tag == "" {
		return nil, ErrRuleEmptyTag
	}

	// Check add case
	if rule.Action == RuleActionAdd && rule.Query == "" && rule.Value == "" {
		return nil, ErrRuleQueryAndValueEmptyForAddCase
	}

	// Check that in add case we haven't query and value at the same time
	if rule.Action == RuleActionAdd && rule.Query != "" && rule.Value != "" {
		return nil, ErrRuleQueryAndValuePopulatedForAddCase
	}

	// Manage conditions
	conditions := make([]*Condition, 0)

	for _, conditionConfig := range ruleConfig.When {
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

	rule.When = conditions

	return rule, nil
}
