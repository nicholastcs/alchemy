package system

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	manifest       executionEnv = "manifest"
	formValidation executionEnv = "formValidation"
)

var (
	cache = cachedProgramsByEnv{
		manifest:       map[string]cel.Program{},
		formValidation: map[string]cel.Program{},
	}

	manifestEnv       *cel.Env
	formValidationEnv *cel.Env
)

type executionEnv string
type cachedProgramsByEnv map[executionEnv]map[string]cel.Program

func init() {
	// example: quantity("10M") will result in 10000000000 in double
	// type format. Which they are useful for situations to compare
	// Kubernetes resource values such as `quantity(this) <
	// quantity("512M")`.
	k8sParseQuantity := cel.Function("quantity",
		cel.Overload("quantity_string",
			[]*cel.Type{cel.StringType},
			cel.DoubleType,
			cel.UnaryBinding(func(u ref.Val) ref.Val {
				qtyLiteral := fmt.Sprintf("%v", u)
				qty, err := resource.ParseQuantity(qtyLiteral)
				if err != nil {
					return types.NewErr("unable to parse string literal '%s' to Kubernetes resource quantity: %w", qtyLiteral, err)
				}

				nativeValByScale := qty.AsFloat64Slow()
				return types.Double(nativeValByScale)
			}),
		),
	)

	manifestEnv, _ = cel.NewEnv([]cel.EnvOption{
		cel.Variable("apiVersion", cel.StringType),
		cel.Variable("kind", cel.StringType),
		cel.Variable("metadata", cel.AnyType),
		cel.Variable("spec", cel.AnyType),
		cel.Variable("status", cel.AnyType),
		k8sParseQuantity,
	}...)

	formValidationEnv, _ = cel.NewEnv([]cel.EnvOption{
		cel.Variable("this", cel.AnyType),
		cel.Variable("result", cel.MapType(cel.StringType, cel.AnyType)),
		k8sParseQuantity,
	}...)
}

// ExecuteCELOnManifest is a function that retrieves field values from CEL
// expression.
func ExecuteCELOnManifest(input core.AbstractedManifest, celExpression string) (ref.Val, error) {
	var mapstr map[string]interface{}
	err := mapstructure.Decode(input, &mapstr)
	if err != nil {
		return nil, err
	}

	if program, ok := cache[manifest][celExpression]; ok {
		out, _, err := program.Eval(mapstr)
		if err != nil {
			return nil, fmt.Errorf("evaluation error: %s", err)
		}

		return out, nil
	}

	ast, issues := manifestEnv.Compile(celExpression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("type-check error found: %w", issues.Err())
	}

	program, err := manifestEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program construction error: %s", err)
	}

	out, _, err := program.Eval(mapstr)
	if err != nil {
		return nil, fmt.Errorf("evaluation error: %s", err)
	}

	cache[manifest][celExpression] = program

	return out, nil
}

// ExecuteCELOnFormValidation is a function that validates input based on
// CEL expression.
func ExecuteCELOnFormValidation(input map[string]interface{}, celExpression string) (bool, error) {
	var out ref.Val
	var err error

	if program, ok := cache[formValidation][celExpression]; ok {
		out, _, err = program.Eval(input)
		if err != nil {
			return false, fmt.Errorf("evaluation error: %s", err)
		}
	} else {
		// TODO: preflight will flag out those compilation errors
		// before form is generated
		ast, issues := formValidationEnv.Compile(celExpression)
		if issues != nil && issues.Err() != nil {
			return false, issues.Err()
		}

		program, err := formValidationEnv.Program(ast)
		if err != nil {
			return false, fmt.Errorf("program construction error: %s", err)
		}

		out, _, err = program.Eval(input)
		if err != nil {
			return false, fmt.Errorf("evaluation error: %s", err)
		}

		cache[formValidation][celExpression] = program
	}

	if out.Type() != cel.BoolType {
		err := fmt.Errorf("output type must be boolean, but found '%s'", out.Type().TypeName())
		return false, err
	}

	outcome, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("unable to assert type boolean")
	}

	return outcome, nil
}
