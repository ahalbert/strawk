package parser

import (
	"fmt"
	"strconv"

	"github.com/ahalbert/strawk/pkg/ast"
	"github.com/ahalbert/strawk/pkg/lexer"
	"github.com/ahalbert/strawk/pkg/token"
)

const (
	_ int = iota
	LOWEST
	BOOLEANLOGIC // && or ||
	REGEXMATCH   // ~ or !~
	MEMBERSHIP   // expr in array
	TERNARY      // condition ? a : b
	EQUALITY     // ==
	CONCATENATE  // implied
	SUM          // +
	PRODUCT      // *, /, %
	EXPONENT     // ^
	PREFIX       // -X or !X
	INDEX        // []
	CALL         // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.OR:            BOOLEANLOGIC,
	token.AND:           BOOLEANLOGIC,
	token.NOTREGEXMATCH: REGEXMATCH,
	token.REGEXMATCH:    REGEXMATCH,
	token.IN:            MEMBERSHIP,
	token.TERNARY:       TERNARY,
	token.EQ:            EQUALITY,
	token.NOT_EQ:        EQUALITY,
	token.LT:            EQUALITY,
	token.GT:            EQUALITY,
	token.LTEQ:          EQUALITY,
	token.GTEQ:          EQUALITY,
	token.PLUS:          SUM,
	token.MINUS:         SUM,
	token.ASTERISK:      PRODUCT,
	token.SLASH:         PRODUCT,
	token.MODULO:        PRODUCT,
	token.EXPONENT:      EXPONENT,
	token.LBRACKET:      INDEX,
	token.LPAREN:        CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	Errors []string

	curToken       token.Token
	peekToken      token.Token
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		Errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifierExpr)
	p.registerPrefix(token.STRING, p.parseStringLiteralExpr)
	p.registerPrefix(token.NUMBER, p.parseNumericLiteralExpr)
	p.registerPrefix(token.SLASH, p.parseRegexExpression)
	p.registerPrefix(token.INCREMENT, p.parsePrefixExpression)
	p.registerPrefix(token.DECREMENT, p.parsePrefixExpression)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EXPONENT, p.parseInfixExpression)

	p.registerInfix(token.REGEXMATCH, p.parseInfixExpression)
	p.registerInfix(token.NOTREGEXMATCH, p.parseInfixExpression)

	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTEQ, p.parseInfixExpression)
	p.registerInfix(token.GTEQ, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseArrayMembershipExpression)

	p.registerInfix(token.TERNARY, p.parseTernaryExpression)
	p.registerInfix(token.LBRACKET, p.parseArrayIndexExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) addParseError(msg string) {
	error_msg := fmt.Sprintf("Parse Error on line %d: %s\n\n", p.curToken.LineNum, msg)
	p.Errors = append(p.Errors, error_msg)
	panic(msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(tokens ...token.TokenType) bool {
	for _, t := range tokens {
		if p.curToken.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) peekTokenIs(tokens ...token.TokenType) bool {
	for _, t := range tokens {
		if p.peekToken.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	defer func() {
		if r := recover(); r != nil {
			for !p.curTokenIs(token.NEWLINE, token.SEMICOLON, token.EOF) {
				p.nextToken()
			}
		}
	}()
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.BEGIN:
		return p.parseBeginStatement()
	case token.END:
		return p.parseEndStatement()
	case token.FUNCTION:
		return p.parseFunctionLiteral()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.DO:
		return p.parseDoWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.PRINT:
		return p.parsePrintStatement()
	case token.PRINTF:
		return p.parsePrintfStatement()
	case token.DELETE:
		return p.parseDeleteStatement()
	case token.PARSE:
		return p.parseParseStatement()
	case token.NEWLINE:
		p.nextToken()
		return nil
	case token.SEMICOLON:
		p.nextToken()
		return nil
	default:
		return p.parseExpressionPrefixedStatements()
	}
}

func (p *Parser) parseExpressionPrefixedStatements() ast.Statement {
	exprs := p.parseExpressionList(token.ASSIGN, token.LBRACE)

	switch p.curToken.Type {
	case token.ASSIGN:
		return p.parseAssignStatement(exprs)
	case token.ASSIGNPLUS:
		return p.parseAssignAndModifyStatement(exprs)
	case token.ASSIGNMINUS:
		return p.parseAssignAndModifyStatement(exprs)
	case token.ASSIGNMULTIPLY:
		return p.parseAssignAndModifyStatement(exprs)
	case token.ASSIGNDIVIDE:
		return p.parseAssignAndModifyStatement(exprs)
	case token.ASSIGNMODULO:
		return p.parseAssignAndModifyStatement(exprs)
	case token.ASSIGNEXPONENT:
		return p.parseAssignAndModifyStatement(exprs)
	case token.LBRACE:
		return p.parseActionBlockStatement(exprs)
	default:
		return &ast.ExpressionStatement{Token: p.curToken, Expressions: exprs}
	}
}

func (p *Parser) parseBeginStatement() *ast.BeginStatement {
	block := &ast.BeginStatement{Token: p.curToken}

	p.nextToken()
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.nextToken()

	return block
}

func (p *Parser) parseEndStatement() *ast.EndStatement {
	block := &ast.EndStatement{Token: p.curToken}

	p.nextToken()
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}
	p.nextToken()

	return block
}

func (p *Parser) parseAssignStatement(targets []ast.Expression) *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curToken}
	if !p.curTokenIs(token.ASSIGN) {
		return nil
	}

	//Convert previously parsed expressions to identifiers
	for _, expr := range targets {
		switch expr.(type) {
		case *ast.Identifier:
			stmt.Targets = append(stmt.Targets, expr)
		case *ast.ArrayIndexExpression:
			stmt.Targets = append(stmt.Targets, expr)
		default:
			p.addParseError("found non-identifier expression on lhs of assign statement")
		}
	}

	p.nextToken()

	stmt.Values = p.parseExpressionList(token.SEMICOLON)

	return stmt
}

