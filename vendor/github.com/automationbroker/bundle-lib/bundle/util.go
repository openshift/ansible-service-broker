//
// Copyright (c) 2018 Red Hat, Inc.
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

package bundle

import (
	"fmt"
	"regexp"
	"strings"

	"encoding/json"

	schema "github.com/lestrrat/go-jsschema"
)

type formItem struct {
	Key   string        `json:"key,omitempty"`
	Title string        `json:"title,omitempty"`
	Type  string        `json:"type,omitempty"`
	Items []interface{} `json:"items,omitempty"`
}

// ConvertPlansToSchema - converts plans to schema
func ConvertPlansToSchema(plans []Plan) ([]SchemaPlan, error) {
	brokerPlans := make([]SchemaPlan, len(plans))
	for i, plan := range plans {
		schemas, err := parametersToSchema(plan)
		if err != nil {
			return nil, err
		}
		brokerPlans[i] = SchemaPlan{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Metadata:    extractBrokerPlanMetadata(plan),
			Free:        plan.Free,
			Bindable:    plan.Bindable,
			UpdatesTo:   plan.UpdatesTo,
			Schemas:     schemas,
		}
	}
	return brokerPlans, nil
}

func planUpdatable(apbPlans []Plan) bool {
	for _, plan := range apbPlans {
		if len(plan.UpdatesTo) > 0 {
			return true
		}
	}
	return false
}

