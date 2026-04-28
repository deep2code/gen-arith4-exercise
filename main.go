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

// TokenType 标记类型
type TokenType int

const (
	// 运算符类型
	TokenAdd TokenType = iota // +
	TokenSub                  // -
	TokenMul                  // ×
	TokenDiv                  // ÷

	// 表达式结构类型
	TokenOperand    // 操作数
	TokenResult     // 结果
	TokenLeftParen  // 左括号
	TokenRightParen // 右括号
	TokenEqual      // 等号
	TokenVar        // 变量
)

// IsOperator 判断是否为运算符类型
func (t TokenType) IsOperator() bool {
	return t >= TokenAdd && t <= TokenDiv
}

// String 返回标记的字符串表示
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
	case TokenOperand:
		return "Operand"
	case TokenResult:
		return "Result"
	case TokenLeftParen:
		return "❨"
	case TokenRightParen:
		return "❩"
	case TokenEqual:
		return "🟰"
	default:
		return "Unknown"
	}
}

// ArithmeticGenerator 四则运算生成器
type ArithmeticGenerator struct {
	rng *rand.Rand
}

// NewArithmeticGenerator 创建生成器实例
func NewArithmeticGenerator() *ArithmeticGenerator {
	source := rand.NewSource(time.Now().UnixNano())
	return &ArithmeticGenerator{
		rng: rand.New(source),
	}
}

// genOperand 生成 10-99 的随机正整数
func (g *ArithmeticGenerator) genOperand() int {
	return g.rng.Intn(maxOperand-minOperand+1) + minOperand
}

// calc 计算 a op b
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

// genValidOp 生成严格合法的运算符
// 保证：减法 a > b，除法能整除且结果>0
func (g *ArithmeticGenerator) genValidOp(a int, b int, allowDiv bool) TokenType {
	ops := []TokenType{TokenAdd, TokenSub, TokenMul, TokenDiv}
	g.rng.Shuffle(len(ops), func(i, j int) {
		ops[i], ops[j] = ops[j], ops[i]
	})

	for _, op := range ops {
		switch op {
		case TokenSub:
			if a > b { // 严格大于，避免结果为0
				return op
			}
		case TokenDiv:
			if allowDiv && b != 0 && a%b == 0 && (a/b) > 0 {
				return op
			}
		default: // 加法和乘法结果必然为正
			return op
		}
	}
	return TokenAdd // 兜底
}

// genThreeNoParen 生成结构: a op1 b op2 c（三个操作数，无括号）
func (g *ArithmeticGenerator) genThreeNoParen() Exercise {
	for {
		a := g.genOperand()
		b := g.genOperand()
		c := g.genOperand()

		op2 := g.genValidOp(b, c, true)
		temp2 := g.calc(b, op2, c)
		if temp2 <= 0 {
			continue
		}

		allowDiv := (a%temp2 == 0)
		op1 := g.genValidOp(a, temp2, allowDiv)
		result := g.calc(a, op1, temp2)

		// 结果必须≤maxResult
		if result > 0 && result <= maxResult {
			return Exercise{
				{Type: TokenOperand, Value: a},
				{Type: op1},
				{Type: TokenOperand, Value: b},
				{Type: op2},
				{Type: TokenOperand, Value: c},
				{Type: TokenEqual},
				{Type: TokenResult, Value: result},
			}
		}
	}
}

// genThreeParenFirst 生成结构: (a op1 b) op2 c（三个操作数，括号在前）
func (g *ArithmeticGenerator) genThreeParenFirst() Exercise {
	for {
		a := g.genOperand()
		b := g.genOperand()
		c := g.genOperand()

		op1 := g.genValidOp(a, b, true)
		temp1 := g.calc(a, op1, b)
		if temp1 <= 0 {
			continue
		}

		allowDiv := (temp1%c == 0)
		op2 := g.genValidOp(temp1, c, allowDiv)
		result := g.calc(temp1, op2, c)

		if result > 0 && result <= maxResult {
			return Exercise{
				{Type: TokenLeftParen},
				{Type: TokenOperand, Value: a},
				{Type: op1},
				{Type: TokenOperand, Value: b},
				{Type: TokenRightParen},
				{Type: op2},
				{Type: TokenOperand, Value: c},
				{Type: TokenEqual},
				{Type: TokenResult, Value: result},
			}
		}
	}
}

