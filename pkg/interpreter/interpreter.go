package interpreter

import (
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ahalbert/strawk/pkg/ast"
	"github.com/ahalbert/strawk/pkg/token"
)

type Interpreter struct {
	BeginBlocks                  []*ast.BeginStatement
	EndBlocks                    []*ast.EndStatement
	Rules                        []ast.Statement
	curStatement                 ast.Statement //For reporting line number a runtime error hits
	Program                      *ast.Program
	Input                        string
	Output                       io.Writer
	WasNextStatementHit          bool
	WasFatalErrorHit             bool
	InputPostion                 int
	Stack                        []CallStackEntry
	GlobalVariables              map[string]ast.Expression
	StdLibFunctions              map[string]func(*Interpreter, []ast.Expression) ast.Expression
	UserDefinedFunctions         map[string]*ast.FunctionLiteral
	mostRecentRegexCaptureGroups map[string]ast.Expression
}

type CallStackEntry struct {
	isFunction     bool
	LocalVariables map[string]ast.Expression
}

func NewInterpreter(program *ast.Program, out io.Writer) *Interpreter {
	i := &Interpreter{
		Program:              program,
		Output:               out,
		GlobalVariables:      make(map[string]ast.Expression),
		StdLibFunctions:      make(map[string]func(*Interpreter, []ast.Expression) ast.Expression),
		UserDefinedFunctions: make(map[string]*ast.FunctionLiteral),
	}
	i.resetStack()
	i.InputPostion = 0
	for _, stmt := range program.Statements {
		switch stmt.(type) {
		case *ast.BeginStatement:
			i.BeginBlocks = append(i.BeginBlocks, stmt.(*ast.BeginStatement))
		case *ast.EndStatement:
			i.EndBlocks = append(i.EndBlocks, stmt.(*ast.EndStatement))
		case *ast.FunctionLiteral:
			i.UserDefinedFunctions[stmt.(*ast.FunctionLiteral).Name.Value] = stmt.(*ast.FunctionLiteral)
		default:
			i.Rules = append(i.Rules, stmt)
		}
	}

	i.StdLibFunctions["atan2"] = Atan2
	i.StdLibFunctions["cos"] = Cos
	i.StdLibFunctions["sin"] = Sin
	i.StdLibFunctions["exp"] = Exp
	i.StdLibFunctions["log"] = Log
	i.StdLibFunctions["sqrt"] = Sqrt
	i.StdLibFunctions["length"] = Length
	i.StdLibFunctions["sub"] = Sub
	i.StdLibFunctions["gsub"] = Gsub
	i.StdLibFunctions["split"] = Split
	i.StdLibFunctions["toupper"] = ToUpper
	i.StdLibFunctions["tolower"] = ToLower
	i.StdLibFunctions["substr"] = Substr
	i.StdLibFunctions["index"] = Index
	i.StdLibFunctions["match"] = Match
	return i
}

func (i *Interpreter) Run(input string) {
	i.Input = input

	for _, block := range i.BeginBlocks {
		for _, stmt := range block.Statements {
			i.topLevelWrapperdoStatement(stmt)
			if i.WasFatalErrorHit {
				return
			}
		}
	}

	i.advanceInput()
	for i.InputPostion < len(i.Input) {
		for _, stmt := range i.Rules {
			i.topLevelWrapperdoStatement(stmt)
			if i.WasNextStatementHit {
				i.WasNextStatementHit = false
				i.resetStack()
				break
			}
			if i.WasFatalErrorHit {
				return
			}
		}
		i.advanceInput()
	}

	for _, block := range i.EndBlocks {
		for _, stmt := range block.Statements {
			i.topLevelWrapperdoStatement(stmt)
			if i.WasFatalErrorHit {
				return
			}
		}
	}
}

func (i *Interpreter) advanceInput() {
	if i.InputPostion >= len(i.Input) {
		return
	}
	i.Stack[0].LocalVariables["$0"] = i.doConcatenate(i.Stack[0].LocalVariables["$0"], &ast.StringLiteral{Value: string(i.Input[i.InputPostion])})
	i.InputPostion += 1
}

