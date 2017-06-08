package broker

import (
	"os"
	"path"
	"regexp"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
)

func ProjectRoot() string {
	gopath := os.Getenv("GOPATH")
	rootPath := path.Join(gopath, "src", "github.com", "openshift",
		"ansible-service-broker")
	return rootPath
}

// TODO: This is going to have to be expanded much more to support things like
// parameters (will need to get passed through as metadata
func SpecToService(spec *apb.Spec) Service {
	parameterDescriptors := make(map[string]interface{})
	parameterDescriptors["parameters"] = spec.Parameters
	for k, v := range spec.Metadata {
		parameterDescriptors[k] = v
	}

	// default plan, used to be in hack.go
	plans := []Plan{
		{
			ID:          uuid.Parse("4c10ff42-be89-420a-9bab-27a9bef9aed8"),
			Name:        "default",
			Description: "Default plan",
			Free:        true,
			Schemas:     ParametersToSchema(spec.Parameters),
			// leaving Bindable undefined, defaults to Service value
		},
	}

	retSvc := Service{
		ID:          uuid.Parse(spec.Id),
		Name:        spec.Name,
		Description: spec.Description,
		Tags:        make([]string, len(spec.Tags)),
		Bindable:    spec.Bindable,
		Plans:       plans,
		// leaving Metadata empty
	}

	copy(retSvc.Tags, spec.Tags)
	return retSvc
}

func getType(paramType string) schema.PrimitiveTypes {
	switch paramType {
	case "string":
		return []schema.PrimitiveType{schema.StringType}
	case "int":
		return []schema.PrimitiveType{schema.IntegerType}
	case "object":
		return []schema.PrimitiveType{schema.ObjectType}
	case "array", "enum":
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

func ParametersToSchema(params []map[string]*apb.ParameterDescriptor) Schema {
	properties := make(map[string]*schema.Schema)

	for _, paramMap := range params {
		for k, pd := range paramMap {
			regex, _ := regexp.Compile(pd.Pattern)
			properties[k] = &schema.Schema{
				Title:       pd.Title,
				Description: pd.Description,
				Default:     pd.Default,
				MaxLength:   schema.Integer{Val: pd.Maxlength, Initialized: true},
				Pattern:     regex,
				Type:        getType(pd.Type),
				//Enum:        pd.Enum, deal with this later
			}
		}
	}

	s := Schema{
		ServiceInstance: ServiceInstance{
			Create: map[string]*schema.Schema{
				"parameters": {
					SchemaRef:  schema.SchemaURL,
					Type:       []schema.PrimitiveType{schema.ObjectType},
					Properties: properties,
					//Required:   required,
				},
			},
		},
	}

	return s
}

/*
func createSchema(key string, pd *apb.ParameterDescriptor) *schema.Schema {
	return nil
}
*/

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