// genThreeParenLast 生成结构: a op1 (b op2 c)（三个操作数，括号在后）
func (g *ArithmeticGenerator) genThreeParenLast() Exercise {
	for {
		a := g.genOperand()
		b := g.genOperand()
		c := g.genOperand()

		op2 := g.genValidOp(b, c, true)
		temp2 := g.calc(b, op2, c)
		if temp2 <= 0 {
			continue
		}

		allowDiv := (a%temp2 == 0)
		op1 := g.genValidOp(a, temp2, allowDiv)
		result := g.calc(a, op1, temp2)

		// 新增：结果必须≤maxResult
		if result > 0 && result <= maxResult {
			return Exercise{
				{Type: TokenOperand, Value: a},
				{Type: op1},
				{Type: TokenLeftParen},
				{Type: TokenOperand, Value: b},
				{Type: op2},
				{Type: TokenOperand, Value: c},
				{Type: TokenRightParen},
				{Type: TokenEqual},
				{Type: TokenResult, Value: result},
			}
		}
	}
}

// genFourParenFirst 生成结构: (a op1 b) op2 c op3 d（四个操作数，括号在前）
func (g *ArithmeticGenerator) genFourParenFirst() Exercise {
	for {
		a := g.genOperand()
		b := g.genOperand()
		c := g.genOperand()
		d := g.genOperand()

		// 先计算 (a op1 b)
		op1 := g.genValidOp(a, b, true)
		temp1 := g.calc(a, op1, b)
		if temp1 <= 0 {
			continue
		}

		// 再计算 temp1 op2 c
		allowDiv2 := (temp1%c == 0)
		op2 := g.genValidOp(temp1, c, allowDiv2)
		temp2 := g.calc(temp1, op2, c)
		if temp2 <= 0 {
			continue
		}

		// 最后计算 temp2 op3 d
		allowDiv3 := (temp2%d == 0)
		op3 := g.genValidOp(temp2, d, allowDiv3)
		result := g.calc(temp2, op3, d)

		// 结果必须为正且≤maxResult
		if result > 0 && result <= maxResult {
			return Exercise{
				{Type: TokenLeftParen},
				{Type: TokenOperand, Value: a},
				{Type: op1},
				{Type: TokenOperand, Value: b},
				{Type: TokenRightParen},
				{Type: op2},
				{Type: TokenOperand, Value: c},
				{Type: op3},
				{Type: TokenOperand, Value: d},
				{Type: TokenEqual},
				{Type: TokenResult, Value: result},
			}
		}
	}
}

// genFourParenLast 生成结构: a op1 b op2 (c op3 d)（四个操作数，括号在后）
func (g *ArithmeticGenerator) genFourParenLast() Exercise {
	for {
		a := g.genOperand()
		b := g.genOperand()
		c := g.genOperand()
		d := g.genOperand()

		// 先计算 (c op3 d)
		op3 := g.genValidOp(c, d, true)
		temp3 := g.calc(c, op3, d)
		if temp3 <= 0 {
			continue
		}

		// 再计算 b op2 temp3
		allowDiv2 := (b%temp3 == 0)
		op2 := g.genValidOp(b, temp3, allowDiv2)
		temp2 := g.calc(b, op2, temp3)
		if temp2 <= 0 {
			continue
		}

		// 最后计算 a op1 temp2
		allowDiv1 := (a%temp2 == 0)
		op1 := g.genValidOp(a, temp2, allowDiv1)
		result := g.calc(a, op1, temp2)

		// 结果必须为正且≤maxResult
		if result > 0 && result <= maxResult {
			return Exercise{
				{Type: TokenOperand, Value: a},
				{Type: op1},
				{Type: TokenOperand, Value: b},
				{Type: op2},
				{Type: TokenLeftParen},
				{Type: TokenOperand, Value: c},
				{Type: op3},
				{Type: TokenOperand, Value: d},
				{Type: TokenRightParen},
				{Type: TokenEqual},
				{Type: TokenResult, Value: result},
			}
		}
	}
}