func (i *Interpreter) backtrackInput() {
	i.InputPostion -= 1
	length := len(i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value)
	i.Stack[0].LocalVariables["$0"] = &ast.StringLiteral{Value: i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value[:length-1]}
}

func (i *Interpreter) consumeInput() {
	i.Stack[0].LocalVariables["$0"] = &ast.StringLiteral{Value: ""}
}

func (i *Interpreter) resetStack() {
	i.Stack = []CallStackEntry{}
	i.Stack = append(i.Stack, CallStackEntry{})
	i.Stack[0].LocalVariables = make(map[string]ast.Expression)
	i.Stack[0].LocalVariables["$0"] = &ast.StringLiteral{Value: ""}
}

func (i *Interpreter) attemptArrayLookup(indicies []ast.Expression, variable ast.Expression) ast.Expression {
	if indicies == nil {
		switch variable.(type) {
		case *ast.StringLiteral:
			return variable
		case *ast.NumericLiteral:
			return variable
		case *ast.AssociativeArray:
			return variable
		default:
			panic("Unknown variable type")
		}
	} else {
		switch variable.(type) {
		case *ast.StringLiteral:
			panic("attempt to address scalar with index")
		case *ast.NumericLiteral:
			panic("attempt to address scalar with index")
		case *ast.AssociativeArray:
			val, ok := variable.(*ast.AssociativeArray).Array[i.transformArrayLookupExpression(indicies)]
			if ok {
				return val
			}
			return &ast.StringLiteral{Value: ""}
		default:
			panic("Unknown variable type")
		}
	}
}

func (i *Interpreter) transformArrayLookupExpression(indicies []ast.Expression) string {
	var idxs []string
	for _, x := range indicies {
		idxs = append(idxs, i.doExpression(x).String())
	}
	return strings.Join(idxs, ",")
}

func (i *Interpreter) lookupVar(varName ast.Expression) ast.Expression {
	var id string
	var index []ast.Expression
	switch varName.(type) {
	case *ast.Identifier:
		id = varName.(*ast.Identifier).Value
		index = nil
	case *ast.ArrayIndexExpression:
		id = varName.(*ast.ArrayIndexExpression).ArrayName
		index = varName.(*ast.ArrayIndexExpression).IndexList
	default:
		panic("Unexpected expression type in lookupVar")
	}
	val, ok := i.Stack[len(i.Stack)-1].LocalVariables[id]
	if ok {
		return i.attemptArrayLookup(index, val)
	}
	val, ok = i.GlobalVariables[id]
	if ok {
		return i.attemptArrayLookup(index, val)
	}
	return &ast.StringLiteral{Value: ""}
}

func (i *Interpreter) setVar(varName ast.Expression, value ast.Expression) {
	var id string
	var index []ast.Expression
	switch varName.(type) {
	case *ast.Identifier:
		id = varName.(*ast.Identifier).Value
		index = nil
	case *ast.ArrayIndexExpression:
		id = varName.(*ast.ArrayIndexExpression).ArrayName
		index = varName.(*ast.ArrayIndexExpression).IndexList
	default:
		panic("Unexpected expression type in lookupVar")
	}
	_, ok := i.Stack[len(i.Stack)-1].LocalVariables[id]
	if ok {
		if index == nil {
			i.Stack[len(i.Stack)-1].LocalVariables[id] = value
		} else {
		}
	} else {
		if index == nil {
			i.GlobalVariables[id] = value
		} else {
			m, ok := i.GlobalVariables[id]
			if ok {
				switch m.(type) {
				case *ast.AssociativeArray:
					m.(*ast.AssociativeArray).Array[i.transformArrayLookupExpression(index)] = value
				default:
					i.GlobalVariables[id] = &ast.AssociativeArray{Array: make(map[string]ast.Expression)}
					i.GlobalVariables[id].(*ast.AssociativeArray).Array[i.transformArrayLookupExpression(index)] = value
				}
			} else {
				i.GlobalVariables[id] = &ast.AssociativeArray{Array: make(map[string]ast.Expression)}
				i.GlobalVariables[id].(*ast.AssociativeArray).Array[i.transformArrayLookupExpression(index)] = value
			}
		}
	}
}