func (p *Parser) parseAssignAndModifyStatement(targets []ast.Expression) *ast.AssignAndModifyStatement {
	if len(targets) != 1 {
		panic(p.curToken.Literal + " should have exactly 1 target")
	}

	operator := p.curToken
	target := targets[0]
	switch target.(type) {
	case *ast.Identifier:
	case *ast.ArrayIndexExpression:
	default:
		p.addParseError("found non-identifier expression on lhs of assign statement")
	}

	p.nextToken()

	return &ast.AssignAndModifyStatement{Operator: operator, Target: target, Value: p.parseExpression(LOWEST)}
}

func (p *Parser) parseActionBlockStatement(conditions []ast.Expression) *ast.ActionBlockStatement {

	if len(conditions) != 1 {
		p.addParseError("Action block should have exactly 1 condition")
	}

	stmt := &ast.ActionBlockStatement{Conditon: conditions[0]}

	//If a regex literal by itself, expand to $0 ~ /regex/
	switch stmt.Conditon.(type) {
	case *ast.RegexLiteral:
		stmt.Conditon = &ast.InfixExpression{Left: &ast.Identifier{Value: "$0"}, Operator: "~$0", Right: stmt.Conditon}
	}
	stmt.Statements = p.parseBlock()

	return stmt
}

func (p *Parser) parseBlock() *ast.ActionBlock {
	block := &ast.ActionBlock{}
	if !p.curTokenIs(token.LBRACE) {
		p.addParseError("Expected {")
	}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) {
		s := p.parseStatement()
		if s != nil {
			block.Statements = append(block.Statements, s)
		}
	}

	p.nextToken()
	return block
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	if !p.curTokenIs(token.IF) {
		p.addParseError("Expected if")
	}
	p.nextToken()
	condition := p.parseExpression(LOWEST)

	if !p.curTokenIs(token.LBRACE) {
		p.addParseError("Expected {")
	}
	consequence := p.parseBlock()
	stmt := &ast.IfStatement{}
	stmt.Conditions = append(stmt.Conditions, condition)
	stmt.Consequences = append(stmt.Consequences, consequence)
	// for p.curTokenIs(token.NEWLINE) {
	// 	p.nextToken()
	// }
	for p.curTokenIs(token.ELSE) && p.peekTokenIs(token.IF) {
		p.nextToken()
		p.nextToken()
		stmt.Conditions = append(stmt.Conditions, p.parseExpression(LOWEST))
		stmt.Consequences = append(stmt.Consequences, p.parseBlock())
	}
	if p.curTokenIs(token.ELSE) && p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		stmt.Else = p.parseBlock()
	}
	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	if !p.curTokenIs(token.WHILE) {
		p.addParseError("Expected while")
	}
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	loop := p.parseBlock()
	return &ast.WhileStatement{Condition: condition, Block: loop}
}

func (p *Parser) parseDoWhileStatement() *ast.DoWhileStatement {
	if !p.curTokenIs(token.DO) {
		p.addParseError("Expected do")
	}
	p.nextToken()
	loop := p.parseBlock()
	if !p.curTokenIs(token.WHILE) {
		p.addParseError("Expected while")
	}
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	return &ast.DoWhileStatement{Condition: condition, Block: loop}
}

