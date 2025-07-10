package semantic

import (
	"fmt"
	"strings"
	"redis-analyzer-api/parser"
)

// SemanticError representa un error semántico
type SemanticError struct {
	Message  string
	Command  string
	Position int
	Type     string
}

func (e SemanticError) Error() string {
	return fmt.Sprintf("Semantic error in command '%s': %s", e.Command, e.Message)
}

// ValidationResult contiene el resultado de la validación semántica
type ValidationResult struct {
	Valid      bool
	Errors     []SemanticError
	Warnings   []string
	CommandInfo map[string]interface{}
}

// CommandSpec define la especificación de un comando Redis
type CommandSpec struct {
	Name         string
	MinArgs      int
	MaxArgs      int // -1 para ilimitado
	KeyPosition  int // posición de la clave (0-based, -1 si no tiene clave)
	ValueTypes   []string // tipos esperados para cada argumento
	Options      map[string]OptionSpec
	Description  string
}

// OptionSpec define la especificación de una opción de comando
type OptionSpec struct {
	HasValue     bool
	ValueType    string
	Description  string
	Conflicts    []string // opciones que no pueden usarse juntas
}

// Analyzer representa el analizador semántico
type Analyzer struct {
	commands map[string]CommandSpec
}

// New crea un nuevo analizador semántico
func New() *Analyzer {
	analyzer := &Analyzer{
		commands: make(map[string]CommandSpec),
	}
	analyzer.initializeCommands()
	return analyzer
}

// initializeCommands inicializa las especificaciones de comandos Redis
func (a *Analyzer) initializeCommands() {
	// Comandos básicos de strings
	a.commands["GET"] = CommandSpec{
		Name:        "GET",
		MinArgs:     1,
		MaxArgs:     1,
		KeyPosition: 0,
		ValueTypes:  []string{"key"},
		Description: "Get the value of a key",
	}
	
	a.commands["SET"] = CommandSpec{
		Name:        "SET",
		MinArgs:     2,
		MaxArgs:     -1,
		KeyPosition: 0,
		ValueTypes:  []string{"key", "value"},
		Options: map[string]OptionSpec{
			"EX": {HasValue: true, ValueType: "integer", Description: "Set expiry in seconds"},
			"PX": {HasValue: true, ValueType: "integer", Description: "Set expiry in milliseconds", Conflicts: []string{"EX"}},
			"NX": {HasValue: false, Description: "Only set if key doesn't exist", Conflicts: []string{"XX"}},
			"XX": {HasValue: false, Description: "Only set if key exists", Conflicts: []string{"NX"}},
		},
		Description: "Set the string value of a key",
	}
	
	a.commands["DEL"] = CommandSpec{
		Name:        "DEL",
		MinArgs:     1,
		MaxArgs:     -1,
		KeyPosition: 0,
		ValueTypes:  []string{"key"},
		Description: "Delete one or more keys",
	}
	
	// Comandos de hash
	a.commands["HGET"] = CommandSpec{
		Name:        "HGET",
		MinArgs:     2,
		MaxArgs:     2,
		KeyPosition: 0,
		ValueTypes:  []string{"key", "field"},
		Description: "Get the value of a hash field",
	}
	
	a.commands["HSET"] = CommandSpec{
		Name:        "HSET",
		MinArgs:     3,
		MaxArgs:     -1,
		KeyPosition: 0,
		ValueTypes:  []string{"key", "field", "value"},
		Description: "Set the string value of a hash field",
	}
	
	// Comandos de sorted sets
	a.commands["ZADD"] = CommandSpec{
		Name:        "ZADD",
		MinArgs:     3,
		MaxArgs:     -1,
		KeyPosition: 0,
		ValueTypes:  []string{"key", "score", "member"},
		Options: map[string]OptionSpec{
			"NX": {HasValue: false, Description: "Only add new elements", Conflicts: []string{"XX"}},
			"XX": {HasValue: false, Description: "Only update existing elements", Conflicts: []string{"NX"}},
		},
		Description: "Add one or more members to a sorted set",
	}
	
	a.commands["ZRANGE"] = CommandSpec{
		Name:        "ZRANGE",
		MinArgs:     3,
		MaxArgs:     -1,
		KeyPosition: 0,
		ValueTypes:  []string{"key", "start", "stop"},
		Options: map[string]OptionSpec{
			"WITHSCORES": {HasValue: false, Description: "Return scores along with members"},
		},
		Description: "Return a range of members in a sorted set",
	}
	
	// Comandos de utilidad
	a.commands["SCAN"] = CommandSpec{
		Name:        "SCAN",
		MinArgs:     1,
		MaxArgs:     -1,
		KeyPosition: -1, // SCAN no opera sobre una clave específica
		ValueTypes:  []string{"cursor"},
		Options: map[string]OptionSpec{
			"MATCH": {HasValue: true, ValueType: "pattern", Description: "Match pattern"},
			"COUNT": {HasValue: true, ValueType: "integer", Description: "Number of elements to return"},
			"TYPE":  {HasValue: true, ValueType: "string", Description: "Filter by type"},
		},
		Description: "Incrementally iterate over keys",
	}
}