func (i *Interpreter) createLocalVar(varName string, value ast.Expression) {
	i.Stack[len(i.Stack)-1].LocalVariables[varName] = value
}

func (i *Interpreter) topLevelWrapperdoStatement(stmt ast.Statement) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	i.doStatement(stmt)
}

func (i *Interpreter) doStatement(stmt ast.Statement) {
	defer func() {
		if r := recover(); r != nil {
			msg := r.(string)
			if msg == "rewinding" {
				return
			} else if msg == "next" {
				i.WasNextStatementHit = true
			} else {
				i.WasNextStatementHit = false
				i.WasFatalErrorHit = true
				err_msg := fmt.Sprintf("Runtime Error on line %d: %s\n", i.curStatement.GetToken().LineNum, msg)
				io.WriteString(i.Output, err_msg)
				panic("rewinding")
			}
		}
	}()

	i.curStatement = stmt

	switch stmt.(type) {
	case *ast.ExpressionStatement:
		i.doExpressionList(stmt.(*ast.ExpressionStatement).Expressions)
	case *ast.PrintStatement:
		i.doPrintStatement(stmt.(*ast.PrintStatement))
	case *ast.ActionBlockStatement:
		i.doBlock(stmt.(*ast.ActionBlockStatement))
	case *ast.AssignStatement:
		i.doAssignStatement(stmt.(*ast.AssignStatement))
	case *ast.AssignAndModifyStatement:
		i.doAssignAndModifyStatement(stmt.(*ast.AssignAndModifyStatement))
	case *ast.IfStatement:
		i.doIfStatement(stmt.(*ast.IfStatement))
	case *ast.NextStatement:
		panic("next")
	case *ast.WhileStatement:
		i.doWhileStatement(stmt.(*ast.WhileStatement))
	case *ast.DoWhileStatement:
		i.doDoWhileStatement(stmt.(*ast.DoWhileStatement))
	case *ast.ForStatement:
		i.doForStatement(stmt.(*ast.ForStatement))
	case *ast.ForEachStatement:
		i.doForEachStatement(stmt.(*ast.ForEachStatement))
	case *ast.DeleteStatement:
		i.doDeleteStatement(stmt.(*ast.DeleteStatement))
	default:
		panic("Unexpected statement type")
	}
}

func (i *Interpreter) doBlock(block ast.Block) {
	shouldExecuteBlock := false
	switch block.(type) {
	case *ast.BeginStatement:
		shouldExecuteBlock = true
	case *ast.EndStatement:
		shouldExecuteBlock = true
	case *ast.ActionBlock:
		shouldExecuteBlock = true
	case *ast.ActionBlockStatement:
		shouldExecuteBlock = i.evaluateActionBlockConditon(block.(*ast.ActionBlockStatement))
	}
	if shouldExecuteBlock {
		i.Stack = append(i.Stack, CallStackEntry{LocalVariables: i.mostRecentRegexCaptureGroups})
		for _, stmt := range block.GetStatements() {
			i.doStatement(stmt)
		}
		i.Stack = i.Stack[:len(i.Stack)-1]
	}
}

func (i *Interpreter) evaluateActionBlockConditon(block *ast.ActionBlockStatement) bool {
	evaluatedExpr := i.doExpression(block.Conditon)
	return ExpressionToBool(evaluatedExpr)
}

func (i *Interpreter) doPrintStatement(stmt *ast.PrintStatement) {
	toBePrinted := i.doExpressionList(stmt.Expressions)
	var asStrings []string
	for _, expr := range toBePrinted {
		asStrings = append(asStrings, expr.String())
	}
	io.WriteString(i.Output, strings.Join(asStrings, " "))
	io.WriteString(i.Output, "\n")
}

