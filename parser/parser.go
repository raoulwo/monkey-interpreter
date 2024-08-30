package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
)

var precedences = map[token.Type]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	currToken token.Token
	peekToken token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	parser := &Parser{l: l, errors: []string{}}

	parser.prefixParseFns = make(map[token.Type]prefixParseFn)
	parser.registerPrefix(token.IDENT, parser.parseIdentifier)
	parser.registerPrefix(token.INT, parser.parseIntegerLiteral)
	parser.registerPrefix(token.BANG, parser.parsePrefixExpression)
	parser.registerPrefix(token.MINUS, parser.parsePrefixExpression)
	parser.registerPrefix(token.TRUE, parser.parseBoolean)
	parser.registerPrefix(token.FALSE, parser.parseBoolean)
	parser.registerPrefix(token.LPAREN, parser.parseGroupedExpression)
	parser.registerPrefix(token.IF, parser.parseIfExpression)
	parser.registerPrefix(token.FUNCTION, parser.parseFunctionLiteral)
	parser.registerPrefix(token.STRING, parser.parseStringLiteral)
	parser.registerPrefix(token.LBRACKET, parser.parseArrayLiteral)
	parser.registerPrefix(token.LBRACE, parser.parseHashLiteral)

	parser.infixParseFns = make(map[token.Type]infixParseFn)
	parser.registerInfix(token.PLUS, parser.parseInfixExpression)
	parser.registerInfix(token.MINUS, parser.parseInfixExpression)
	parser.registerInfix(token.SLASH, parser.parseInfixExpression)
	parser.registerInfix(token.ASTERISK, parser.parseInfixExpression)
	parser.registerInfix(token.EQ, parser.parseInfixExpression)
	parser.registerInfix(token.NOT_EQ, parser.parseInfixExpression)
	parser.registerInfix(token.LT, parser.parseInfixExpression)
	parser.registerInfix(token.GT, parser.parseInfixExpression)
	parser.registerInfix(token.LPAREN, parser.parseCallExpression)
	parser.registerInfix(token.LBRACKET, parser.parseIndexExpression)

	parser.nextToken()
	parser.nextToken()

	return parser
}

func (parser *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	parser.prefixParseFns[tokenType] = fn
}

func (parser *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	parser.infixParseFns[tokenType] = fn
}

func (parser *Parser) Errors() []string {
	return parser.errors
}

func (parser *Parser) peekError(tokenType token.Type) {
	message := fmt.Sprintf("expected next token to be %s, got %s instead", tokenType, parser.peekToken.Type)
	parser.errors = append(parser.errors, message)
}

func (parser *Parser) noPrefixParseFnError(tokenType token.Type) {
	message := fmt.Sprintf("no prefix parse function for %s found", tokenType)
	parser.errors = append(parser.errors, message)
}

func (parser *Parser) currPrecedence() int {
	if precedence, ok := precedences[parser.currToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (parser *Parser) peekPrecedence() int {
	if precedence, ok := precedences[parser.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (parser *Parser) nextToken() {
	parser.currToken = parser.peekToken
	parser.peekToken = parser.l.NextToken()
}

func (parser *Parser) expectPeek(tokenType token.Type) bool {
	if parser.peekTokenIs(tokenType) {
		parser.nextToken()
		return true
	} else {
		parser.peekError(tokenType)
		return false
	}
}

func (parser *Parser) currTokenIs(tokenType token.Type) bool {
	return parser.currToken.Type == tokenType
}

func (parser *Parser) peekTokenIs(tokenType token.Type) bool {
	return parser.peekToken.Type == tokenType
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !parser.currTokenIs(token.EOF) {
		statement := parser.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		parser.nextToken()
	}

	return program
}

func (parser *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: parser.currToken}
	block.Statements = []ast.Statement{}

	parser.nextToken()

	for !parser.currTokenIs(token.RBRACE) && !parser.currTokenIs(token.EOF) {
		statement := parser.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		parser.nextToken()
	}

	return block
}

func (parser *Parser) parseStatement() ast.Statement {
	switch parser.currToken.Type {
	case token.LET:
		return parser.parseLetStatement()
	case token.RETURN:
		return parser.parseReturnStatement()
	default:
		return parser.parseExpressionStatement()
	}
}

func (parser *Parser) parseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: parser.currToken}

	if !parser.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{Token: parser.currToken, Value: parser.currToken.Literal}

	if !parser.expectPeek(token.ASSIGN) {
		return nil
	}

	parser.nextToken()

	statement.Value = parser.parseExpression(LOWEST)

	for parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: parser.currToken}

	parser.nextToken()

	statement.ReturnValue = parser.parseExpression(LOWEST)

	for parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: parser.currToken}

	statement.Expression = parser.parseExpression(LOWEST)

	if parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseExpression(precedence int) ast.Expression {
	prefixFn := parser.prefixParseFns[parser.currToken.Type]
	if prefixFn == nil {
		parser.noPrefixParseFnError(parser.currToken.Type)
		return nil
	}
	leftExp := prefixFn()

	for !parser.peekTokenIs(token.SEMICOLON) && precedence < parser.peekPrecedence() {
		infixFn := parser.infixParseFns[parser.peekToken.Type]
		if infixFn == nil {
			return leftExp
		}

		parser.nextToken()

		leftExp = infixFn(leftExp)
	}

	return leftExp
}

