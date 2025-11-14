package interpreter

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func Printf(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) < 1 {
		panic("Incorrect number of arguments to printf")
	}

	var fmtString string
	switch args[0].(type) {
	case *ast.StringLiteral:
		fmtString = args[0].(*ast.StringLiteral).Value
	case *ast.NumericLiteral:
		panic("first argument to printf is not a string.")
	default:
		panic("first argument to printf is not a string.")
	}

	var sprintfArgs []any
	for index, arg := range args[1:] {
		switch arg.(type) {
		case *ast.StringLiteral:
			sprintfArgs = append(sprintfArgs, arg.(*ast.StringLiteral).Value)
		case *ast.NumericLiteral:
			sprintfArgs = append(sprintfArgs, arg.(*ast.NumericLiteral).Value)
		default:
			panic(fmt.Sprintf("argument %d to printf is not a literal.", index+1))
		}
	}
	result := fmt.Sprintf(fmtString, sprintfArgs...)
	return ast.NewLiteral(result)
}

func Sprintf(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) < 1 {
		panic("Incorrect number of arguments to function sprintf")
	}

	var fmtString string
	switch args[0].(type) {
	case *ast.StringLiteral:
		fmtString = args[0].(*ast.StringLiteral).Value
	case *ast.NumericLiteral:
		panic("first argument to function sprintf is not a string.")
	default:
		panic("first argument to function sprintf is not a string.")
	}

	var sprintfArgs []any
	for index, arg := range args[1:] {
		switch arg.(type) {
		case *ast.StringLiteral:
			sprintfArgs = append(sprintfArgs, arg.(*ast.StringLiteral).Value)
		case *ast.NumericLiteral:
			sprintfArgs = append(sprintfArgs, arg.(*ast.NumericLiteral).Value)
		default:
			panic(fmt.Sprintf("argument %d to function sprintf is not a literal.", index+1))
		}
	}
	result := fmt.Sprintf(fmtString, sprintfArgs...)
	return ast.NewLiteral(result)
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

// Math Functions

func Atan2(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 2 {
		panic("Incorrect number of arguments to function atan2")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function atan2 is not a scalar.")
	}

	switch args[1].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("second argument to function atan2 is not a scalar.")
	}

	y := convertLiteralForMathOp(args[0])
	x := convertLiteralForMathOp(args[1])
	answer := math.Atan2(y, x)
	return &ast.NumericLiteral{Value: answer}
}

func Cos(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect number of arguments to function cos")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function cos is not a scalar.")
	}

	x := convertLiteralForMathOp(args[0])
	answer := math.Cos(x)
	return &ast.NumericLiteral{Value: answer}
}

func Sin(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect number of arguments to function sin")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function sin is not a scalar.")
	}

	x := convertLiteralForMathOp(args[0])
	answer := math.Sin(x)
	return &ast.NumericLiteral{Value: answer}
}

func Exp(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect number of arguments to function exp")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function exp is not a scalar.")
	}

	x := convertLiteralForMathOp(args[0])
	answer := math.Exp(x)
	return &ast.NumericLiteral{Value: answer}
}

func Log(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect number of arguments to function log")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function log is not a scalar.")
	}

	x := convertLiteralForMathOp(args[0])
	answer := math.Log(x)
	return &ast.NumericLiteral{Value: answer}
}

func Sqrt(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect number of arguments to function sqrt")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function sqrt is not a scalar.")
	}

	x := convertLiteralForMathOp(args[0])
	answer := math.Sqrt(x)
	return &ast.NumericLiteral{Value: answer}
}

func Int(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 1 {
		panic("Incorrect number of arguments to function int")
	}
	switch args[0].(type) {
	case *ast.StringLiteral:
	case *ast.NumericLiteral:
	default:
		panic("first argument to function int is not a scalar.")
	}

	x := convertLiteralForMathOp(args[0])
	int_x := int(x)
	answer := float64(int_x)
	return &ast.NumericLiteral{Value: answer}
}

func Rand(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) != 0 {
		panic("Incorrect number of arguments to function rand")
	}

	r := rand.New(rand.NewSource(i.seed))
	return &ast.NumericLiteral{Value: r.Float64()}
}

func Srand(i *Interpreter, args []ast.Expression) ast.Expression {
	if len(args) > 1 {
		panic("Incorrect number of arguments to function srand")
	}

	prev := float64(i.seed)
	var newSeed int
	if len(args) == 1 {
		switch args[0].(type) {
		case *ast.StringLiteral:
		case *ast.NumericLiteral:
		default:
			panic("first argument to function int is not a scalar.")
		}
		newSeed = int(convertLiteralForMathOp(args[0]))
	} else {
		newSeed = int(time.Now().UnixNano())
	}
	i.seed = int64(newSeed)
	return &ast.NumericLiteral{Value: prev}
}
