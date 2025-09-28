package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// 函数标识结构体
type Function struct {
	Name     string // 函数名
	Receiver string // 接收器类型（方法专属）
	Package  string // 所属包名
}

// 调用关系图
type CallGraph struct {
	Nodes map[Function]bool
	Edges map[Function]map[Function]bool
}

func NewCallGraph() *CallGraph {
	return &CallGraph{
		Nodes: make(map[Function]bool),
		Edges: make(map[Function]map[Function]bool),
	}
}

func (cg *CallGraph) AddNode(fn Function) {
	cg.Nodes[fn] = true
}

func (cg *CallGraph) AddEdge(from, to Function) {
	if cg.Edges[from] == nil {
		cg.Edges[from] = make(map[Function]bool)
	}
	cg.Edges[from][to] = true
}

// 主函数
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	graph := NewCallGraph()

	// 递归解析目录
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过测试文件和目录
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		processFile(path, graph)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// 生成 Graphviz DOT 格式
	generateDot(graph)
}

func processFile(path string, graph *CallGraph) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", path, err)
		return
	}

	// 当前包名
	pkg := file.Name.Name

	// 收集当前文件的所有函数声明
	funcDecls := make(map[string]*ast.FuncDecl)
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcName := fn.Name.Name
			funcDecls[funcName] = fn
		}
	}

	// 分析函数调用关系
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			analyzeFunction(fn, pkg, funcDecls, graph)
		}
	}
}

func analyzeFunction(fn *ast.FuncDecl, pkg string, funcDecls map[string]*ast.FuncDecl, graph *CallGraph) {
	// 创建当前函数节点
	fromNode := Function{
		Package: pkg,
		Name:    fn.Name.Name,
	}
	if fn.Recv != nil {
		// 处理接收器类型
		recv := fn.Recv.List[0].Type
		if star, ok := recv.(*ast.StarExpr); ok {
			fromNode.Receiver = fmt.Sprintf("*%s", exprToString(star.X))
		} else {
			fromNode.Receiver = exprToString(recv)
		}
	}
	graph.AddNode(fromNode)

	// 遍历函数体
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 解析被调用函数
		toNode := parseCallExpr(call, pkg, funcDecls)
		if toNode.Name != "" {
			graph.AddNode(toNode)
			graph.AddEdge(fromNode, toNode)
		}

		return true
	})
}

func parseCallExpr(call *ast.CallExpr, pkg string, funcDecls map[string]*ast.FuncDecl) Function {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		// 普通函数调用
		if _, exists := funcDecls[fn.Name]; exists {
			return Function{Package: pkg, Name: fn.Name}
		}
	case *ast.SelectorExpr:
		// 方法调用或包函数调用
		if ident, ok := fn.X.(*ast.Ident); ok {
			// 方法调用
			if _, exists := funcDecls[fn.Sel.Name]; exists {
				return Function{
					Package:  pkg,
					Receiver: ident.Name,
					Name:     fn.Sel.Name,
				}
			}
		}
	}
	return Function{}
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	default:
		return ""
	}
}

func generateDot(graph *CallGraph) {
	fmt.Println("digraph G {")
	fmt.Println("  node [shape=box, style=filled];")
	fmt.Println("  edge [arrowsize=0.8];")
	fmt.Println()

	// 输出节点
	for node := range graph.Nodes {
		label := node.Name
		if node.Receiver != "" {
			label = fmt.Sprintf("(%s).%s", node.Receiver, node.Name)
		}
		
		// 根据函数名首字母大小写设置不同颜色
		fillColor := "orange" // 默认蓝色（小写开头）
		if len(node.Name) > 0 && unicode.IsUpper(rune(node.Name[0])) {
			fillColor = "lightcoral" // 大写开头用红色
		}
		
		fmt.Printf("  \"%s\" [label=\"%s\", fillcolor=\"%s\"];\n", node, label, fillColor)
	}

	fmt.Println()

	// 输出边
	for from, tos := range graph.Edges {
		for to := range tos {
			fmt.Printf("  \"%s\" -> \"%s\";\n", from, to)
		}
	}

	fmt.Println("}")
}