func (parser *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    parser.currToken,
		Operator: parser.currToken.Literal,
	}

	parser.nextToken()

	expression.Right = parser.parseExpression(PREFIX)

	return expression
}

func (parser *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    parser.currToken,
		Operator: parser.currToken.Literal,
		Left:     left,
	}

	precedence := parser.currPrecedence()
	parser.nextToken()
	expression.Right = parser.parseExpression(precedence)

	return expression
}

func (parser *Parser) parseGroupedExpression() ast.Expression {
	parser.nextToken()

	expression := parser.parseExpression(LOWEST)

	if !parser.expectPeek(token.RPAREN) {
		return nil
	}

	return expression
}

func (parser *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: parser.currToken}

	if !parser.expectPeek(token.LPAREN) {
		return nil
	}

	parser.nextToken()
	expression.Condition = parser.parseExpression(LOWEST)

	if !parser.expectPeek(token.RPAREN) {
		return nil
	}

	if !parser.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = parser.parseBlockStatement()

	if parser.peekTokenIs(token.ELSE) {
		parser.nextToken()

		if !parser.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = parser.parseBlockStatement()
	}

	return expression
}

func (parser *Parser) parseFunctionLiteral() ast.Expression {
	literal := &ast.FunctionLiteral{Token: parser.currToken}

	if !parser.expectPeek(token.LPAREN) {
		return nil
	}

	literal.Parameters = parser.parseFunctionParameters()

	if !parser.expectPeek(token.LBRACE) {
		return nil
	}

	literal.Body = parser.parseBlockStatement()

	return literal
}

func (parser *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if parser.peekTokenIs(token.RPAREN) {
		parser.nextToken()
		return identifiers
	}

	parser.nextToken()

	identifier := &ast.Identifier{Token: parser.currToken, Value: parser.currToken.Literal}
	identifiers = append(identifiers, identifier)

	for parser.peekTokenIs(token.COMMA) {
		parser.nextToken()
		parser.nextToken()
		identifier := &ast.Identifier{Token: parser.currToken, Value: parser.currToken.Literal}
		identifiers = append(identifiers, identifier)
	}

	if !parser.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (parser *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expression := &ast.CallExpression{Token: parser.currToken, Function: function}
	expression.Arguments = parser.parseExpressionList(token.RPAREN)
	return expression
}

func (parser *Parser) parseIndexExpression(array ast.Expression) ast.Expression {
	expression := &ast.IndexExpression{Token: parser.currToken, Left: array}

	parser.nextToken()
	expression.Index = parser.parseExpression(LOWEST)

	if !parser.expectPeek(token.RBRACKET) {
		return nil
	}

	return expression
}

func (parser *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: parser.currToken, Value: parser.currToken.Literal}
}

func (parser *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: parser.currToken}

	value, err := strconv.ParseInt(parser.currToken.Literal, 0, 64)
	if err != nil {
		message := fmt.Sprintf("could not parse %q as integer", parser.currToken.Literal)
		parser.errors = append(parser.errors, message)
		return nil
	}
	literal.Value = value

	return literal
}

func (parser *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: parser.currToken, Value: parser.currTokenIs(token.TRUE)}
}

func (parser *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: parser.currToken, Value: parser.currToken.Literal}
}

func (parser *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: parser.currToken}

	array.Elements = parser.parseExpressionList(token.RBRACKET)

	return array
}

func (parser *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := []ast.Expression{}

	if parser.peekTokenIs(end) {
		parser.nextToken()
		return list
	}

	parser.nextToken()
	list = append(list, parser.parseExpression(LOWEST))

	for parser.peekTokenIs(token.COMMA) {
		parser.nextToken()
		parser.nextToken()
		list = append(list, parser.parseExpression(LOWEST))
	}

	if !parser.expectPeek(end) {
		return nil
	}

	return list
}

func (parser *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: parser.currToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !parser.peekTokenIs(token.RBRACE) {
		parser.nextToken()
		key := parser.parseExpression(LOWEST)

		if !parser.expectPeek(token.COLON) {
			return nil
		}

		parser.nextToken()
		value := parser.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !parser.peekTokenIs(token.RBRACE) && !parser.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !parser.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}