// genFourNoParen 生成结构: a op1 b op2 c op3 d（四个操作数，无括号）
func (g *ArithmeticGenerator) genFourNoParen() Exercise {
	for {
		a := g.genOperand()
		b := g.genOperand()
		c := g.genOperand()
		d := g.genOperand()

		// 先计算 c op3 d
		op3 := g.genValidOp(c, d, true)
		temp3 := g.calc(c, op3, d)
		if temp3 <= 0 {
			continue
		}

		// 再计算 b op2 temp3
		allowDiv2 := (b%temp3 == 0)
		op2 := g.genValidOp(b, temp3, allowDiv2)
		temp2 := g.calc(b, op2, temp3)
		if temp2 <= 0 {
			continue
		}

		// 最后计算 a op1 temp2
		allowDiv1 := (a%temp2 == 0)
		op1 := g.genValidOp(a, temp2, allowDiv1)
		result := g.calc(a, op1, temp2)

		// 结果必须为正且≤maxResult
		if result > 0 && result <= maxResult {
			return Exercise{
				{Type: TokenOperand, Value: a},
				{Type: op1},
				{Type: TokenOperand, Value: b},
				{Type: op2},
				{Type: TokenOperand, Value: c},
				{Type: op3},
				{Type: TokenOperand, Value: d},
				{Type: TokenEqual},
				{Type: TokenResult, Value: result},
			}
		}
	}
}

// Token 表达式标记
type Token struct {
	Type  TokenType
	Value int // 对于操作数和结果存储数值，对于运算符存储运算符类型本身(通过Type字段识别)
}

// Exercise 练习题结构 - 使用标记数组表示，最后一项一定是结果
type Exercise []Token

// Generate 生成指定数量的练习题
func (g *ArithmeticGenerator) Generate(count int) []Exercise {
	structures := []func() Exercise{
		g.genThreeNoParen, g.genThreeParenFirst, g.genThreeParenLast,
		g.genFourNoParen, g.genFourParenFirst, g.genFourParenLast}
	exercises := make([]Exercise, 0, count)

	for range count {
		idx := g.rng.Intn(len(structures))
		exercises = append(exercises, structures[idx]())
	}

	return exercises
}

// String 将 Exercise 转换为字符串表示，随机将一个操作数或结果替换为空白
func (e Exercise) String(rng *rand.Rand) string {
	var parts []string

	// 统计操作数和结果的数量
	numbers := make([]int, 0)
	for i, token := range e {
		if token.Type == TokenOperand || token.Type == TokenResult {
			numbers = append(numbers, i)
		}
	}

	// 随机选择一个位置替换为变量
	if len(numbers) > 0 {
		randIndex := rng.Intn(len(numbers))
		replaceIdx := numbers[randIndex]
		e[replaceIdx].Type = TokenVar
	}

	// 构建字符串 - 减少空格使表达式更紧凑
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

// ExerciseRow 表示表格中的一行（包含两个练习题）
type ExerciseRow struct {
	Expr1 string
	Expr2 string
}

// SaveToHTML 将练习题保存为HTML格式文件
func (g *ArithmeticGenerator) SaveToHTML(exercises []Exercise, filename string) error {
	// 定义HTML模板
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
        
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
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
        
        /* 打印优化 */
        @media print {
            body {
                margin: 0;
                padding: 10px;
            }
            
            table {
                page-break-inside: auto;
            }
            
            tr {
                page-break-inside: avoid;
                page-break-after: auto;
            }
            
            td {
                padding: 12px 8px;
                font-size: 18pt;
            }
        }
        
        /* 屏幕显示优化 */
        @media screen {
            td:hover {
                background-color: #f5f5f5;
            }
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

	// 准备数据：将练习题转换为行数据
	var rows []ExerciseRow
	for i := 0; i < len(exercises); i += 2 {
		row := ExerciseRow{}

		// 第一列
		e1 := exercises[i]
		row.Expr1 = e1.String(g.rng)

		// 第二列（如果存在）
		if i+1 < len(exercises) {
			e2 := exercises[i+1]
			row.Expr2 = e2.String(g.rng)
		}

		rows = append(rows, row)
	}

	// 创建输出文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 执行模板并写入文件
	return tmpl.Execute(file, rows)
}

func main() {
	count := flag.Int("count", 45*5+3, "练习题数量")
	filename := flag.String("filename", "", "输出文件名（默认：arth4-时间）")
	flag.Parse()

	// 如果未指定文件名，使用默认格式
	outputFile := *filename
	if outputFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		outputFile = fmt.Sprintf("arth4-%s.html", timestamp)
	}

	gen := NewArithmeticGenerator()

	// 生成指定数量的练习题
	exercises := gen.Generate(*count)

	// 根据格式保存到文件
	var err error
	err = gen.SaveToHTML(exercises, outputFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已生成 %d 道练习题，保存到 %s\n", *count, outputFile)
}
