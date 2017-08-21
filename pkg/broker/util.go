package broker

import (
	"fmt"
	"regexp"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

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
			ID:          plan.Name,
			Name:        plan.Name,
			Description: plan.Description,
			Metadata:    updateMetadata(plan.Metadata, plan.Parameters),
			Free:        plan.Free,
			Bindable:    plan.Bindable,
			Schemas:     parametersToSchema(plan.Parameters),
		}
		i++
	}
	return brokerPlans
}

func updateMetadata(metadata map[string]interface{}, params []apb.ParameterDescriptor) map[string]interface{} {
	if params == nil || len(params) == 0 {
		return metadata
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	formDefinition := make([]interface{}, 0)

	for paramIdx := 0; paramIdx < len(params); {
		var formItem interface{}
		pd := params[paramIdx]
		if pd.DisplayGroup == "" {
			formItem, paramIdx = createUIFormItem(pd, paramIdx)
		} else {
			formItem, paramIdx = createUIFormGroup(params, pd.DisplayGroup, paramIdx)
		}

		formDefinition = append(formDefinition, formItem)
	}

	metadata["schemas"] = map[string]interface{}{
		"service_instance": map[string]interface{}{
			"create": map[string]interface{}{
				"form_definition": formDefinition,
			},
			"update": map[string]interface{}{},
		},
		"service_binding": map[string]interface{}{
			"create": map[string]interface{}{
				"form_definition": formDefinition,
			},
		},
	}

	return metadata
}

func createUIFormGroup(params []apb.ParameterDescriptor, groupName string, paramIndex int) (map[string]interface{}, int) {
	var formItem interface{}
	items := []interface{}{}

	for paramIndex < len(params) {
		pd := params[paramIndex]
		if pd.DisplayGroup != groupName {
			break
		}

		formItem, paramIndex = createUIFormItem(pd, paramIndex)
		items = append(items, formItem)
	}

	group := make(map[string]interface{})
	group["type"] = "fieldset"
	group["items"] = items
	return group, paramIndex + 1
}

func createUIFormItem(pd apb.ParameterDescriptor, paramIndex int) (interface{}, int) {
	var formItem interface{}

	// if the name is the only key, it defaults to a string instead of a dictionary
	if pd.DisplayType == "" {
		formItem = pd.Name
	} else {
		formItem = map[string]string{
			"key":  pd.Name,
			"type": pd.DisplayType,
		}
	}

	return formItem, paramIndex + 1
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

func parametersToSchema(params []apb.ParameterDescriptor) Schema {
	// parametersToSchema converts the apb parameters into a JSON Schema format.
	properties := make(map[string]*schema.Schema)
	required := extractRequired(params)

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

	// builds a Schema object for the various methods.
	s := Schema{
		ServiceInstance: ServiceInstance{
			Create: map[string]*schema.Schema{
				"parameters": {
					SchemaRef:  schema.SchemaURL,
					Type:       []schema.PrimitiveType{schema.ObjectType},
					Properties: properties,
					Required:   required,
				},
			},
			Update: map[string]*schema.Schema{},
		},
		ServiceBinding: ServiceBinding{
			Create: map[string]*schema.Schema{
				"parameters": {
					SchemaRef:  schema.SchemaURL,
					Type:       []schema.PrimitiveType{schema.ObjectType},
					Properties: properties,
				},
			},
		},
	}

	return s
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