func (i *Interpreter) doAssignStatement(stmt *ast.AssignStatement) {
	for idx, target := range stmt.Targets {
		i.setVar(target, i.doExpression(stmt.Values[idx]))
	}
}

func (i *Interpreter) doAssignAndModifyStatement(stmt *ast.AssignAndModifyStatement) {
	var newValue ast.Expression
	switch stmt.Operator.Type {
	case token.ASSIGNPLUS:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "+", Right: stmt.Value})
	case token.ASSIGNMINUS:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "-", Right: stmt.Value})
	case token.ASSIGNMULTIPLY:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "*", Right: stmt.Value})
	case token.ASSIGNDIVIDE:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "/", Right: stmt.Value})
	case token.ASSIGNMODULO:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "%", Right: stmt.Value})
	case token.ASSIGNEXPONENT:
		newValue = i.doExpression(&ast.InfixExpression{Left: stmt.Target, Operator: "^", Right: stmt.Value})
	default:
		panic("Unknown Operator.")
	}
	i.setVar(stmt.Target, newValue)
}

func (i *Interpreter) doIfStatement(stmt *ast.IfStatement) {
	shouldExecuteElse := true
	for idx, condition := range stmt.Conditions {
		if ExpressionToBool(i.doExpression(condition)) {
			shouldExecuteElse = false
			i.doBlock(stmt.Consequences[idx])
			break
		}
	}
	if shouldExecuteElse && stmt.Else != nil {
		i.doBlock(stmt.Else)
	}
}

func (i *Interpreter) doWhileStatement(stmt *ast.WhileStatement) {
	stmt.ShouldBreak = false
	for ExpressionToBool(i.doExpression(stmt.Condition)) && !stmt.ShouldBreak {
		stmt.ShouldContinue = false
		for _, st := range stmt.Block.Statements {
			switch st.(type) {
			case *ast.ContinueStatement:
				stmt.ShouldContinue = true
			case *ast.BreakStatement:
				stmt.ShouldBreak = true
			default:
				i.doStatement(st)
			}
			if stmt.ShouldBreak {
				break
			}
			if stmt.ShouldContinue {
				break
			}
		}
	}
}

func (i *Interpreter) doDoWhileStatement(stmt *ast.DoWhileStatement) {
	stmt.ShouldBreak = false
	initRun := true
	for initRun || (ExpressionToBool(i.doExpression(stmt.Condition)) && !stmt.ShouldBreak) {
		initRun = false
		stmt.ShouldContinue = false
		for _, st := range stmt.Block.Statements {
			switch st.(type) {
			case *ast.ContinueStatement:
				stmt.ShouldContinue = true
			case *ast.BreakStatement:
				stmt.ShouldBreak = true
			default:
				i.doStatement(st)
			}
			if stmt.ShouldBreak {
				break
			}
			if stmt.ShouldContinue {
				break
			}
		}
	}
}

func (i *Interpreter) doForStatement(stmt *ast.ForStatement) {
	stmt.ShouldBreak = false
	i.doStatement(stmt.Initialization)
	for ExpressionToBool(i.doExpression(stmt.Condition)) && !stmt.ShouldBreak {
		stmt.ShouldContinue = false
		for _, st := range stmt.Block.Statements {
			switch st.(type) {
			case *ast.ContinueStatement:
				stmt.ShouldContinue = true
			case *ast.BreakStatement:
				stmt.ShouldBreak = true
			default:
				i.doStatement(st)
			}
			if stmt.ShouldBreak {
				break
			}
			if stmt.ShouldContinue {
				break
			}
		}
		i.doStatement(stmt.Action)
	}
}

