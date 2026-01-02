package engine

import (
	"log"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"sakin-go/pkg/models"
)

// compiledRule caches the VM program for a rule
type compiledRule struct {
	Rule    *models.Rule
	Program *vm.Program
}

// Engine evaluates events against rules.
type Engine struct {
	rules map[string]*compiledRule
}

func NewEngine() *Engine {
	return &Engine{
		rules: make(map[string]*compiledRule),
	}
}

// LoadRules compiles and loads rules into the engine.
func (e *Engine) LoadRules(rules []*models.Rule) {
	newRules := make(map[string]*compiledRule)

	for _, r := range rules {
		// Compile expression: e.g., "Event.Severity == 'critical' && Event.Source == 'firewall'"
		// 'Env' defines the structure available in the expression.
		program, err := expr.Compile(r.Condition, expr.Env(map[string]interface{}{
			"Event": &models.Event{},
		}))

		if err != nil {
			log.Printf("[Engine] Failed to compile rule %s: %v", r.Name, err)
			continue
		}

		newRules[r.ID] = &compiledRule{
			Rule:    r,
			Program: program,
		}
	}

	e.rules = newRules
	log.Printf("[Engine] Loaded %d rules", len(e.rules))
}

// Evaluate checks an event against all loaded rules.
// Returns a list of rules that matched.
func (e *Engine) Evaluate(evt *models.Event) []*models.Rule {
	var matches []*models.Rule

	env := map[string]interface{}{
		"Event": evt,
	}

	for _, cr := range e.rules {
		output, err := expr.Run(cr.Program, env)
		if err != nil {
			log.Printf("[Engine] Runtime error in rule %s: %v", cr.Rule.Name, err)
			continue
		}

		if matched, ok := output.(bool); ok && matched {
			matches = append(matches, cr.Rule)
		}
	}

	return matches
}
