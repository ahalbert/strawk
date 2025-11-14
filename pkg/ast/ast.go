package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/ahalbert/strawk/pkg/token"
)

type Node interface {
	GetToken() token.Token
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

type Block interface {
	GetStatements() []Statement
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

type ActionBlock struct {
	Statements []Statement
}

func (ab *ActionBlock) GetStatements() []Statement { return ab.Statements }
func (ab *ActionBlock) String() string             { return "" }

func (p *Program) GetToken() token.Token {
	if len(p.Statements) > 0 {
		return p.Statements[0].GetToken()
	} else {
		return token.Token{Type: token.ILLEGAL}
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type ExpressionStatement struct {
	Token       token.Token // the first token of the expression
	Expressions []Expression
}

func (es *ExpressionStatement) statementNode()        {}
func (es *ExpressionStatement) GetToken() token.Token { return es.Token }
func (es *ExpressionStatement) String() string {
	var out bytes.Buffer

	for _, exp := range es.Expressions {
		out.WriteString(exp.String())
	}

	return out.String()
}

type ActionBlockStatement struct {
	Token      token.Token // the { token
	Conditon   Expression
	Statements *ActionBlock
}

func (as *ActionBlockStatement) statementNode()             {}
func (as *ActionBlockStatement) GetToken() token.Token      { return as.Token }
func (as *ActionBlockStatement) GetStatements() []Statement { return as.Statements.Statements }
func (as *ActionBlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range as.Statements.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type BeginStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BeginStatement) statementNode()             {}
func (bs *BeginStatement) GetToken() token.Token      { return bs.Token }
func (bs *BeginStatement) GetStatements() []Statement { return bs.Statements }
func (bs *BeginStatement) String() string {
	var out bytes.Buffer

	out.WriteString("BEGIN {\n")
	for _, s := range bs.Statements {
		out.WriteString(s.String() + ";\n")
	}
	out.WriteString("}\n")

	return out.String()
}

type EndStatement struct {
	Token      token.Token
	Statements []Statement
}

func (es *EndStatement) statementNode()             {}
func (es *EndStatement) GetToken() token.Token      { return es.Token }
func (es *EndStatement) GetStatements() []Statement { return es.Statements }
func (es *EndStatement) String() string {
	var out bytes.Buffer

	for _, s := range es.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type AssignStatement struct {
	Token   token.Token // the { token
	Targets []Expression
	Values  []Expression
}

func (as *AssignStatement) statementNode()        {}
func (as *AssignStatement) GetToken() token.Token { return as.Token }
func (as *AssignStatement) String() string {
	var out bytes.Buffer

	for idx, s := range as.Targets {
		out.WriteString(s.String() + " = " + as.Values[idx].String())
	}

	return out.String()
}

type AssignAndModifyStatement struct {
	Token    token.Token // the { token
	Operator token.Token
	Target   Expression
	Value    Expression
}

func (ams *AssignAndModifyStatement) statementNode()        {}
func (ams *AssignAndModifyStatement) GetToken() token.Token { return ams.Token }
func (ams *AssignAndModifyStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ams.Target.String() + " " + ams.Operator.Literal + " " + ams.Value.String())

	return out.String()
}

type PrintStatement struct {
	Token       token.Token // the print token
	Expressions []Expression
}

func (ps *PrintStatement) statementNode()        {}
func (ps *PrintStatement) GetToken() token.Token { return ps.Token }
func (ps *PrintStatement) String() string {
	var out bytes.Buffer
	out.WriteString("print ")

	for idx, s := range ps.Expressions {
		out.WriteString(s.String())
		if idx < len(ps.Expressions) {
			out.WriteString(",")
		}
	}

	return out.String()
}

type PrintfStatement struct {
	Token       token.Token // the print token
	Expressions []Expression
}

func (ps *PrintfStatement) statementNode()        {}
func (ps *PrintfStatement) GetToken() token.Token { return ps.Token }
func (ps *PrintfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("printf ")

	for idx, s := range ps.Expressions {
		out.WriteString(s.String())
		if idx < len(ps.Expressions) {
			out.WriteString(",")
		}
	}

	return out.String()
}

type IfStatement struct {
	Token        token.Token
	Conditions   []Expression
	Consequences []*ActionBlock
	Else         *ActionBlock
}

func (is *IfStatement) statementNode()        {}
func (is *IfStatement) GetToken() token.Token { return is.Token }
func (is *IfStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type WhileStatement struct {
	Token          token.Token
	Condition      Expression
	Block          *ActionBlock
	ShouldBreak    bool
	ShouldContinue bool
}

func (ws *WhileStatement) statementNode()        {}
func (ws *WhileStatement) GetToken() token.Token { return ws.Token }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type DoWhileStatement struct {
	Token          token.Token
	Condition      Expression
	Block          *ActionBlock
	ShouldBreak    bool
	ShouldContinue bool
}

func (ds *DoWhileStatement) statementNode()        {}
func (ds *DoWhileStatement) GetToken() token.Token { return ds.Token }
func (ds *DoWhileStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type ForStatement struct {
	Token          token.Token
	Initialization Statement
	Condition      Expression
	Action         Statement
	Block          *ActionBlock
	ShouldBreak    bool
	ShouldContinue bool
}

func (fs *ForStatement) statementNode()        {}
func (fs *ForStatement) GetToken() token.Token { return fs.Token }
func (fs *ForStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type ForEachStatement struct {
	Token          token.Token
	VarName        *Identifier
	Array          *Identifier
	Block          *ActionBlock
	ShouldBreak    bool
	ShouldContinue bool
}

func (fs *ForEachStatement) statementNode()        {}
func (fs *ForEachStatement) GetToken() token.Token { return fs.Token }
func (fs *ForEachStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (cs *ReturnStatement) statementNode()        {}
func (cs *ReturnStatement) GetToken() token.Token { return cs.Token }
func (cs *ReturnStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()        {}
func (bs *BreakStatement) GetToken() token.Token { return bs.Token }
func (bs *BreakStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Name       Identifier
	Parameters []Identifier
	Body       *ActionBlock
}

func (fl *FunctionLiteral) statementNode()        {}
func (fl *FunctionLiteral) GetToken() token.Token { return fl.Token }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.Name.Value)
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()        {}
func (cs *ContinueStatement) GetToken() token.Token { return cs.Token }
func (cs *ContinueStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

type NextStatement struct {
	Token token.Token
}

func (ns *NextStatement) statementNode()        {}
func (ns *NextStatement) GetToken() token.Token { return ns.Token }
func (ns *NextStatement) String() string {
	var out bytes.Buffer

	return out.String()
}

// Expressions

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()       {}
func (ce *CallExpression) GetToken() token.Token { return ce.Token }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type TernaryExpression struct {
	Token     token.Token // The '(' token
	Condition Expression
	IfTrue    Expression
	IfFalse   Expression
}

func (te *TernaryExpression) expressionNode()       {}
func (te *TernaryExpression) GetToken() token.Token { return te.Token }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString(te.Condition.String() + " ? " + te.IfTrue.String() + " : " + te.IfFalse.String())

	return out.String()
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()       {}
func (pe *PrefixExpression) GetToken() token.Token { return pe.Token }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type PostfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode()       {}
func (pe *PostfixExpression) GetToken() token.Token { return pe.Token }
func (pe *PostfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Operator)
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()       {}
func (ie *InfixExpression) GetToken() token.Token { return ie.Token }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()       {}
func (i *Identifier) GetToken() token.Token { return i.Token }
func (i *Identifier) String() string        { return i.Value }

type NumericLiteral struct {
	Token token.Token
	Value float64
}

func (il *NumericLiteral) expressionNode()       {}
func (il *NumericLiteral) GetToken() token.Token { return il.Token }
func (il *NumericLiteral) String() string {
	if il.Value == float64(int(il.Value)) {
		return fmt.Sprintf("%d", int(il.Value))
	}
	return fmt.Sprintf("%.5g", il.Value)
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()       {}
func (sl *StringLiteral) GetToken() token.Token { return sl.Token }
func (sl *StringLiteral) String() string        { return sl.Value }

func NewLiteral(val string) Expression {
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return &StringLiteral{Value: val}
	}
	return &NumericLiteral{Value: parsed}
}

type RegexLiteral struct {
	Token token.Token
	Value string
}

func (rl *RegexLiteral) expressionNode()       {}
func (rl *RegexLiteral) GetToken() token.Token { return rl.Token }
func (rl *RegexLiteral) String() string        { return rl.Value }

type AssociativeArray struct {
	Token token.Token
	Array map[string]Expression
}

func (aa *AssociativeArray) expressionNode()       {}
func (aa *AssociativeArray) GetToken() token.Token { return aa.Token }
func (aa *AssociativeArray) String() string {
	var out bytes.Buffer
	entries := []string{}
	for k, v := range aa.Array {
		entries = append(entries, k+" : "+v.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(entries, ", "))
	out.WriteString("}")

	return out.String()
}

type ArrayIndexExpression struct {
	Token     token.Token
	ArrayName string
	IndexList []Expression
}

func (aie *ArrayIndexExpression) expressionNode()       {}
func (aie *ArrayIndexExpression) GetToken() token.Token { return aie.Token }
func (aie *ArrayIndexExpression) String() string {
	var out bytes.Buffer
	indicies := []string{}

	for _, i := range aie.IndexList {
		indicies = append(indicies, i.String())
	}
	out.WriteString(aie.ArrayName + "[" + strings.Join(indicies, ", ") + "]")

	return out.String()
}

type DeleteStatement struct {
	Token    token.Token
	ToDelete *ArrayIndexExpression
}

func (ds *DeleteStatement) statementNode()        {}
func (ds *DeleteStatement) GetToken() token.Token { return ds.Token }
func (ds *DeleteStatement) String() string {
	var out bytes.Buffer
	return out.String()
}