func (i *Interpreter) doForEachStatement(stmt *ast.ForEachStatement) {
	val, ok := i.Stack[len(i.Stack)-1].LocalVariables[stmt.Array.Value]
	if !ok {
		val, ok = i.GlobalVariables[stmt.Array.Value]
	}
	if !ok {
		panic("Attempt to foreach on non-existent array")
	}
	array, ok := val.(*ast.AssociativeArray)
	if !ok {
		panic("Attempt to foreach on scalar variable")
	}
	keys := []string{}
	for k := range array.Array {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		i.setVar(stmt.VarName, &ast.StringLiteral{Value: k})
		for _, st := range stmt.Block.Statements {
			switch st.(type) {
			case *ast.ContinueStatement:
				stmt.ShouldContinue = true
			case *ast.BreakStatement:
				stmt.ShouldBreak = true
			default:
				i.doStatement(st)
			}
			if stmt.ShouldBreak {
				break
			}
			if stmt.ShouldContinue {
				break
			}
		}
	}
}
func (i *Interpreter) doExpressionList(expressions []ast.Expression) []ast.Expression {
	var results []ast.Expression
	for _, expr := range expressions {
		results = append(results, i.doExpression(expr))
	}
	return results
}

func (i *Interpreter) doExpression(expr ast.Expression) ast.Expression {
	switch expr.(type) {
	case *ast.TernaryExpression:
		return i.doTernaryExpression(expr.(*ast.TernaryExpression))
	case *ast.PrefixExpression:
		return i.doPrefixExpression(expr.(*ast.PrefixExpression))
	case *ast.InfixExpression:
		return i.doInfixExpression(expr.(*ast.InfixExpression))
	case *ast.CallExpression:
		return i.doFunctionCall(expr.(*ast.CallExpression))
	case *ast.PostfixExpression:
		return i.doPostfixExpression(expr.(*ast.PostfixExpression))
	case *ast.Identifier:
		return i.lookupVar(expr)
	case *ast.ArrayIndexExpression:
		return i.lookupVar(expr)
	}
	return expr
}

func (i *Interpreter) doPrefixExpression(expression *ast.PrefixExpression) ast.Expression {

	switch expression.Operator {
	case "!":
		return invertBool(i.doExpression(expression.Right))
	case "-":
		return negate(i.doExpression(expression.Right))
	case "++":
		i.setVar(expression.Right, i.doExpression(&ast.InfixExpression{Left: expression.Right, Operator: "+", Right: &ast.NumericLiteral{Value: 1}}))
		return i.lookupVar(expression.Right)
	case "--":
		i.setVar(expression.Right, i.doExpression(&ast.InfixExpression{Left: expression.Right, Operator: "-", Right: &ast.NumericLiteral{Value: 1}}))
		return i.lookupVar(expression.Right)
	default:
		panic("Unknown prefix operator")
	}
}

func (i *Interpreter) doInfixExpression(expression *ast.InfixExpression) ast.Expression {
	left := i.doExpression(expression.Left)
	right := i.doExpression(expression.Right)
	switch expression.Operator {
	case ".":
		return i.doConcatenate(left, right)
	case "~":
		return i.doRegexMatch(left, right, false)
	case "!~":
		return invertBool(i.doRegexMatch(left, right, false))
	case "~$0":
		return i.doRegexMatch(left, right, true)
	case "+":
		return i.doAdd(left, right)
	case "-":
		return i.doMinus(left, right)
	case "*":
		return i.doMultiply(left, right)
	case "/":
		return i.doDivide(left, right)
	case "%":
		return i.doModulus(left, right)
	case "^":
		return i.doExponentiation(left, right)
	case "==":
		return i.doEquality(left, right)
	case "!=":
		return invertBool(i.doEquality(left, right))
	case "<":
		return i.doLessThan(left, right)
	case ">":
		return i.doGreaterThan(left, right)
	case "<=":
		return i.doLessThanEqualTo(left, right)
	case ">=":
		return i.doGreaterThanEqualTo(left, right)
	case "in":
		return i.doArrayMembership(left, right)
	case "&&":
		return i.doBooleanAnd(left, right)
	case "||":
		return i.doBooleanOr(left, right)
	default:
		panic("Unknown Operator!")
	}
}

