package billingexpr

import (
	"fmt"
	"math"
	"strings"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

// ScaleCurrency multiplies every tier's monetary result by factor while
// preserving request multipliers outside the tier. fixed() leaves are scaled
// at their literal so they continue to satisfy the fixed-price grammar.
func ScaleCurrency(expression string, factor float64) (string, error) {
	if factor <= 0 || math.IsNaN(factor) || math.IsInf(factor, 0) {
		return "", fmt.Errorf("currency factor must be finite and positive")
	}
	if _, err := CompileFromCache(expression); err != nil {
		return "", err
	}
	version, body := ParseExprVersion(expression)
	tree, err := parser.Parse(body)
	if err != nil {
		return "", fmt.Errorf("parse billing expression: %w", err)
	}
	patcher := &currencyScalePatcher{factor: factor}
	ast.Walk(&tree.Node, patcher)
	if patcher.err != nil {
		return "", patcher.err
	}
	if patcher.tiers == 0 {
		return "", fmt.Errorf("billing expression has no tier pricing leaf")
	}
	scaled := tree.Node.String()
	if version != DefaultExprVersion || strings.HasPrefix(expression, "v1:") {
		scaled = fmt.Sprintf("v%d:%s", version, scaled)
	}
	if _, err := CompileFromCache(scaled); err != nil {
		return "", fmt.Errorf("compile scaled billing expression: %w", err)
	}
	return scaled, nil
}

type currencyScalePatcher struct {
	factor float64
	tiers  int
	err    error
}

func (p *currencyScalePatcher) Visit(node *ast.Node) {
	if p.err != nil {
		return
	}
	call, ok := (*node).(*ast.CallNode)
	if !ok || len(call.Arguments) != 2 {
		return
	}
	callee, ok := call.Callee.(*ast.IdentifierNode)
	if !ok || callee.Value != "tier" {
		return
	}
	p.tiers++
	if ast.Find(call.Arguments[1], func(candidate ast.Node) bool {
		inner, ok := candidate.(*ast.CallNode)
		if !ok {
			return false
		}
		identifier, ok := inner.Callee.(*ast.IdentifierNode)
		return ok && identifier.Value == "tier"
	}) != nil {
		p.err = fmt.Errorf("nested tier pricing cannot be migrated safely")
		return
	}
	if fixed, ok := call.Arguments[1].(*ast.CallNode); ok {
		identifier, identifierOK := fixed.Callee.(*ast.IdentifierNode)
		if identifierOK && identifier.Value == "fixed" {
			if len(fixed.Arguments) != 1 {
				p.err = fmt.Errorf("fixed pricing leaf has invalid arguments")
				return
			}
			amount, ok := requestRuleNumber(fixed.Arguments[0])
			if !ok {
				p.err = fmt.Errorf("fixed pricing amount is not a numeric literal")
				return
			}
			fixed.Arguments[0] = &ast.FloatNode{Value: amount * p.factor}
			return
		}
	}
	call.Arguments[1] = &ast.BinaryNode{
		Operator: "*",
		Left:     call.Arguments[1],
		Right:    &ast.FloatNode{Value: p.factor},
	}
}
