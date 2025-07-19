package libconfig

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// TokenType represents different types of tokens.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenString
	TokenInteger
	TokenFloat
	TokenBoolean
	TokenAssign       // = or :
	TokenSemicolon    // ;
	TokenComma        // ,
	TokenLeftBrace    // {
	TokenRightBrace   // }
	TokenLeftBracket  // [
	TokenRightBracket // ]
	TokenLeftParen    // (
	TokenRightParen   // )
	TokenInclude      // @include
	TokenError
)

// Token represents a single token.
type Token struct {
	Value  string
	Type   TokenType
	Line   int
	Column int
}

// String returns a string representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("{%s: %q at %d:%d}", t.Type, t.Value, t.Line, t.Column)
}

// String returns a string representation of the token type.
func (tt TokenType) String() string {
	switch tt {
	case TokenEOF:
		return "EOF"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenString:
		return "STRING"
	case TokenInteger:
		return "INTEGER"
	case TokenFloat:
		return "FLOAT"
	case TokenBoolean:
		return "BOOLEAN"
	case TokenAssign:
		return "ASSIGN"
	case TokenSemicolon:
		return "SEMICOLON"
	case TokenComma:
		return "COMMA"
	case TokenLeftBrace:
		return "LEFT_BRACE"
	case TokenRightBrace:
		return "RIGHT_BRACE"
	case TokenLeftBracket:
		return "LEFT_BRACKET"
	case TokenRightBracket:
		return "RIGHT_BRACKET"
	case TokenLeftParen:
		return "LEFT_PAREN"
	case TokenRightParen:
		return "RIGHT_PAREN"
	case TokenInclude:
		return "INCLUDE"
	case TokenError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Lexer tokenizes libconfig input.
type Lexer struct {
	tokens   []Token
	input    string
	pos      int
	line     int
	column   int
	tokenPos int
	current  rune
}

// NewLexer creates a new lexer for the given input.
func NewLexer(reader io.Reader) *Lexer {
	// Read all input into memory for easier processing
	buf := strings.Builder{}
	if _, err := io.Copy(&buf, reader); err != nil {
		// Handle error gracefully by creating an empty lexer
		return &Lexer{
			input:  "",
			pos:    0,
			line:   1,
			column: 1,
			tokens: []Token{{Value: "", Type: TokenEOF, Line: 1, Column: 1}},
		}
	}

	input := buf.String()
	lexer := &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}

	if len(input) > 0 {
		lexer.current = rune(input[0])
	}

	// Tokenize the entire input
	lexer.tokenize()

	return lexer
}

// advance moves to the next character.
func (l *Lexer) advance() {
	if l.pos >= len(l.input)-1 {
		l.current = 0 // EOF
		return
	}

	if l.current == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	l.pos++
	l.current = rune(l.input[l.pos])
}

// peek returns the next character without advancing.
func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}

	return rune(l.input[l.pos+1])
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.current) {
		l.advance()
	}
}

// skipComment skips comments (C-style, C++-style, and script-style).
func (l *Lexer) skipComment() bool {
	if l.current == '/' {
		next := l.peek()
		if next == '/' {
			// C++-style comment: skip to end of line
			for l.current != '\n' && l.current != 0 {
				l.advance()
			}

			return true
		} else if next == '*' {
			// C-style comment: skip to */
			l.advance() // skip '/'
			l.advance() // skip '*'

			for l.current != 0 {
				if l.current == '*' && l.peek() == '/' {
					l.advance() // skip '*'
					l.advance() // skip '/'

					break
				}

				l.advance()
			}

			return true
		}
	} else if l.current == '#' {
		// Script-style comment: skip to end of line
		for l.current != '\n' && l.current != 0 {
			l.advance()
		}

		return true
	}

	return false
}

// readString reads a quoted string with escape sequence support.
func (l *Lexer) readString() string {
	var result strings.Builder

	l.advance() // skip opening quote

	for l.current != '"' && l.current != 0 {
		if l.current == '\\' {
			l.advance()

			switch l.current {
			case 'n':
				result.WriteRune('\n')
			case 'r':
				result.WriteRune('\r')
			case 't':
				result.WriteRune('\t')
			case 'b':
				result.WriteRune('\b')
			case 'f':
				result.WriteRune('\f')
			case 'a':
				result.WriteRune('\a')
			case 'v':
				result.WriteRune('\v')
			case '\\':
				result.WriteRune('\\')
			case '"':
				result.WriteRune('"')
			case 'x':
				// Hexadecimal escape \xNN
				l.advance()

				hex := ""

				for i := 0; i < 2 && l.current != 0; i++ {
					if (l.current >= '0' && l.current <= '9') ||
						(l.current >= 'a' && l.current <= 'f') ||
						(l.current >= 'A' && l.current <= 'F') {
						hex += string(l.current)
						l.advance()
					} else {
						break
					}
				}

				if len(hex) == 2 {
					if val, err := strconv.ParseInt(hex, 16, 8); err == nil {
						result.WriteRune(rune(val))
					}
				}

				continue
			default:
				result.WriteRune(l.current)
			}
		} else {
			result.WriteRune(l.current)
		}

		l.advance()
	}

	if l.current == '"' {
		l.advance() // skip closing quote
	}

	return result.String()
}

// readIdentifier reads an identifier.
func (l *Lexer) readIdentifier() string {
	var result strings.Builder

	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) ||
		l.current == '_' || l.current == '-' || l.current == '*' {
		result.WriteRune(l.current)
		l.advance()
	}

	return result.String()
}