func (i *Interpreter) doPostfixExpression(expr *ast.PostfixExpression) ast.Expression {
	variable := i.lookupVar(expr.Left)
	value := &ast.StringLiteral{}
	switch variable.(type) {
	case *ast.ArrayIndexExpression:
		panic("attempt to postfix array")
	default:
		value.Value = variable.String()
	}
	switch expr.Operator {
	case "++":
		i.setVar(expr.Left, i.doExpression(&ast.InfixExpression{Left: expr.Left, Operator: "+", Right: &ast.NumericLiteral{Value: 1}}))
		return value
	case "--":
		i.setVar(expr.Left, i.doExpression(&ast.InfixExpression{Left: expr.Left, Operator: "-", Right: &ast.NumericLiteral{Value: 1}}))
		return value
	default:
		panic("Unknown postfix operator!")
	}
}

func (i *Interpreter) doRegexMatch(left ast.Expression, right ast.Expression, isReadingFromInput bool) ast.Expression {
	i.mostRecentRegexCaptureGroups = make(map[string]ast.Expression)
	var str string
	var regex string
	if isReadingFromInput && len(i.Stack) == 1 {
		isReadingFromInput = true
	} else {
		isReadingFromInput = false
	}

	switch left.(type) {
	case *ast.Identifier:
		lookup := i.lookupVar(left)
		switch lookup.(type) {
		case *ast.ArrayIndexExpression:
			panic("attempt to postfix array")
		default:
			str = lookup.String()
		}
	case *ast.StringLiteral:
		str = (left.(*ast.StringLiteral).Value)
	default:
		panic("non-string match against regex")
	}

	switch right.(type) {
	case *ast.RegexLiteral:
		regex = right.(*ast.RegexLiteral).Value
	default:
		panic("non-regex match against string")
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		panic("invalid regex")
	}

	matches := re.FindStringSubmatch(str)
	if matches != nil {
		if isReadingFromInput {
			prevMatches := matches
			prevMatch := &matches[0]
			i.advanceInput()
			newMatches := re.FindStringSubmatch(i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value)
			newMatch := &newMatches[0]
			for *prevMatch != *newMatch {
				i.advanceInput()
				prevMatches = newMatches
				prevMatch = newMatch
				newMatches = re.FindStringSubmatch(i.Stack[0].LocalVariables["$0"].(*ast.StringLiteral).Value)
				newMatch = &newMatches[0]
			}

			i.backtrackInput()
			i.consumeInput()
			matches = prevMatches
		}
		i.mostRecentRegexCaptureGroups["$MATCHES"] = &ast.AssociativeArray{Array: make(map[string]ast.Expression)}
		matchesArray, _ := i.mostRecentRegexCaptureGroups["$MATCHES"].(*ast.AssociativeArray)
		for idx, match := range matches {
			stridx := "$" + strconv.Itoa(idx)
			i.mostRecentRegexCaptureGroups[stridx] = ast.NewLiteral(match)
			matchesArray.Array[stridx] = ast.NewLiteral(match)
		}
		return boolToExpression(true)
	}
	return boolToExpression(false)
}

func (i *Interpreter) doFunctionCall(call *ast.CallExpression) ast.Expression {
	evaluatedArgs := i.doExpressionList(call.Arguments)
	function, ok := i.StdLibFunctions[call.Function.String()]
	if ok {
		return function(i, evaluatedArgs)
	}
	udf, ok := i.UserDefinedFunctions[call.Function.String()]
	if !ok {
		panic("attempt to call non-existent function")
	}

	if len(udf.Parameters) != len(call.Arguments) {
		panic("incorrect number of arguments to function.")
	}
	i.Stack = append(i.Stack, CallStackEntry{isFunction: true, LocalVariables: make(map[string]ast.Expression)})
	for idx, param := range udf.Parameters {
		_, ok = i.GlobalVariables[param.Value]
		if !ok {
			i.createLocalVar(param.Value, call.Arguments[idx])
		}
	}

	for _, stmt := range udf.Body.Statements {
		switch stmt.(type) {
		case *ast.ReturnStatement:
			i.Stack = i.Stack[:len(i.Stack)-1]
			return i.doExpression(stmt.(*ast.ReturnStatement).Value)
		default:
			i.doStatement(stmt)
		}
	}
	i.Stack = i.Stack[:len(i.Stack)-1]
	return &ast.StringLiteral{Value: ""}
}