func extractBrokerPlanMetadata(apbPlan Plan) map[string]interface{} {
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

func createFormDefinition(params []ParameterDescriptor) []interface{} {
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

func createUIFormGroup(params []ParameterDescriptor, groupName string, paramIndex int) (formItem, int) {
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

func createUIFormItem(pd ParameterDescriptor, paramIndex int) (interface{}, int) {
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
func getType(paramType string) (schema.PrimitiveTypes, error) {
	switch strings.ToLower(paramType) {
	case "string", "enum":
		return []schema.PrimitiveType{schema.StringType}, nil
	case "int", "integer":
		return []schema.PrimitiveType{schema.IntegerType}, nil
	case "object":
		return []schema.PrimitiveType{schema.ObjectType}, nil
	case "array":
		return []schema.PrimitiveType{schema.ArrayType}, nil
	case "bool", "boolean":
		return []schema.PrimitiveType{schema.BooleanType}, nil
	case "number":
		return []schema.PrimitiveType{schema.NumberType}, nil
	case "nil", "null":
		return []schema.PrimitiveType{schema.NullType}, nil
	}
	return nil, fmt.Errorf("Could not find the parameter type for: %v", paramType)
}

func parametersToSchema(plan Plan) (Schema, error) {
	// parametersToSchema converts the apb parameters into a JSON Schema format.
	createProperties, err := extractProperties(plan.Parameters)
	if err != nil {
		return Schema{}, err
	}
	createRequired := extractRequired(plan.Parameters)

	bindProperties, err := extractProperties(plan.BindParameters)
	if err != nil {
		return Schema{}, err
	}
	bindRequired := extractRequired(plan.BindParameters)

	updatableProperties, err := extractUpdatable(plan.Parameters)
	if err != nil {
		return Schema{}, err
	}
	updatableRequired := extractUpdatableRequired(createRequired, updatableProperties)

	// builds a Schema object for the various methods.
	s := Schema{
		ServiceInstance: ServiceInstanceSchema{
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
		ServiceBinding: ServiceBindingSchema{
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

	return s, nil
}

func extractProperties(params []ParameterDescriptor) (map[string]*schema.Schema, error) {
	properties := make(map[string]*schema.Schema)

	for _, pd := range params {
		k := pd.Name

		t, err := getType(pd.Type)
		if err != nil {
			return properties, err
		}

		properties[k] = &schema.Schema{
			Title:       pd.Title,
			Description: pd.Description,
			Default:     pd.Default,
			Type:        t,
		}

		setStringValidators(pd, properties[k])
		setNumberValidators(pd, properties[k])
		setEnum(pd, properties[k])
	}

	return properties, nil
}

func setStringValidators(pd ParameterDescriptor, prop *schema.Schema) {
	if prop.Type[0] != schema.StringType {
		return
	}

	// we can NOT set values on the Schema object if we want to be
	// omitempty. Setting maxlength to 0 is NOT the same as omitting it.
	// 0 is a worthless value for DeprecatedMaxlength so we will not set it

	// maxlength
	if pd.DeprecatedMaxlength > 0 {
		prop.MaxLength = schema.Integer{Val: pd.DeprecatedMaxlength, Initialized: true}
	}

	// max_length overrides maxlength
	if pd.MaxLength > 0 {
		prop.MaxLength = schema.Integer{Val: pd.MaxLength, Initialized: true}
	}
	// min_length
	if pd.MinLength > 0 {
		prop.MinLength = schema.Integer{Val: pd.MinLength, Initialized: true}
	}

	// do not set the regexp if it does not compile
	if pd.Pattern != "" {
		patternRegex, err := regexp.Compile(pd.Pattern)
		if err != nil {
			fmt.Printf("Invalid pattern: %s", err.Error())
			return
		}
		prop.Pattern = patternRegex
	}
}

func setNumberValidators(pd ParameterDescriptor, prop *schema.Schema) {
	if prop.Type[0] != schema.NumberType && prop.Type[0] != schema.IntegerType {
		return
	}

	// since 0 is not useful as a value for multipleOf,
	// we can use it as a float64 and not worry about nil
	if pd.MultipleOf > 0 {
		prop.MultipleOf = schema.Number{Val: pd.MultipleOf, Initialized: true}
	}

	// since 0 is a valid value for maximum, minimum, exclusiveMaximum, and exclusiveMinimum,
	// we have to allow for empty.
	if pd.Maximum != nil {
		prop.Maximum = schema.Number{Val: float64(*pd.Maximum), Initialized: true}
	}
	if pd.Minimum != nil {
		prop.Minimum = schema.Number{Val: float64(*pd.Minimum), Initialized: true}
	}

	// JSON Schema defines exclusiveMaximum and exclusiveMinimum as numbers separate from maximum and minimum
	// but go-jsschema defines ExclusiveMaximum and ExclusiveMinimum as bool and reuses Maximum and Minimum
	if pd.ExclusiveMaximum != nil {
		prop.Maximum = schema.Number{Val: float64(*pd.ExclusiveMaximum), Initialized: true}
		prop.ExclusiveMaximum = schema.Bool{Val: true, Default: false, Initialized: true}
	}
	if pd.ExclusiveMinimum != nil {
		prop.Minimum = schema.Number{Val: float64(*pd.ExclusiveMinimum), Initialized: true}
		prop.ExclusiveMinimum = schema.Bool{Val: true, Default: false, Initialized: true}
	}
}

func setEnum(pd ParameterDescriptor, prop *schema.Schema) {
	if len(pd.Enum) > 0 {
		prop.Enum = make([]interface{}, len(pd.Enum))
		for i, v := range pd.Enum {
			prop.Enum[i] = v
		}
	}
}

func extractRequired(params []ParameterDescriptor) []string {
	req := make([]string, 0, len(params))
	for _, param := range params {
		if param.Required {
			req = append(req, param.Name)
		}
	}
	return req
}

func extractUpdatable(params []ParameterDescriptor) (map[string]*schema.Schema, error) {
	upd := make(map[string]*schema.Schema)
	for _, v := range params {
		t, err := getType(v.Type)
		if err != nil {
			return upd, err
		}
		if v.Updatable {
			k := v.Name
			upd[k] = &schema.Schema{
				Title:       v.Title,
				Description: v.Description,
				Default:     v.Default,
				Type:        t,
			}

			setStringValidators(v, upd[k])
			setNumberValidators(v, upd[k])
			setEnum(v, upd[k])
		}
	}
	return upd, nil
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
