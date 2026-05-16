package main

import (
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	maxResult  = 200
	minOperand = 1
	maxOperand = 99
)

type TokenType int

const (
	TokenAdd TokenType = iota
	TokenSub
	TokenMul
	TokenDiv
	TokenOperand
	TokenResult
	TokenLeftParen
	TokenRightParen
	TokenEqual
	TokenVar
)

func (t TokenType) String() string {
	switch t {
	case TokenAdd:
		return "➕"
	case TokenSub:
		return "➖"
	case TokenMul:
		return "X"
	case TokenDiv:
		return "➗"
	case TokenLeftParen:
		return "❨"
	case TokenRightParen:
		return "❩"
	case TokenEqual:
		return "🟰"
	default:
		return ""
	}
}

type ArithmeticGenerator struct {
	rng *rand.Rand
}

func NewArithmeticGenerator() *ArithmeticGenerator {
	source := rand.NewSource(time.Now().UnixNano())
	return &ArithmeticGenerator{rng: rand.New(source)}
}

func (g *ArithmeticGenerator) genOperand() int {
	return g.rng.Intn(maxOperand-minOperand+1) + minOperand
}

func (g *ArithmeticGenerator) calc(a int, op TokenType, b int) int {
	switch op {
	case TokenAdd:
		return a + b
	case TokenSub:
		return a - b
	case TokenMul:
		return a * b
	case TokenDiv:
		return a / b
	default:
		return 0
	}
}

func (g *ArithmeticGenerator) genValidOp(allowDiv bool) TokenType {
	ops := []TokenType{TokenAdd, TokenSub, TokenMul}
	if allowDiv {
		ops = append(ops, TokenDiv)
	}
	return ops[g.rng.Intn(len(ops))]
}

type ExprResult struct {
	tokens Exercise
	value  int
}

func getPrecedence(op TokenType) int {
	switch op {
	case TokenMul, TokenDiv:
		return 2
	case TokenAdd, TokenSub:
		return 1
	default:
		return 0
	}
}

func (e ExprResult) getRootOperator() TokenType {
	if len(e.tokens) == 1 {
		return TokenOperand
	}

	parenCount := 0
	for _, token := range e.tokens {
		if token.Type == TokenLeftParen {
			parenCount++
		} else if token.Type == TokenRightParen {
			parenCount--
		} else if parenCount == 0 && (token.Type == TokenAdd || token.Type == TokenSub || token.Type == TokenMul || token.Type == TokenDiv) {
			return token.Type
		}
	}
	return TokenOperand
}

func (e ExprResult) needsParenForOp(op TokenType, isRight bool) bool {
	rootOp := e.getRootOperator()
	if rootOp == TokenOperand {
		return false
	}

	opPrecedence := getPrecedence(op)
	rootPrecedence := getPrecedence(rootOp)

	if opPrecedence > rootPrecedence {
		return false
	}

	if opPrecedence < rootPrecedence {
		return true
	}

	if isRight && (op == TokenSub || op == TokenDiv) {
		return true
	}

	return false
}

func (e ExprResult) HasAddOrSub() bool {
	for _, token := range e.tokens {
		if token.Type == TokenAdd || token.Type == TokenSub {
			return true
		}
	}
	return false
}

func (g *ArithmeticGenerator) wrapWithParen(tokens Exercise, needsParen bool) Exercise {
	if needsParen && len(tokens) > 1 {
		return append(append(Exercise{{Type: TokenLeftParen}}, tokens...), Token{Type: TokenRightParen})
	}
	return tokens
}

func (g *ArithmeticGenerator) genExpr(depth int, maxDepth int) ExprResult {
	if depth >= maxDepth {
		val := g.genOperand()
		return ExprResult{
			tokens: Exercise{{Type: TokenOperand, Value: val}},
			value:  val,
		}
	}

	left := g.genExpr(depth+1, maxDepth)
	right := g.genExpr(depth+1, maxDepth)

	allowDiv := (left.value%right.value == 0)
	op := g.genValidOp(allowDiv)
	result := g.calc(left.value, op, right.value)

	if result <= 0 || result > maxResult {
		return g.genExpr(depth, maxDepth)
	}

	tokens := make(Exercise, 0)

	leftNeedsParen := left.needsParenForOp(op, false) || (len(left.tokens) > 1 && g.rng.Intn(3) == 0)
	rightNeedsParen := right.needsParenForOp(op, true) || (len(right.tokens) > 1 && g.rng.Intn(3) == 0)

	tokens = append(tokens, g.wrapWithParen(left.tokens, leftNeedsParen)...)
	tokens = append(tokens, Token{Type: op})
	tokens = append(tokens, g.wrapWithParen(right.tokens, rightNeedsParen)...)

	return ExprResult{
		tokens: tokens,
		value:  result,
	}
}

