/*
 * Author: Alexandre Havrileck (Oxyno-zeta)
 * Date: 21/04/2019
 * Licence: See Readme
 */
package rules

import (
	"reflect"
	"testing"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"
)

func TestCalculateTags(t *testing.T) {
	type args struct {
		actualTags         []*tags.Tag
		availableTagValues map[string]interface{}
		rules              []*Rule
	}
	tests := []struct {
		name    string
		args    args
		want    *tags.TagDelta
		wantErr bool
		err     error
	}{
		{
			"no rules",
			args{
				actualTags:         []*tags.Tag{},
				availableTagValues: nil,
				rules:              []*Rule{},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"available tags is nil",
			args{
				actualTags:         []*tags.Tag{},
				availableTagValues: nil,
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Query:  "test",
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"tag not present for delete rule",
			args{
				actualTags: []*tags.Tag{},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionDelete,
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"tag present for delete rule",
			args{
				actualTags: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "value-test",
					},
				},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionDelete,
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList: []*tags.Tag{},
				DeleteList: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "value-test",
					},
				},
			},
			false,
			nil,
		},
		{
			"No query result in add rule",
			args{
				actualTags: []*tags.Tag{},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Query:  "fake",
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"query result in add rule",
			args{
				actualTags: []*tags.Tag{},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Query:  "key1",
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "value1",
					},
				},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"direct value in add rule",
			args{
				actualTags: []*tags.Tag{},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Value:  "value-test",
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "value-test",
					},
				},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"skip add tag when tag and value already exists for add rule",
			args{
				actualTags: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "value-test",
					},
				},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Value:  "value-test",
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"tag value must be updated when value is different for add rule",
			args{
				actualTags: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "wrong-value",
					},
				},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Value:  "value-test",
						Tag:    "tag-test",
					},
				},
			},
			&tags.TagDelta{
				AddList: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "value-test",
					},
				},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"rule should be ignored if conditions are not valid because tag should be equal",
			args{
				actualTags: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "wrong-value",
					},
				},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Value:  "value-test",
						Tag:    "tag-test",
						When: []*Condition{
							&Condition{
								Condition: "key1",
								Operator:  ConditionOperatorEqual,
								Value:     "Value-not-valid",
							},
						},
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"rule should be ignored if conditions are not valid because tag is not present",
			args{
				actualTags: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "wrong-value",
					},
				},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Value:  "value-test",
						Tag:    "tag-test",
						When: []*Condition{
							&Condition{
								Condition: "key",
								Operator:  ConditionOperatorEqual,
								Value:     "Value-not-valid",
							},
						},
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
		{
			"rule should be ignored if conditions are not valid because tag should not be equal",
			args{
				actualTags: []*tags.Tag{
					&tags.Tag{
						Key:   "tag-test",
						Value: "wrong-value",
					},
				},
				availableTagValues: map[string]interface{}{
					"key1": "value1",
				},
				rules: []*Rule{
					&Rule{
						Action: RuleActionAdd,
						Value:  "value-test",
						Tag:    "tag-test",
						When: []*Condition{
							&Condition{
								Condition: "key1",
								Operator:  ConditionOperatorNotEqual,
								Value:     "value1",
							},
						},
					},
				},
			},
			&tags.TagDelta{
				AddList:    []*tags.Tag{},
				DeleteList: []*tags.Tag{},
			},
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateTags(tt.args.actualTags, tt.args.availableTagValues, tt.args.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.err != err {
				t.Errorf("CalculateTags() error '%v', expected err '%v'", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