// ValidateCommand valida un comando Redis parseado
func (a *Analyzer) ValidateCommand(cmd *parser.RedisCommand) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		Errors:      []SemanticError{},
		Warnings:    []string{},
		CommandInfo: make(map[string]interface{}),
	}
	
	commandName := strings.ToUpper(cmd.Command.Value)
	spec, exists := a.commands[commandName]
	
	if !exists {
		result.Valid = false
		result.Errors = append(result.Errors, SemanticError{
			Message: fmt.Sprintf("Unknown command: %s", commandName),
			Command: commandName,
			Type:    "UNKNOWN_COMMAND",
		})
		return result
	}
	
	// Validar número de argumentos
	argCount := len(cmd.Arguments)
	if argCount < spec.MinArgs {
		result.Valid = false
		result.Errors = append(result.Errors, SemanticError{
			Message: fmt.Sprintf("Too few arguments. Expected at least %d, got %d", spec.MinArgs, argCount),
			Command: commandName,
			Type:    "INSUFFICIENT_ARGS",
		})
	}
	
	if spec.MaxArgs != -1 && argCount > spec.MaxArgs {
		result.Valid = false
		result.Errors = append(result.Errors, SemanticError{
			Message: fmt.Sprintf("Too many arguments. Expected at most %d, got %d", spec.MaxArgs, argCount),
			Command: commandName,
			Type:    "EXCESSIVE_ARGS",
		})
	}
	
	// Validar tipos de argumentos
	a.validateArgumentTypes(cmd, spec, &result)
	
	// Validar opciones
	a.validateOptions(cmd, spec, &result)
	
	// Agregar información del comando
	result.CommandInfo["name"] = commandName
	result.CommandInfo["description"] = spec.Description
	result.CommandInfo["has_key"] = spec.KeyPosition >= 0
	if spec.KeyPosition >= 0 && spec.KeyPosition < len(cmd.Arguments) {
		result.CommandInfo["key"] = cmd.Arguments[spec.KeyPosition].String()
	}
	
	return result
}

// validateArgumentTypes valida los tipos de argumentos
func (a *Analyzer) validateArgumentTypes(cmd *parser.RedisCommand, spec CommandSpec, result *ValidationResult) {
	for i, arg := range cmd.Arguments {
		if i >= len(spec.ValueTypes) {
			break // No hay más especificaciones de tipo
		}
		
		expectedType := spec.ValueTypes[i]
		actualType := arg.Type()
		
		switch expectedType {
		case "key":
			if actualType != "Identifier" && actualType != "StringLiteral" && actualType != "PatternExpression" {
				result.Errors = append(result.Errors, SemanticError{
					Message: fmt.Sprintf("Argument %d should be a key (identifier or string), got %s", i+1, actualType),
					Command: cmd.Command.Value,
					Type:    "TYPE_MISMATCH",
				})
				result.Valid = false
			}
		case "value":
			// Los valores pueden ser de cualquier tipo
		case "field":
			if actualType != "Identifier" && actualType != "StringLiteral" {
				result.Errors = append(result.Errors, SemanticError{
					Message: fmt.Sprintf("Argument %d should be a field name, got %s", i+1, actualType),
					Command: cmd.Command.Value,
					Type:    "TYPE_MISMATCH",
				})
				result.Valid = false
			}
		case "score":
			if actualType != "IntegerLiteral" && actualType != "FloatLiteral" {
				result.Errors = append(result.Errors, SemanticError{
					Message: fmt.Sprintf("Argument %d should be a numeric score, got %s", i+1, actualType),
					Command: cmd.Command.Value,
					Type:    "TYPE_MISMATCH",
				})
				result.Valid = false
			}
		case "member":
			if actualType != "Identifier" && actualType != "StringLiteral" {
				result.Errors = append(result.Errors, SemanticError{
					Message: fmt.Sprintf("Argument %d should be a member name, got %s", i+1, actualType),
					Command: cmd.Command.Value,
					Type:    "TYPE_MISMATCH",
				})
				result.Valid = false
			}
		case "cursor":
			if actualType != "IntegerLiteral" {
				result.Errors = append(result.Errors, SemanticError{
					Message: fmt.Sprintf("Argument %d should be a cursor (integer), got %s", i+1, actualType),
					Command: cmd.Command.Value,
					Type:    "TYPE_MISMATCH",
				})
				result.Valid = false
			}
		case "start", "stop":
			if actualType != "IntegerLiteral" {
				result.Errors = append(result.Errors, SemanticError{
					Message: fmt.Sprintf("Argument %d should be an index (integer), got %s", i+1, actualType),
					Command: cmd.Command.Value,
					Type:    "TYPE_MISMATCH",
				})
				result.Valid = false
			}
		}
	}
}