func convertLiteralForMathOp(expr ast.Expression) float64 {
	switch expr.(type) {
	case *ast.StringLiteral:
		return 0.0
	case *ast.NumericLiteral:
		return (expr.(*ast.NumericLiteral).Value)
	default:
		panic("error in math")
	}
}

func (i *Interpreter) doAdd(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) + convertLiteralForMathOp(right)}
}

func (i *Interpreter) doMinus(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) - convertLiteralForMathOp(right)}
}

func (i *Interpreter) doMultiply(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) * convertLiteralForMathOp(right)}
}

func (i *Interpreter) doDivide(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: convertLiteralForMathOp(left) / convertLiteralForMathOp(right)}
}

func (i *Interpreter) doModulus(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: math.Mod(convertLiteralForMathOp(left), convertLiteralForMathOp(right))}
}

func (i *Interpreter) doExponentiation(left ast.Expression, right ast.Expression) ast.Expression {
	return &ast.NumericLiteral{Value: math.Pow(convertLiteralForMathOp(left), convertLiteralForMathOp(right))}
}

func convertLiteralForStringOp(expr ast.Expression) string {
	switch expr.(type) {
	case *ast.StringLiteral:
		return expr.(*ast.StringLiteral).Value
	case *ast.NumericLiteral:
		return (expr.(*ast.NumericLiteral).String())
	default:
		panic("error in literal to string conversion")
	}
}

func (i *Interpreter) doConcatenate(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return &ast.StringLiteral{Value: lhs + rhs}
}

func boolToExpression(b bool) ast.Expression {
	if b {
		return &ast.StringLiteral{Value: "1"}
	} else {
		return &ast.StringLiteral{Value: "0"}
	}
}

func invertBool(expr ast.Expression) ast.Expression {
	switch expr.(type) {
	case *ast.StringLiteral:
		if expr.(*ast.StringLiteral).Value == "0" {
			return &ast.StringLiteral{Value: "1"}
		}
		return &ast.StringLiteral{Value: "0"}
	case *ast.NumericLiteral:
		if expr.(*ast.NumericLiteral).Value == 0.0 {
			return &ast.NumericLiteral{Value: 1.0}
		}
		return &ast.NumericLiteral{Value: 0.0}
	default:
		panic("error inverting expression!")
	}
}

func negate(expr ast.Expression) ast.Expression {
	switch expr.(type) {
	case *ast.StringLiteral:
		return &ast.NumericLiteral{Value: 0.0}
	case *ast.NumericLiteral:
		return &ast.NumericLiteral{Value: expr.(*ast.NumericLiteral).Value * -1.0}
	default:
		panic("error inverting expression!")
	}
}

func (i *Interpreter) doEquality(left ast.Expression, right ast.Expression) ast.Expression {
	lhs := convertLiteralForStringOp(left)
	rhs := convertLiteralForStringOp(right)
	return boolToExpression(lhs == rhs)
}

func convertLiteralForComparisonOp(expr ast.Expression) (float64, error) {
	switch expr.(type) {
	case *ast.StringLiteral:
		return 0.0, errors.New("failed to convert to float")
	case *ast.NumericLiteral:
		return (expr.(*ast.NumericLiteral).Value), nil
	default:
		panic("error in math")
	}
}

func (i *Interpreter) doGreaterThan(left ast.Expression, right ast.Expression) ast.Expression {
	lhs_float, lerr := convertLiteralForComparisonOp(left)
	rhs_float, rerr := convertLiteralForComparisonOp(right)
	if lerr == nil && rerr == nil {
		return boolToExpression(lhs_float > rhs_float)
	}

	lhs_str := convertLiteralForStringOp(left)
	rhs_str := convertLiteralForStringOp(right)
	return boolToExpression(lhs_str > rhs_str)
}