func (g *ArithmeticGenerator) genRecursiveExercise() Exercise {
	maxDepth := 2

	for {
		expr := g.genExpr(0, maxDepth)

		operandCount := 0
		for _, token := range expr.tokens {
			if token.Type == TokenOperand {
				operandCount++
			}
		}

		if operandCount >= 3 && operandCount <= 4 {
			result := make(Exercise, 0, len(expr.tokens)+2)
			result = append(result, expr.tokens...)
			result = append(result, Token{Type: TokenEqual})
			result = append(result, Token{Type: TokenResult, Value: expr.value})
			return result
		}
	}
}

func (g *ArithmeticGenerator) Generate(count int) []Exercise {
	exercises := make([]Exercise, 0, count)
	for range count {
		exercises = append(exercises, g.genRecursiveExercise())
	}
	return exercises
}

type Token struct {
	Type  TokenType
	Value int
}

type Exercise []Token

func (e Exercise) String(rng *rand.Rand) string {
	var parts []string

	indices := make([]int, 0)
	for i, token := range e {
		if token.Type == TokenOperand || token.Type == TokenResult {
			indices = append(indices, i)
		}
	}

	if len(indices) > 0 {
		randIndex := rng.Intn(len(indices))
		e[indices[randIndex]].Type = TokenVar
	}

	for _, token := range e {
		switch token.Type {
		case TokenAdd, TokenSub, TokenMul, TokenDiv, TokenEqual, TokenLeftParen, TokenRightParen:
			parts = append(parts, " "+token.Type.String()+" ")
		case TokenOperand, TokenResult:
			parts = append(parts, fmt.Sprintf("%d", token.Value))
		case TokenVar:
			parts = append(parts, "___")
		}
	}

	return strings.Join(parts, "")
}

type ExerciseRow struct {
	Expr1 string
	Expr2 string
}

func (g *ArithmeticGenerator) SaveToHTML(exercises []Exercise, filename string) error {
	tmpl := template.Must(template.New("exercises").Parse(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: "Microsoft YaHei", Arial, sans-serif;
            margin: 20px;
            padding: 0;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 0 auto;
        }
        td {
            border: 1px solid #ddd;
            padding: 15px 10px;
            text-align: center;
            font-size: 16px;
            vertical-align: middle;
        }
        @media print {
            body { margin: 0; padding: 10px; }
            table { page-break-inside: auto; }
            tr { page-break-inside: avoid; page-break-after: auto; }
            td { padding: 12px 8px; font-size: 18pt; }
        }
        @media screen {
            td:hover { background-color: #f5f5f5; }
        }
    </style>
</head>
<body>
    <table>
{{range .}}
        <tr>
            <td>{{.Expr1}}</td>
            <td>{{.Expr2}}</td>
        </tr>
{{end}}
    </table>
</body>
</html>`))

	var rows []ExerciseRow
	for i := 0; i < len(exercises); i += 2 {
		row := ExerciseRow{}
		e1 := exercises[i]
		row.Expr1 = e1.String(g.rng)
		if i+1 < len(exercises) {
			e2 := exercises[i+1]
			row.Expr2 = e2.String(g.rng)
		}
		rows = append(rows, row)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, rows)
}

func main() {
	count := flag.Int("count", 45*5+3, "练习题数量")
	filename := flag.String("filename", "", "输出文件名")
	flag.Parse()

	outputFile := *filename
	if outputFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		outputFile = fmt.Sprintf("arth4-%s.html", timestamp)
	}

	gen := NewArithmeticGenerator()
	exercises := gen.Generate(*count)
	err := gen.SaveToHTML(exercises, outputFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已生成 %d 道练习题，保存到 %s\n", *count, outputFile)
}
