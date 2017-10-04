//
// Copyright (c) 2017 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package broker

import (
	"fmt"
	"regexp"

	"encoding/json"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

type formItem struct {
	Key   string        `json:"key,omitempty"`
	Title string        `json:"title,omitempty"`
	Type  string        `json:"type,omitempty"`
	Items []interface{} `json:"items,omitempty"`
}

// SpecToService converts an apb Spec into a Service usable by the service
// catalog.
func SpecToService(spec *apb.Spec) Service {
	retSvc := Service{
		ID:          spec.ID,
		Name:        spec.FQName,
		Description: spec.Description,
		Tags:        make([]string, len(spec.Tags)),
		Bindable:    spec.Bindable,
		Plans:       toBrokerPlans(spec.Plans),
		Metadata:    spec.Metadata,
	}

	copy(retSvc.Tags, spec.Tags)
	return retSvc
}

func toBrokerPlans(apbPlans []apb.Plan) []Plan {
	brokerPlans := make([]Plan, len(apbPlans))
	i := 0
	for _, plan := range apbPlans {
		brokerPlans[i] = Plan{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Metadata:    extractBrokerPlanMetadata(plan),
			Free:        plan.Free,
			Bindable:    plan.Bindable,
			UpdatesTo:   plan.UpdatesTo,
			Schemas:     parametersToSchema(plan),
		}
		i++
	}
	return brokerPlans
}

func extractBrokerPlanMetadata(apbPlan apb.Plan) map[string]interface{} {
	metadata, err := initMetadataCopy(apbPlan.Metadata)

	if err != nil {
		return apbPlan.Metadata
	}

	instanceFormDefn := createFormDefinition(apbPlan.Parameters)
	bindingFormDefn := createFormDefinition(apbPlan.BindParameters)

	metadata["schemas"] = map[string]interface{}{
		"service_instance": map[string]interface{}{
			"create": map[string]interface{}{
				"openshift_form_definition": instanceFormDefn,
			},
			"update": map[string]interface{}{},
		},
		"service_binding": map[string]interface{}{
			"create": map[string]interface{}{
				"openshift_form_definition": bindingFormDefn,
			},
		},
	}

	return metadata
}

func initMetadataCopy(original map[string]interface{}) (map[string]interface{}, error) {
	dst := make(map[string]interface{})

	if original == nil {
		return dst, nil
	}
	bytes, err := json.Marshal(original)
	if err != nil {
		return dst, err
	}
	json.Unmarshal(bytes, &dst)
	if err != nil {
		return dst, err
	}
	return dst, nil
}

func createFormDefinition(params []apb.ParameterDescriptor) []interface{} {
	formDefinition := make([]interface{}, 0)

	if params == nil || len(params) == 0 {
		return formDefinition
	}

	for paramIdx := 0; paramIdx < len(params); {
		var item interface{}
		var numItems int

		pd := params[paramIdx]
		if pd.DisplayGroup == "" {
			item, numItems = createUIFormItem(pd, paramIdx)
		} else {
			item, numItems = createUIFormGroup(params, pd.DisplayGroup, paramIdx)
		}
		paramIdx = paramIdx + numItems

		formDefinition = append(formDefinition, item)
	}
	return formDefinition
}

func createUIFormGroup(params []apb.ParameterDescriptor, groupName string, paramIndex int) (formItem, int) {
	items := []interface{}{}

	for paramIndex < len(params) {
		pd := params[paramIndex]
		if pd.DisplayGroup != groupName {
			break
		}

		item, numItems := createUIFormItem(pd, paramIndex)
		items = append(items, item)
		paramIndex = paramIndex + numItems
	}

	group := formItem{
		Title: groupName,
		Type:  "fieldset",
		Items: items,
	}

	return group, len(items)
}

func createUIFormItem(pd apb.ParameterDescriptor, paramIndex int) (interface{}, int) {
	var item interface{}

	// if the name is the only key, it defaults to a string instead of a dictionary
	if pd.DisplayType == "" {
		item = pd.Name
	} else {
		item = formItem{
			Key:  pd.Name,
			Type: pd.DisplayType,
		}
	}

	return item, 1
}

// getType transforms an apb parameter type to a JSON Schema type
func getType(paramType string) schema.PrimitiveTypes {
	switch paramType {
	case "string", "enum":
		return []schema.PrimitiveType{schema.StringType}
	case "int":
		return []schema.PrimitiveType{schema.IntegerType}
	case "object":
		return []schema.PrimitiveType{schema.ObjectType}
	case "array":
		return []schema.PrimitiveType{schema.ArrayType}
	case "bool", "boolean":
		return []schema.PrimitiveType{schema.BooleanType}
	case "number":
		return []schema.PrimitiveType{schema.NumberType}
	case "nil", "null":
		return []schema.PrimitiveType{schema.NullType}
	}
	return []schema.PrimitiveType{schema.UnspecifiedType}
}

func parametersToSchema(plan apb.Plan) Schema {
	// parametersToSchema converts the apb parameters into a JSON Schema format.
	createProperties := extractProperties(plan.Parameters)
	createRequired := extractRequired(plan.Parameters)

	bindProperties := extractProperties(plan.BindParameters)
	bindRequired := extractRequired(plan.BindParameters)

	updatableProperties := extractUpdatable(plan.Parameters)
	updatableRequired := extractUpdatableRequired(createRequired, updatableProperties)

	// builds a Schema object for the various methods.
	s := Schema{
		ServiceInstance: ServiceInstance{
			Create: map[string]*schema.Schema{
				"parameters": {
					SchemaRef:  schema.SchemaURL,
					Type:       []schema.PrimitiveType{schema.ObjectType},
					Properties: createProperties,
					Required:   createRequired,
				},
			},
			Update: map[string]*schema.Schema{
				"parameters": {
					SchemaRef:  schema.SchemaURL,
					Type:       []schema.PrimitiveType{schema.ObjectType},
					Properties: updatableProperties,
					Required:   updatableRequired,
				},
			},
		},
		ServiceBinding: ServiceBinding{
			Create: map[string]*schema.Schema{
				"parameters": {
					SchemaRef:  schema.SchemaURL,
					Type:       []schema.PrimitiveType{schema.ObjectType},
					Properties: bindProperties,
					Required:   bindRequired,
				},
			},
		},
	}

	return s
}

func extractProperties(params []apb.ParameterDescriptor) map[string]*schema.Schema {
	properties := make(map[string]*schema.Schema)

	var patternRegex *regexp.Regexp
	var err error

	for _, pd := range params {
		k := pd.Name

		properties[k] = &schema.Schema{
			Title:       pd.Title,
			Description: pd.Description,
			Default:     pd.Default,
			Type:        getType(pd.Type),
		}

		// we can NOT set values on the Schema object if we want to be
		// omitempty. Setting maxlength to 0 is NOT the same as omitting it.
		// 0 is a worthless value for Maxlength so we will not set it
		if pd.Maxlength > 0 {
			properties[k].MaxLength = schema.Integer{Val: pd.Maxlength, Initialized: true}
		}

		// do not set the regexp if it does not compile
		if pd.Pattern != "" {
			patternRegex, err = regexp.Compile(pd.Pattern)
			properties[k].Pattern = patternRegex

			if err != nil {
				fmt.Printf("Invalid pattern: %s", err.Error())
			}
		}

		// setup enums
		if len(pd.Enum) > 0 {
			properties[k].Enum = make([]interface{}, len(pd.Enum))
			for i, v := range pd.Enum {
				properties[k].Enum[i] = v
			}
		}
	}

	return properties
}

func extractRequired(params []apb.ParameterDescriptor) []string {
	req := make([]string, 0, len(params))
	for _, param := range params {
		if param.Required {
			req = append(req, param.Name)
		}
	}
	return req
}

func extractUpdatable(params []apb.ParameterDescriptor) map[string]*schema.Schema {
	upd := make(map[string]*schema.Schema)
	for _, v := range params {
		if v.Updatable {
			k := v.Name
			upd[k] = &schema.Schema{
				Title:       v.Title,
				Description: v.Description,
				Default:     v.Default,
				Type:        getType(v.Type),
			}
		}
	}
	return upd
}

func extractUpdatableRequired(required []string, updatableProperties map[string]*schema.Schema) []string {
	var updReq []string

	for _, element := range required {
		if _, exists := updatableProperties[element]; exists {
			updReq = append(updReq, element)
		}
	}
	return updReq
}

// StateToLastOperation converts apb State objects into LastOperationStates.
func StateToLastOperation(state apb.State) LastOperationState {
	switch state {
	case apb.StateInProgress:
		return LastOperationStateInProgress
	case apb.StateSucceeded:
		return LastOperationStateSucceeded
	case apb.StateFailed:
		return LastOperationStateFailed
	default:
		return LastOperationStateFailed
	}
}
