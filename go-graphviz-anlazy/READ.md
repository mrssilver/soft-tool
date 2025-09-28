
1. go build

2. 运行命令分析指定目录：

go run main.go /path/to/your/gocode > callgraph.dot


3. 使用 Graphviz 生成图像：

dot -Tpng -o callgraph.png callgraph.dot


功能特点

1. 递归目录扫描：自动处理所有子目录中的 Go 文件

2. 智能识别：

  ◦ 普通函数调用

  ◦ 结构体方法调用

  ◦ 指针接收器方法

3. 排除测试文件：自动忽略 _test.go 文件

4. 清晰可视化：

  ◦ 矩形节点表示函数

  ◦ 箭头表示调用关系

  ◦ 接收器类型显示在方法名前

输出示例 (DOT 格式)

digraph G {
  node [shape=box, style=filled, fillcolor=orange];
  edge [arrowsize=0.8];
  
  "main" [label="main"];
  "(Server).Start" [label="(Server).Start"];
  "(Server).Stop" [label="(Server).Stop"];
  
  "main" -> "(Server).Start";
  "main" -> "(Server).Stop";
  "(Server).Start" -> "Logger.Log";
}
（导出的函数）显示为红色节点，小写开头的函数显示为orange节点。