func (p *Parser) parseForStatement() ast.Statement {
	t := p.curToken
	if !p.curTokenIs(token.FOR) {
		p.addParseError("Expected do")
	}
	p.nextToken()
	if !p.curTokenIs(token.LPAREN) {
		p.addParseError("Expected (")
	}
	p.nextToken()

	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.IN) {
		keyVariable := p.parseIdentifierExpr()
		p.nextToken()
		arrayName := p.parseIdentifierExpr()
		if !p.curTokenIs(token.RPAREN) {
			p.addParseError("Expected )")
		}
		p.nextToken()
		return &ast.ForEachStatement{Token: t,
			VarName: keyVariable.(*ast.Identifier),
			Array:   arrayName.(*ast.Identifier),
			Block:   p.parseBlock()}
	}

	init := p.parseStatement()
	if !p.curTokenIs(token.SEMICOLON) {
		p.addParseError("Expected ;")
	}
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if !p.curTokenIs(token.SEMICOLON) {
		p.addParseError("Expected ;")
	}
	p.nextToken()
	action := p.parseStatement()
	if !p.curTokenIs(token.RPAREN) {
		p.addParseError("Expected )")
	}
	p.nextToken()
	block := p.parseBlock()
	return &ast.ForStatement{
		Initialization: init,
		Condition:      condition,
		Action:         action,
		Block:          block,
	}
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	p.nextToken()
	return &ast.BreakStatement{}
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	p.nextToken()
	return &ast.ContinueStatement{}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	p.nextToken()
	return &ast.ReturnStatement{Value: p.parseExpression(LOWEST)}
}

func (p *Parser) parseNextStatement() *ast.NextStatement {
	p.nextToken()
	return &ast.NextStatement{}
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}
	p.nextToken()
	stmt.Expressions = p.parseExpressionList(token.SEMICOLON)
	return stmt
}

func (p *Parser) parsePrintfStatement() *ast.PrintfStatement {
	stmt := &ast.PrintfStatement{Token: p.curToken}
	p.nextToken()
	stmt.Expressions = p.parseExpressionList(token.SEMICOLON)
	return stmt
}

func (p *Parser) parseExpressionList(end ...token.TokenType) []ast.Expression {

	list := []ast.Expression{}

	if p.curTokenIs(end...) {
		p.nextToken()
		return nil
	}

	list = append(list, p.parseExpression(LOWEST))

	for p.curTokenIs(token.COMMA) {
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	return list
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.addParseError(fmt.Sprintf("no prefix parse function for %s found", p.curToken.Type))
		return nil
	}
	leftExp := prefix()

	if !p.curTokenIs(token.SEMICOLON, token.COMMA, token.ASSIGN, token.NEWLINE, token.ASSIGNPLUS, token.LBRACE, token.COLON, token.RBRACKET) && precedence < p.curPrecedence() {
		for !p.curTokenIs(token.SEMICOLON, token.COMMA, token.ASSIGN, token.NEWLINE, token.ASSIGNPLUS, token.LBRACE, token.COLON, token.RBRACKET) && precedence < p.curPrecedence() {
			infix := p.infixParseFns[p.curToken.Type]
			if infix == nil {
				return leftExp
			}
			leftExp = infix(leftExp)
		}
		return leftExp
	}

	_, pok := p.prefixParseFns[p.curToken.Type]
	if pok && precedence <= p.curPrecedence() {
		return p.parseConcatenateExpression(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIdentifierExpr() ast.Expression {
	defer p.nextToken()
	ident := &ast.Identifier{Value: p.curToken.Literal}
	if p.peekTokenIs(token.INCREMENT) {
		expr := &ast.PostfixExpression{Left: ident, Operator: p.peekToken.Literal}
		p.nextToken()
		return expr
	} else if p.peekTokenIs(token.DECREMENT) {
		expr := &ast.PostfixExpression{Left: ident, Operator: p.peekToken.Literal}
		p.nextToken()
		return expr
	}
	return ident
}

func (p *Parser) parseStringLiteralExpr() ast.Expression {
	lit := &ast.StringLiteral{Value: p.curToken.Literal}
	p.nextToken()
	return lit
}

func (p *Parser) parseNumericLiteralExpr() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addParseError("unparsable numeric type")
	}
	lit := &ast.NumericLiteral{Value: val}
	p.nextToken()
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Operator: p.curToken.Literal}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseConcatenateExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: ".",
		Left:     left,
	}

	precedence := CONCATENATE
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseRegexExpression() ast.Expression {
	if !p.curTokenIs(token.SLASH) {
		return nil
	}

	p.l.ExpectRegex = true
	var doubleBacktrack bool
	if p.curTokenIs(token.SLASH) && p.peekTokenIs(token.SLASH) {
		p.l.BacktrackToChar('/')
		p.l.BacktrackToChar('/')
		doubleBacktrack = true
	} else {
		p.l.BacktrackToChar('/')
	}
	p.nextToken()
	regex := p.peekToken.Literal
	p.l.ExpectRegex = false
	if doubleBacktrack {
		p.nextToken()
		p.nextToken()
	}

	for !p.curTokenIs(token.SLASH) {
		p.nextToken()
	}

	p.nextToken()

	return &ast.RegexLiteral{Value: regex}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	p.nextToken()
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	p.nextToken()
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exprs := p.parseExpressionList(token.RPAREN)
	var expr ast.Expression
	if len(exprs) > 1 {
		expr = &ast.ArrayIndexExpression{IndexList: exprs}
	} else {
		expr = exprs[0]
	}
	if !p.curTokenIs(token.RPAREN) {
		p.addParseError("expected (")
		// p.addError(fmt.Sprintf("expected ), got %s %s", p.curToken.Type, p.curToken.Literal))
	}
	p.nextToken()
	return expr
}

