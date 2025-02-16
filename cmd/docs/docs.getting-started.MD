## Getting started

1. Fork this Alchemy code base to your private code repository.

2. Your playground is in `embed` directory, where you will write your Form and CodeTemplate API implementations.

3. After your implementations merged to your `master` branch, you could deliver them through your CICD pipeline or simply use `go install` in consumers' machine.

## Example API implementations

### Form API usage
```YAML
apiVersion: alchemy.io/v1alpha
kind: Form
metadata:
  name: test-form
spec:
  confirmationRequired: true  # (1)
  fields:
    - name: name
      title: Name
      description: Name of service
      inputType: text  # (2)
      constraint:
        cel:  # (3)
          expressions:
            - message: "length of name must be greater than 0."
              value: "this.size() > 0" # (4)
```

Above is the very first Form API in `v1alpha`. With the numbers in parentheses:

1. *Confirmation required* - usually you want to make this as `true`. This is because after the form was filled, it might override some code locally.

2. *Input types* - which can be either `text`, `multiline-text`, `single-select-numerical`, `multi-select-numerical`, `single-select-text` or `multi-select-text`:
   ```
   const (
           TextInputType                  string = "text"
           NumericalInputType             string = "numerical"
           MultilineTextInputType         string = "multiline-text"
           SingleSelectNumericalInputType string = "single-select-numerical"
           MultiSelectNumericalInputType  string = "multi-select-numerical"
           SingleSelectTextInputType      string = "single-select-text"
           MultiSelectTextInputType       string = "multi-select-text"
           BooleanInputType               string = "boolean"
   )
   ```
   In general, they are self-explanatory. They are basically indicating what kind of field need to be generated in Charmbracelet/Huh API.

3. *Form validations* - with Charmbracelet/Huh API builtin validation, it is further extended to support [CEL expression evaluation](https://github.com/google/cel-go). 

4. The CEL expression `this.size() > 0` - indicates that the length of the name must be greater than `0`.
