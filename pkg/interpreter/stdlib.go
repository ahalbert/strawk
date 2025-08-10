package interpreter

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ahalbert/strawk/pkg/ast"
)

func Length(i *Interpreter, args []ast.Expression) ast.Expression {

	if len(args) != 1 {
		panic("Incorrect arguments to function length")
	}

	var ret float64
	switch args[0].(type) {
	case *ast.StringLiteral:
		arg := args[0].(*ast.StringLiteral).Value
		ret = float64(len(arg))
	case *ast.NumericLiteral:
		arg := args[0].(*ast.NumericLiteral).String()
		ret = float64(len(arg))
	case *ast.AssociativeArray:
		ret = float64(len(args[0].(*ast.AssociativeArray).Array))
	default:
		panic("Incorrect argument type to function length")
	}
	return &ast.NumericLiteral{Value: ret}
}

func Sub(i *Interpreter, args []ast.Expression) ast.Expression {
	var in ast.Expression
	if len(args) < 2 || len(args) > 3 {
		panic("Incorrect arguments to function sub")
	}
	if len(args) == 2 {
		in = i.lookupVar(&ast.Identifier{Value: "$0"})
	} else {
		in = args[2]
	}

	switch args[0].(type) {
	case *ast.RegexLiteral:
	default:
		panic("first argument to function sub is not a regex")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function sub is not a scalar")
	}

	switch in.(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("third argument to function sub is not a scalar")
	}

	re, err := regexp.Compile(args[0].(*ast.RegexLiteral).Value)
	if err != nil {
		panic("First argument to sub not a valid regex")
	}
	found := re.FindString(in.String())
	replaced := in.String()
	if found != "" {
		replaced = strings.Replace(in.String(), found, args[1].String(), 1)
	}
	return ast.NewLiteral(replaced)
}

func Gsub(i *Interpreter, args []ast.Expression) ast.Expression {
	var in ast.Expression
	if len(args) < 2 || len(args) > 3 {
		panic("Incorrect arguments to function sub")
	}
	if len(args) == 2 {
		in = i.lookupVar(&ast.Identifier{Value: "$0"})
	} else {
		in = args[2]
	}

	switch args[0].(type) {
	case *ast.RegexLiteral:
	default:
		panic("first argument to function sub is not a regex")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function sub is not a scalar")
	}

	switch in.(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("third argument to function sub is not a scalar")
	}

	re, err := regexp.Compile(args[0].(*ast.RegexLiteral).Value)
	if err != nil {
		panic("First argument to sub not a valid regex")
	}

	replaced := re.ReplaceAllString(in.String(), args[1].String())
	return ast.NewLiteral(replaced)
}

func Split(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 2 {
		panic("Incorrect arguments to function split")
	}

	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function sub is not a regex")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function sub is not a scalar")
	}
	splits := strings.Split(args[0].String(), args[1].String())
	ret := make(map[string]ast.Expression)
	for idx, split := range splits {
		ret[strconv.Itoa(idx+1)] = &ast.StringLiteral{Value: split}
	}
	return &ast.AssociativeArray{Array: ret}
}

func ToLower(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect arguments to function split")
	}

	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function tolower is not a scalar.")
	}
	ret := strings.ToLower(args[0].String())
	return ast.NewLiteral(ret)
}

func ToUpper(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect arguments to function toupper")
	}

	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function toupper is not a scalar.")
	}
	ret := strings.ToUpper(args[0].String())
	return ast.NewLiteral(ret)
}

func Substr(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) < 2 || len(args) > 3 {
		panic("Incorrect number of arguments to function substr")
	}

	var s string
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function substr is not a scalar.")
	}
	s = args[0].String()

	var m int
	switch args[1].(type) {
	case *ast.StringLiteral:
		val, err := strconv.Atoi(args[1].String())
		if err != nil {
			panic("second argument to function substr is not an integer.")
		}
		m = val
	case *ast.NumericLiteral:
		val := args[1].(*ast.NumericLiteral).Value
		if val == float64(int(val)) {
			m = int(val)
		} else {
			panic("second argument to function substr is not an integer.")
		}
	default:
		panic("second argument to function substr is not a scalar.")
	}

	var n int
	if len(args) == 2 {
		n = -1
	} else {
		switch args[2].(type) {
		case *ast.StringLiteral:
			val, err := strconv.Atoi(args[2].String())
			if err != nil {
				panic("second argument to function substr is not an integer.")
			}
			n = val
		case *ast.NumericLiteral:
			val := args[2].(*ast.NumericLiteral).Value
			if val == float64(int(val)) {
				n = int(val)
			} else {
				panic("second argument to function substr is not an integer.")
			}
		default:
			panic("second argument to function substr is not a scalar.")
		}
	}

	if m >= len(s) {
		return ast.NewLiteral("")
	}
	if m+n >= len(s) || n == -1 {
		return ast.NewLiteral(s[m:])
	}

	return ast.NewLiteral(s[m : m+n])
}

func Index(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 2 {
		panic("Incorrect number of arguments to function index")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function index is not a scalar.")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function index is not a scalar.")
	}
	ret := strings.Index(args[0].String(), args[1].String())
	return ast.NewLiteral(strconv.Itoa(ret))
}

func Match(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 2 {
		panic("Incorrect number of arguments to function match")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function match is not a scalar.")
	}

	switch args[1].(type) {
	case *ast.RegexLiteral:
	default:
		panic("second argument to function match is not a regex")
	}

	re, err := regexp.Compile(args[0].(*ast.RegexLiteral).Value)
	if err != nil {
		panic("Second argument to function match not a valid regex")
	}
	loc := re.FindStringIndex(args[0].String())
	if loc == nil {
		return ast.NewLiteral(strconv.Itoa(-1))
	}
	return ast.NewLiteral(strconv.Itoa(loc[0]))
}