// readNumber reads a number (integer or float).
func (l *Lexer) readNumber() (TokenType, string) {
	var result strings.Builder

	tokenType := TokenInteger

	// Handle different number prefixes
	if l.current == '0' {
		result.WriteRune(l.current)
		l.advance()

		switch l.current {
		case 'x', 'X':
			// Hexadecimal
			result.WriteRune(l.current)
			l.advance()

			for (l.current >= '0' && l.current <= '9') ||
				(l.current >= 'a' && l.current <= 'f') ||
				(l.current >= 'A' && l.current <= 'F') {
				result.WriteRune(l.current)
				l.advance()
			}
		case 'b', 'B':
			// Binary
			result.WriteRune(l.current)
			l.advance()

			for l.current == '0' || l.current == '1' {
				result.WriteRune(l.current)
				l.advance()
			}
		case 'o', 'O', 'q', 'Q':
			// Octal (new format)
			result.WriteRune(l.current)
			l.advance()

			for l.current >= '0' && l.current <= '7' {
				result.WriteRune(l.current)
				l.advance()
			}
		default:
			// Continue reading as decimal (might be a float starting with 0)
			for unicode.IsDigit(l.current) {
				result.WriteRune(l.current)
				l.advance()
			}
		}
	} else {
		// Regular decimal number
		for unicode.IsDigit(l.current) {
			result.WriteRune(l.current)
			l.advance()
		}
	}

	// Check for decimal point (float)
	if l.current == '.' && unicode.IsDigit(l.peek()) {
		tokenType = TokenFloat

		result.WriteRune(l.current)
		l.advance()

		for unicode.IsDigit(l.current) {
			result.WriteRune(l.current)
			l.advance()
		}
	}

	// Check for exponent (float)
	if l.current == 'e' || l.current == 'E' {
		tokenType = TokenFloat

		result.WriteRune(l.current)
		l.advance()

		if l.current == '+' || l.current == '-' {
			result.WriteRune(l.current)
			l.advance()
		}

		for unicode.IsDigit(l.current) {
			result.WriteRune(l.current)
			l.advance()
		}
	}

	// Check for long suffix
	if l.current == 'L' || l.current == 'l' {
		result.WriteRune(l.current)
		l.advance()
	}

	return tokenType, result.String()
}

// tokenize processes the entire input and creates tokens.
func (l *Lexer) tokenize() {
	for l.current != 0 {
		startLine := l.line
		startColumn := l.column

		l.skipWhitespace()

		if l.current == 0 {
			break
		}

		if l.skipComment() {
			continue
		}

		switch l.current {
		case '=', ':':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenAssign, Line: startLine, Column: startColumn})
			l.advance()
		case ';':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenSemicolon, Line: startLine, Column: startColumn})
			l.advance()
		case ',':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenComma, Line: startLine, Column: startColumn})
			l.advance()
		case '{':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenLeftBrace, Line: startLine, Column: startColumn})
			l.advance()
		case '}':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenRightBrace, Line: startLine, Column: startColumn})
			l.advance()
		case '[':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenLeftBracket, Line: startLine, Column: startColumn})
			l.advance()
		case ']':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenRightBracket, Line: startLine, Column: startColumn})
			l.advance()
		case '(':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenLeftParen, Line: startLine, Column: startColumn})
			l.advance()
		case ')':
			l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenRightParen, Line: startLine, Column: startColumn})
			l.advance()
		case '"':
			value := l.readString()
			l.tokens = append(l.tokens, Token{Value: value, Type: TokenString, Line: startLine, Column: startColumn})
		case '@':
			l.advance()

			if l.current == 'i' {
				ident := l.readIdentifier()
				if ident == "include" {
					l.tokens = append(l.tokens, Token{Value: "@include", Type: TokenInclude, Line: startLine, Column: startColumn})
				} else {
					l.tokens = append(l.tokens, Token{Value: "@" + ident, Type: TokenError, Line: startLine, Column: startColumn})
				}
			} else {
				l.tokens = append(l.tokens, Token{Value: "@", Type: TokenError, Line: startLine, Column: startColumn})
			}
		default:
			switch {
			case unicode.IsDigit(l.current) || (l.current == '-' && unicode.IsDigit(l.peek())):
				// Handle negative numbers
				sign := ""
				if l.current == '-' {
					sign = "-"

					l.advance()
				}

				tokenType, value := l.readNumber()
				l.tokens = append(l.tokens, Token{Value: sign + value, Type: tokenType, Line: startLine, Column: startColumn})
			case unicode.IsLetter(l.current) || l.current == '_' || l.current == '*':
				ident := l.readIdentifier()
				// Check for boolean values
				lower := strings.ToLower(ident)
				if lower == "true" || lower == "false" {
					l.tokens = append(l.tokens, Token{Value: lower, Type: TokenBoolean, Line: startLine, Column: startColumn})
				} else {
					l.tokens = append(l.tokens, Token{Value: ident, Type: TokenIdentifier, Line: startLine, Column: startColumn})
				}
			default:
				l.tokens = append(l.tokens, Token{Value: string(l.current), Type: TokenError, Line: startLine, Column: startColumn})
				l.advance()
			}
		}
	}

	l.tokens = append(l.tokens, Token{Value: "", Type: TokenEOF, Line: l.line, Column: l.column})
}

// NextToken returns the next token.
func (l *Lexer) NextToken() Token {
	if l.tokenPos >= len(l.tokens) {
		return Token{Value: "", Type: TokenEOF, Line: l.line, Column: l.column}
	}

	token := l.tokens[l.tokenPos]
	l.tokenPos++

	return token
}

// PeekToken returns the next token without consuming it.
func (l *Lexer) PeekToken() Token {
	if l.tokenPos >= len(l.tokens) {
		return Token{Value: "", Type: TokenEOF, Line: l.line, Column: l.column}
	}

	return l.tokens[l.tokenPos]
}
