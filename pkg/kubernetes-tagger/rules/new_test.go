package rules

import (
	"reflect"
	"testing"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/config"
)

func TestNew(t *testing.T) {
	type args struct {
		ruleConfigs []*config.RuleConfig
	}
	tests := []struct {
		name    string
		args    args
		want    []*Rule
		wantErr bool
	}{
		{"NilList", args{ruleConfigs: nil}, []*Rule{}, false},
		{"EmptyList", args{ruleConfigs: []*config.RuleConfig{}}, []*Rule{}, false},
		{
			"List",
			args{[]*config.RuleConfig{
				&config.RuleConfig{
					Action: "add",
					Tag:    "test-tag",
					Query:  "query",
					When:   []*config.ConditionConfig{},
				},
			}},
			[]*Rule{
				&Rule{
					Action: RuleActionAdd,
					Tag:    "test-tag",
					Query:  "query",
					When:   []*Condition{},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ruleConfigs)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newFromRuleConfig(t *testing.T) {
	type args struct {
		ruleConfig *config.RuleConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *Rule
		wantErr bool
		err     error
	}{
		{
			"Nil",
			args{ruleConfig: nil},
			nil,
			false,
			nil,
		},
		{
			"Empty rule config",
			args{ruleConfig: &config.RuleConfig{}},
			nil,
			true,
			ErrRuleActionNotSupported,
		},
		{
			"Action not supported",
			args{ruleConfig: &config.RuleConfig{
				Action: "not-supported",
			}},
			nil,
			true,
			ErrRuleActionNotSupported,
		},
		{
			"No tag",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
			}},
			nil,
			true,
			ErrRuleEmptyTag,
		},
		{
			"No query and value in add case",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
			}},
			nil,
			true,
			ErrRuleQueryAndValueEmptyForAddCase,
		},
		{
			"No value and empty query in add case",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Query:  "",
			}},
			nil,
			true,
			ErrRuleQueryAndValueEmptyForAddCase,
		},
		{
			"No query and empty value in add case",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Value:  "",
			}},
			nil,
			true,
			ErrRuleQueryAndValueEmptyForAddCase,
		},
		{
			"Query and value populated at the same time in add case",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Value:  "value",
				Query:  "query",
			}},
			nil,
			true,
			ErrRuleQueryAndValuePopulatedForAddCase,
		},
		{
			"Valid valued rule without conditions",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Value:  "value",
				When:   []*config.ConditionConfig{},
			}},
			&Rule{
				Action: RuleActionAdd,
				Tag:    "test-tag",
				Value:  "value",
				When:   []*Condition{},
			},
			false,
			nil,
		},
		{
			"Valid query rule without conditions",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Query:  "query",
				When:   []*config.ConditionConfig{},
			}},
			&Rule{
				Action: RuleActionAdd,
				Tag:    "test-tag",
				Query:  "query",
				When:   []*Condition{},
			},
			false,
			nil,
		},
		{
			"When with empty condition",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Query:  "query",
				When: []*config.ConditionConfig{
					&config.ConditionConfig{Condition: ""},
				},
			}},
			nil,
			true,
			ErrRuleEmptyWhenCondition,
		},
		{
			"When with empty operator",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Query:  "query",
				When: []*config.ConditionConfig{
					&config.ConditionConfig{Condition: "condition", Operator: ""},
				},
			}},
			nil,
			true,
			ErrRuleConditionOperatorNotSupported,
		},
		{
			"rule is valid with when condition",
			args{ruleConfig: &config.RuleConfig{
				Action: "add",
				Tag:    "test-tag",
				Query:  "query",
				When: []*config.ConditionConfig{
					&config.ConditionConfig{
						Condition: "condition",
						Operator:  "Equal",
						Value:     "",
					},
				},
			}},
			&Rule{
				Action: RuleActionAdd,
				Tag:    "test-tag",
				Query:  "query",
				When: []*Condition{
					&Condition{
						Operator:  ConditionOperatorEqual,
						Condition: "condition",
						Value:     "",
					},
				},
			},
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newFromRuleConfig(tt.args.ruleConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFromRuleConfig() error = '%v', wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.err != err {
				t.Errorf("newFromRuleConfig() error '%v', expected err '%v'", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFromRuleConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
