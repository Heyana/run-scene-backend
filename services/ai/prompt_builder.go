package ai

import (
	"fmt"
	"strings"

	"go_wails_project_manager/models"
)

// BuildPrompt 构建AI Prompt
func BuildPrompt(userRequest string, metadata map[string]models.NodeMetadata) string {
	systemPrompt := getSystemPrompt()
	nodeList := buildNodeList(metadata)
	examples := getExamples()

	return fmt.Sprintf(`%s

可用节点列表：
%s

%s

用户需求：%s

请生成蓝图JSON。`, systemPrompt, nodeList, examples, userRequest)
}

// getSystemPrompt 获取系统提示词
func getSystemPrompt() string {
	return `你是专业的3D场景蓝图生成助手。

生成规则：
1. 节点从左到右排列，水平间距300px，垂直间距150px
2. event类型连接event端口，models类型连接models端口
3. 节点ID从1开始递增，连接ID也从1开始递增
4. 只返回JSON，不要有其他内容
5. 不要设置title字段，使用节点的默认标题
6. 需要在最前端的event节点前添加"开始运行"节点
7. 用户如果传入了name 那么你生成节点的 properties 也要有相应的名称
对于大部分节点 都叫做name
对于模型/获取选中模型 这个节点 叫做 widget_model0（数字根据模型个数自增），widget_model1
8. 如果用户 要求生成按钮或者弹窗 但是没给名称 你自己给一个合适的名称
9. 如果用户 传入了确切的模型名称 那么应该用 模型/获取选中模型 传入widget_model0模型名称 而不是搜索节点 来保证精准

JSON格式：
{
  "nodes": [
    {
      "id": 1,
      "type": "节点类型",
      "pos": [x, y],
      "properties": {}
    }
  ],
  "links": [
    [linkId, originNodeId, originSlot, targetNodeId, targetSlot, "type"]
  ]
}`
}

// buildNodeList 构建节点列表
func buildNodeList(metadata map[string]models.NodeMetadata) string {
	var lines []string

	// 按类别分组
	categories := make(map[string][]models.NodeMetadata)
	for _, node := range metadata {
		categories[node.Category] = append(categories[node.Category], node)
	}

	// 生成节点列表
	for category, nodes := range categories {
		lines = append(lines, fmt.Sprintf("\n## %s", category))
		for _, node := range nodes {
			inputs := []string{}
			for _, i := range node.Inputs {
				inputs = append(inputs, fmt.Sprintf("%s(%s)", i.Name, i.Type))
			}
			outputs := []string{}
			for _, o := range node.Outputs {
				outputs = append(outputs, fmt.Sprintf("%s(%s)", o.Name, o.Type))
			}

			line := fmt.Sprintf("- %s", node.Type)
			if len(inputs) > 0 {
				line += fmt.Sprintf("\n  输入: [%s]", strings.Join(inputs, ", "))
			}
			if len(outputs) > 0 {
				line += fmt.Sprintf("\n  输出: [%s]", strings.Join(outputs, ", "))
			}

			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// getExamples 获取示例
func getExamples() string {
	return `
示例1 - 点击按钮播放动画：
用户需求：创建一个点击按钮播放动画
生成结果：
{
  "nodes": [
    {"id": 1, "type": "UI/按钮", "pos": [0, 0], "properties": {"name": "播放"}},
    {"id": 2, "type": "时间轴/播放", "pos": [300, 0], "properties": {"name": "动画1"}}
  ],
  "links": [[1, 1, 0, 2, 0, "event"]]
}

示例2 - 模型显示隐藏：
用户需求：隐藏模型
生成结果：
{
  "nodes": [
    {"id": 1, "type": "模型/获取模型", "pos": [0, 0], "properties": {}},
    {"id": 2, "type": "模型/修改/修改显示隐藏", "pos": [300, 0], "properties": {}}
  ],
  "links": [[1, 1, 1, 2, 1, "models"]]
}

示例3 - 点击按钮切换材质：
用户需求：点击按钮在两个材质之间切换
生成结果：
{
  "nodes": [
    {"id": 1, "type": "UI/按钮", "pos": [0, 0], "properties": {"name": "切换材质"}},
    {"id": 2, "type": "节点工具/自动切换", "pos": [300, 0], "properties": {}},
    {"id": 3, "type": "模型/获取选中模型", "pos": [600, -100], "properties": {
	"widget_model0":"模型名称1","widget_model1":"模型名称2"}},
    {"id": 4, "type": "材质/修改", "pos": [900, -100], "properties": {"name": "材质1"}},
    {"id": 5, "type": "模型/获取选中模型", "pos": [600, 100], "properties": {}},
    {"id": 6, "type": "材质/修改", "pos": [900, 100], "properties": {"name": "材质2"}}
  ],
  "links": [
    [1, 1, 0, 2, 0, "event"],
    [2, 2, 0, 4, 1, "event"],
    [3, 2, 1, 6, 1, "event"],
    [4, 3, 0, 4, 0, "models"],
    [5, 5, 0, 6, 0, "models"]
  ]
}

注意：
- "节点工具/自动切换"用于在多个通路之间切换，每次触发会轮流执行不同的输出
- 材质修改节点的properties中可以设置"name"属性指定材质名称`
}