func (i *Interpreter) doGreaterThanEqualTo(left ast.Expression, right ast.Expression) ast.Expression {
	lhs_float, lerr := convertLiteralForComparisonOp(left)
	rhs_float, rerr := convertLiteralForComparisonOp(right)
	if lerr == nil && rerr == nil {
		return boolToExpression(lhs_float >= rhs_float)
	}

	lhs_str := convertLiteralForStringOp(left)
	rhs_str := convertLiteralForStringOp(right)
	return boolToExpression(lhs_str >= rhs_str)
}

func (i *Interpreter) doLessThan(left ast.Expression, right ast.Expression) ast.Expression {
	lhs_float, lerr := convertLiteralForComparisonOp(left)
	rhs_float, rerr := convertLiteralForComparisonOp(right)
	if lerr == nil && rerr == nil {
		return boolToExpression(lhs_float < rhs_float)
	}

	lhs_str := convertLiteralForStringOp(left)
	rhs_str := convertLiteralForStringOp(right)
	return boolToExpression(lhs_str < rhs_str)
}

func (i *Interpreter) doLessThanEqualTo(left ast.Expression, right ast.Expression) ast.Expression {
	lhs_float, lerr := convertLiteralForComparisonOp(left)
	rhs_float, rerr := convertLiteralForComparisonOp(right)
	if lerr == nil && rerr == nil {
		return boolToExpression(lhs_float <= rhs_float)
	}

	lhs_str := convertLiteralForStringOp(left)
	rhs_str := convertLiteralForStringOp(right)
	return boolToExpression(lhs_str <= rhs_str)
}

func ExpressionToBool(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.StringLiteral:
		if expr.(*ast.StringLiteral).String() == "0" {
			return false
		}
		return true
	case *ast.NumericLiteral:
		if (expr.(*ast.NumericLiteral).Value) == 0.0 {
			return false
		}
		return true
	case *ast.AssociativeArray:
		panic("Got Array in Scalar Context!")
	default:
		panic("Expected Bool expression!!!")
	}
}

func (i *Interpreter) doTernaryExpression(expr *ast.TernaryExpression) ast.Expression {
	if ExpressionToBool(i.doExpression(expr.Condition)) {
		return i.doExpression(expr.IfTrue)
	}
	return i.doExpression(expr.IfFalse)
}

func isKeyInExpression(key string, expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.AssociativeArray:
		_, ok := expr.(*ast.AssociativeArray).Array[key]
		return ok
	default:
		panic("attempt to test membership of non-array")
	}
}

func (i *Interpreter) doArrayMembership(left ast.Expression, right ast.Expression) ast.Expression {
	var key string
	switch left.(type) {
	case *ast.ArrayIndexExpression:
		key = i.transformArrayLookupExpression(left.(*ast.ArrayIndexExpression).IndexList)
	default:
		key = i.doExpression(left).String()
	}
	var m map[string]ast.Expression
	switch right.(type) {
	case *ast.AssociativeArray:
		m = right.(*ast.AssociativeArray).Array
	default:
		return boolToExpression(false)
	}
	_, ok := m[key]
	return boolToExpression(ok)
}

func (i *Interpreter) doBooleanAnd(left ast.Expression, right ast.Expression) ast.Expression {
	l := ExpressionToBool(left)
	r := ExpressionToBool(right)
	result := l && r
	return boolToExpression(result)
}

func (i *Interpreter) doBooleanOr(left ast.Expression, right ast.Expression) ast.Expression {
	l := ExpressionToBool(left)
	r := ExpressionToBool(right)
	result := l || r
	return boolToExpression(result)
}

func (i *Interpreter) doDeleteStatement(stmt *ast.DeleteStatement) {
	val, ok := i.Stack[len(i.Stack)-1].LocalVariables[stmt.ToDelete.ArrayName]
	if !ok {
		val, ok = i.GlobalVariables[stmt.ToDelete.ArrayName]
	}
	if !ok {
		panic("Attempt to delete on non-existent variable")
	}
	array, ok := val.(*ast.AssociativeArray)
	if !ok {
		panic("Attempt to delete on scalar variable")
	}
	delete(array.Array, i.transformArrayLookupExpression(stmt.ToDelete.IndexList))
}