func (p *Parser) parseTernaryExpression(expr ast.Expression) ast.Expression {
	ternaryExpr := &ast.TernaryExpression{Condition: expr}
	p.nextToken()
	ternaryExpr.IfTrue = p.parseExpression(LOWEST)
	if !p.curTokenIs(token.COLON) {
		p.addParseError("expected :")
		// p.addError(fmt.Sprintf("expected ), got %s %s", p.curToken.Type, p.curToken.Literal))
	}
	p.nextToken()
	ternaryExpr.IfFalse = p.parseExpression(LOWEST)
	return ternaryExpr
}

func (p *Parser) parseArrayIndexExpression(expr ast.Expression) ast.Expression {
	var id string
	switch expr.(type) {
	case *ast.Identifier:
		id = expr.String()
	default:
		p.addParseError("Attempt to address array with non-identifier")
	}
	p.nextToken()
	indicies := p.parseExpressionList()
	arrayIndexExpression := &ast.ArrayIndexExpression{ArrayName: id, IndexList: indicies}

	if p.peekTokenIs(token.INCREMENT) {
		expr := &ast.PostfixExpression{Left: arrayIndexExpression, Operator: p.peekToken.Literal}
		p.nextToken()
		p.nextToken()
		return expr
	} else if p.peekTokenIs(token.DECREMENT) {
		expr := &ast.PostfixExpression{Left: arrayIndexExpression, Operator: p.peekToken.Literal}
		p.nextToken()
		p.nextToken()
		return expr
	}

	p.nextToken()
	return arrayIndexExpression
}

func (p *Parser) parseArrayMembershipExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	p.nextToken()
	if !p.curTokenIs(token.IDENT) {
		p.addParseError("key in array expression: expected identifier on the right.")
	}
	right := p.parseExpression(LOWEST)
	expr.Right = right
	return expr
}

func (p *Parser) parseFunctionLiteral() *ast.FunctionLiteral {
	function := &ast.FunctionLiteral{}
	if !p.curTokenIs(token.FUNCTION) {
		p.addParseError("expected function keyword")
	}
	p.nextToken()
	if !p.curTokenIs(token.IDENT) {
		p.addParseError("expected identifier for function name.")
	}
	function.Name = ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	if !p.curTokenIs(token.LPAREN) {
		p.addParseError("expected (")
	}
	for !p.curTokenIs(token.RPAREN) {
		p.nextToken()
		//In case no params are passed in
		if p.curTokenIs(token.RPAREN) {
			break
		}
		if !p.curTokenIs(token.IDENT) {
			p.addParseError("Expected identifier when parsing function params")
		}
		function.Parameters = append(function.Parameters, ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		p.nextToken()
		if !p.curTokenIs(token.COMMA) && !p.curTokenIs(token.RPAREN) {
			p.addParseError("Expected , or ) when parsing function params")
		}
	}
	p.nextToken()
	function.Body = p.parseBlock()
	return function
}

func (p *Parser) parseDeleteStatement() *ast.DeleteStatement {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	switch expr.(type) {
	case *ast.ArrayIndexExpression:
	default:
		p.addParseError("Expected Array Index Expression with delete statement")
	}
	return &ast.DeleteStatement{ToDelete: expr.(*ast.ArrayIndexExpression)}
}

func (p *Parser) parseParseStatement() *ast.ParseStatement {
	p.nextToken()
	var stmt *ast.ParseStatement
	stmt = &ast.ParseStatement{}
	if !p.curTokenIs(token.LBRACE) {
		expr := p.parseExpression(LOWEST)
		switch expr.(type) {
		case *ast.StringLiteral:
		case *ast.Identifier:
		default:
			p.addParseError("Expected string literal or identfier expression with parse statement")
		}
		stmt.ToParse = expr
	}

	if !p.curTokenIs(token.LBRACE) {
		p.addParseError("Expected block after parse statement.")
	}
	stmt.Actions = p.parseBlock()
	return stmt
}
