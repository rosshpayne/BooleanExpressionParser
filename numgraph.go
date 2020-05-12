package main

import (
	"fmt"

	"github.com/GFilterExpressionParser/lexer"
	"github.com/GFilterExpressionParser/parser"
	"github.com/GFilterExpressionParser/token"
)

const (
	LPAREN uint8 = 1
)

const (
	NIL operator = "-"
)

func buildExprGraph(input string) *expression {

	type state struct {
		opr token.TokenType
	}

	var (
		tok         *token.Token
		loperand    *dGfunc
		roperand    *dGfunc
		operandL    bool            // put next INT in numL
		extendRight bool            // Used when a higher precedence operation detected. Assigns the latest expression to the right operand of the current expression.
		opr         token.TokenType // string
		opr_        token.TokenType // state held copy of opr
		e, en       *expression     // "e" points to current expression in graph while "en" is the latest expression to be created and added to the graph using addParent() or extendRight() functions.
		lp          []state
	)

	pushState := func() {
		s := state{opr: opr}
		lp = append(lp, s)
	}

	popState := func() {
		var s state
		s, lp = lp[len(lp)-1], lp[:len(lp)-1]
		opr_ = s.opr
	}

	// as the parser processes the input left to right it builds a tree (graph) by creating an expression as each operator is parsed and then immediately links
	// it to the previous expression. If the expression is at the same precedence level it links the new expression as the parent of the current expression. In the case
	// of higher precedence operations it links to the right of the current expression (func: extendRight). Walking the tree and evaluating each expression returns the final result.

	fmt.Printf("\n %s \n", input)

	l := lexer.New(input)
	p := parser.New(l)
	operandL = true

	// TODO - initial full parse to validate left and right parenthesis match

	for {

		tok = p.CurToken
		p.NextToken()
		fmt.Printf("\ntoken: %s\n", tok.Type)

		switch tok.Type {
		case token.EOF:
			break
		case token.LPAREN:
			//
			// LPAREN is represented in the graph by a "NULL" expression (node) consisting of operator "+" and left operand of 1.
			//
			// or ( true
			//

			pushState()

			if opr != "" {

				if loperand != nil {

					en, opr = makeExpr(loperand, opr, nil)
					if e == nil {
						e, en = en, nil
					} else {
						e = e.extendRight(en)
					}

				} else {

					en, opr = makeExpr(nil, opr, nil)
					if e == nil {
						e, en = en, nil
					} else {
						e = e.addParent(en)
					}
				}
			}
			//
			// add NULL expression representing "(". Following operation will be extend Right.
			//
			en = &expression{left: nil, opr: token.TokenType("-"), right: nil}
			if e == nil {
				e, en = en, nil
			} else {
				e = e.extendRight(en)
			}

			extendRight = true
			operandL = true

		case token.RPAREN:

			popState()

			// navigate current expression e, up to next LPARAM expression
			for e = e.parent; e.parent != nil && e.opr != "-"; e = e.parent {
			}

			if e.parent != nil && e.parent.opr != "-" {
				// opr_ represents the operator that existed at the associated "(". Sourced from state.
				if opr_ == "AND" {
					fmt.Println("opr_ is adjusting to e.parent ", opr_)
					e = e.parent
				}
			}

		case token.TRUE, token.FALSE:

			bl := false
			if tok.Type == token.TRUE {
				bl = true
			}
			//
			// look ahead to next operator and check for higher precedence operation
			//
			tok := p.CurToken
			if opr == token.OR && tok.Type == token.AND {
				//
				if extendRight {
					en, opr = makeExpr(loperand, opr, nil)
					if e == nil {
						e, en = en, nil
					} else {
						e = e.extendRight(en)
						extendRight = false
					}

				} else if loperand == nil {
					// add operator only node to graph - no left, right operands. addParent will attach left, and future ExtendRIght will attach right.
					en, opr = makeExpr(nil, opr, nil)
					e = e.addParent(en)

				} else {
					// make expr for existing numL and opr
					en, opr = makeExpr(loperand, opr, nil)
					if e == nil {
						e, en = en, nil
					} else {
						e = e.addParent(en)
					}
				}
				// all higher precedence operations or explicit (), perform an "extendRight" to create a new branch in the graph.
				extendRight = true
				// new branches begin with a left operand
				operandL = true
			}

			if operandL {

				loperand = &dGfunc{value: bl}
				operandL = false

			} else {

				roperand = &dGfunc{value: bl}

				if loperand != nil {
					en, opr = makeExpr(loperand, opr, roperand)
					if e == nil {
						e, en = en, nil
					} else {
						e = e.extendRight(en)
					}

				} else {

					en, opr = makeExpr(nil, opr, roperand)

					if extendRight {
						e = e.extendRight(en)
					} else {
						e = e.addParent(en)
					}

				}
				extendRight = false
				operandL = false
				loperand = nil

			}

		case token.OR, token.AND:

			opr = tok.Type

		case token.NOT:

			// handle any operands not formed into expression
			if loperand != nil {
				en, opr = makeExpr(loperand, opr, nil)
				if e == nil {
					e, en = en, nil
				} else {
					e = e.extendRight(en)
				}
				extendRight = true
			}
			// create NOT node (expression)
			en, opr = makeExpr(nil, token.NOT, nil)
			if e == nil {
				e, en = en, nil
			} else {
				if extendRight {
					e = e.extendRight(en)
				} else {
					e = e.addParent(en)
				}
			}
			// make following expression extend right after NOT expression
			extendRight = true
			operandL = false

		}
		if tok.Type == token.EOF {
			break
		}

	}
	return findRoot(e)
}
