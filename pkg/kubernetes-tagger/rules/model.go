package rules

// ActionType Action type
type ActionType string

// RuleActionAdd Rule Action Add
const RuleActionAdd = ActionType("add")

// RuleActionDelete Rule Action Delete
const RuleActionDelete = ActionType("delete")

// ConditionOperator Condition operator
type ConditionOperator string

// ConditionOperatorEqual Equal
const ConditionOperatorEqual = ConditionOperator("Equal")

// ConditionOperatorNotEqual Not Equal
const ConditionOperatorNotEqual = ConditionOperator("NotEqual")

// Rule rule
type Rule struct {
	Tag    string
	Query  string
	Value  string
	Action ActionType
	When   []*Condition
}

// Condition condition
type Condition struct {
	Condition string
	Value     string
	Operator  ConditionOperator
}