// validateOptions valida las opciones del comando
func (a *Analyzer) validateOptions(cmd *parser.RedisCommand, spec CommandSpec, result *ValidationResult) {
	usedOptions := make(map[string]bool)
	
	for i := 0; i < len(cmd.Arguments); i++ {
		arg := cmd.Arguments[i]
		if arg.Type() == "KeywordExpression" {
			optionName := strings.ToUpper(arg.String())
			
			// Verificar si la opción es válida para este comando
			optionSpec, exists := spec.Options[optionName]
			if !exists {
				result.Warnings = append(result.Warnings, 
					fmt.Sprintf("Unknown option '%s' for command %s", optionName, cmd.Command.Value))
				continue
			}
			
			// Verificar conflictos
			for _, conflict := range optionSpec.Conflicts {
				if usedOptions[conflict] {
					result.Errors = append(result.Errors, SemanticError{
						Message: fmt.Sprintf("Option '%s' conflicts with '%s'", optionName, conflict),
						Command: cmd.Command.Value,
						Type:    "OPTION_CONFLICT",
					})
					result.Valid = false
				}
			}
			
			usedOptions[optionName] = true
			
			// Verificar si la opción requiere un valor
			if optionSpec.HasValue {
				if i+1 >= len(cmd.Arguments) {
					result.Errors = append(result.Errors, SemanticError{
						Message: fmt.Sprintf("Option '%s' requires a value", optionName),
						Command: cmd.Command.Value,
						Type:    "MISSING_OPTION_VALUE",
					})
					result.Valid = false
				} else {
					// Validar el tipo del valor de la opción
					valueArg := cmd.Arguments[i+1]
					a.validateOptionValue(optionName, optionSpec, valueArg, result)
					i++ // Saltar el valor de la opción
				}
			}
		}
	}
}

// validateOptionValue valida el valor de una opción
func (a *Analyzer) validateOptionValue(optionName string, spec OptionSpec, value parser.Expression, result *ValidationResult) {
	valueType := value.Type()
	
	// Obtener el nombre del comando de forma segura
	commandName := "unknown"
	if name, ok := result.CommandInfo["name"]; ok && name != nil {
		if nameStr, ok := name.(string); ok {
			commandName = nameStr
		}
	}
	
	switch spec.ValueType {
	case "integer":
		if valueType != "IntegerLiteral" {
			result.Errors = append(result.Errors, SemanticError{
				Message: fmt.Sprintf("Option '%s' expects an integer value, got %s", optionName, valueType),
				Command: commandName,
				Type:    "OPTION_TYPE_MISMATCH",
			})
			result.Valid = false
		} else {
			// Validar rangos específicos
			if intLit, ok := value.(*parser.IntegerLiteral); ok {
				if optionName == "EX" || optionName == "PX" {
					if intLit.Value <= 0 {
						result.Errors = append(result.Errors, SemanticError{
							Message: fmt.Sprintf("Expiration time must be positive, got %d", intLit.Value),
							Command: commandName,
							Type:    "INVALID_VALUE_RANGE",
						})
						result.Valid = false
					}
				}
			}
		}
	case "pattern":
		if valueType != "PatternExpression" && valueType != "StringLiteral" && valueType != "Identifier" {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Option '%s' expects a pattern, got %s", optionName, valueType))
		}
	case "string":
		if valueType != "StringLiteral" && valueType != "Identifier" {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Option '%s' expects a string, got %s", optionName, valueType))
		}
	}
}

// ValidateProgram valida un programa completo con múltiples comandos
func (a *Analyzer) ValidateProgram(program *parser.Program) []ValidationResult {
	results := make([]ValidationResult, 0, len(program.Statements))
	
	for _, stmt := range program.Statements {
		if cmd, ok := stmt.(*parser.RedisCommand); ok {
			result := a.ValidateCommand(cmd)
			results = append(results, result)
		}
	}
	
	return results
}

// GetCommandSpecs devuelve todas las especificaciones de comandos
func (a *Analyzer) GetCommandSpecs() map[string]CommandSpec {
	return a.commands
}

// AddCommandSpec añade una nueva especificación de comando
func (a *Analyzer) AddCommandSpec(spec CommandSpec) {
	a.commands[strings.ToUpper(spec.Name)] = spec
}

