package rules

import (
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/oxyno-zeta/kubernetes-tagger/pkg/kubernetes-tagger/tags"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

// ErrCannotStringifyAvailableTagValues Cannot stringify available tag values.
var ErrCannotStringifyAvailableTagValues = errors.New("error cannot transform to string available tag values")

// CalculateTags Calculate tags delta to add/update or delete tags on resource.
func CalculateTags(actualTags []*tags.Tag, availableTagValues map[string]interface{}, rules []*Rule) (*tags.TagDelta, error) {
	logrus.Debug("Begin calculate tags from available values and rules")
	// Create GJSON result to filter tags
	jsonBytes, err := json.Marshal(availableTagValues)
	if err != nil {
		logrus.Debugf("Error: cannot stringify available tag values: %v", err)

		return nil, ErrCannotStringifyAvailableTagValues
	}

	jsonString := string(jsonBytes)
	gjsonResult := gjson.Parse(jsonString)

	// TODO for delete case, need to calculate all theoretical add and see if need to be remove from add list

	// Manage rules
	addList := make([]*tags.Tag, 0)
	deleteList := make([]*tags.Tag, 0)

	for _, rule := range rules {
		// Eval conditions
		whenResult := evalConditions(rule.When, gjsonResult)
		if !whenResult {
			continue
		}

		// Create tag
		tag := &tags.Tag{
			Key: rule.Tag,
		}

		if rule.Action == RuleActionDelete {
			// Delete case
			// In the delete case, no value is required
			// Value is the actual one in fact, if it exists
			// Filter to check if the value already exists on the resource
			filterResult, _ := funk.Filter(actualTags, func(actualTag *tags.Tag) bool {
				return actualTag.Key == tag.Key
			}).([]*tags.Tag)
			// Check if tag already exists and need to be removed
			if len(filterResult) != 0 {
				// Add actual tag value
				tag.Value = filterResult[0].Value
				// Add it to delete list
				deleteList = append(deleteList, tag)
			} else {
				logrus.Infof("Tag %s is already deleted -> skipping", tag.Key)
			}
		} else {
			// Add case

			// In Add Action, value is required

			// Check if we are in query case
			if rule.Query != "" {
				queryResult := gjsonResult.Get(rule.Query).String()
				if queryResult == "" {
					// Stop here, cannot get value
					logrus.Infof("Tag %s with query %s doesn't give any results -> skip it", rule.Tag, rule.Query)

					continue
				}
				tag.Value = queryResult
			} else {
				// Value directly case
				tag.Value = rule.Value
			}

			// Filter to test if value if necessary added / updated
			filterResult, _ := funk.Filter(actualTags, func(actualTag *tags.Tag) bool {
				return actualTag.Key == tag.Key && actualTag.Value == tag.Value
			}).([]*tags.Tag)

			// Check if tag already exists and need to be added / updated
			if len(filterResult) == 0 {
				addList = append(addList, tag)
			} else {
				logrus.Infof("Tag %s with value \"%s\" is already present -> skipping", tag.Key, tag.Value)
			}
		}
	}

	delta := &tags.TagDelta{AddList: addList, DeleteList: deleteList}

	return delta, nil
}

func evalConditions(conditions []*Condition, gjsonResult gjson.Result) bool {
	result := true

	for _, condition := range conditions {
		queryResult := gjsonResult.Get(condition.Condition).String()
		if condition.Operator == ConditionOperatorEqual {
			result = result && queryResult == condition.Value
		} else {
			result = result && queryResult != condition.Value
		}
		// Quit when false arrive => ASAP
		if !result {
			return result
		}
	}

	return result
}
