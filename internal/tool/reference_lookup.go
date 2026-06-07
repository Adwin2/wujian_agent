package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const referenceLookupToolName = "reference_lookup"

// ReferenceLookupInput identifies a safe reference topic.
type ReferenceLookupInput struct {
	Topic string `json:"topic"`
}

// ReferenceLookupOutput contains source-attributed reference text.
type ReferenceLookupOutput struct {
	Topic      string `json:"topic"`
	Content    string `json:"content"`
	Source     string `json:"source"`
	Disclaimer string `json:"disclaimer"`
}

// ReferenceLookup returns bounded Phase 1 medical reference snippets.
type ReferenceLookup struct {
	references map[string]ReferenceLookupOutput
}

var _ einotool.InvokableTool = (*ReferenceLookup)(nil)

// NewReferenceLookup creates a reference lookup tool.
func NewReferenceLookup() *ReferenceLookup {
	refs := map[string]ReferenceLookupOutput{
		"bmi_formula": {
			Topic:      "bmi_formula",
			Content:    "BMI = 体重(kg) / 身高(m)^2。青少年 BMI 计算值本身只是一个数值，解释时需要结合年龄和性别参考。",
			Source:     "standard_bmi_formula",
			Disclaimer: "不能替代医生诊断。",
		},
		"bmi_interpretation_limitations": {
			Topic:      "bmi_interpretation_limitations",
			Content:    "儿童和青少年 BMI 不能直接套用成人阈值；通常需要结合年龄、性别和权威百分位参考表解释。",
			Source:     "phase1_safety_reference",
			Disclaimer: "如需医学判断，请咨询专业医生或学校卫生/儿保机构。",
		},
		"growth_reference_limitations": {
			Topic:      "growth_reference_limitations",
			Content:    "Phase 2 尚未配置权威 WHO/CDC 或中国青少年生长曲线表，因此不会给出权威百分位或诊断性结论。",
			Source:     "phase2_reference_placeholder",
			Disclaimer: "生长发育评估需结合长期记录和专业评估。",
		},
		"adolescent_sleep": {
			Topic:      "adolescent_sleep",
			Content:    "多数青少年常见睡眠建议约为每晚 8-10 小时；长期不足可能与疲劳、注意力和情绪状态相关。",
			Source:     "phase2_adolescent_sleep_reference",
			Disclaimer: "睡眠建议需结合个体情况；持续疲劳或睡眠问题建议咨询专业人员。",
		},
		"youth_safety": {
			Topic:      "youth_safety",
			Content:    "面向未成年人的健康建议应避免诊断化、极端饮食或减重处方，并在高风险时转交专业人员。",
			Source:     "phase2_safety_reference",
			Disclaimer: "本系统不能替代医生、营养师或心理专业人员。",
		},
	}
	return &ReferenceLookup{references: refs}
}

// Lookup returns a safe reference snippet by topic.
func (t *ReferenceLookup) Lookup(_ context.Context, input ReferenceLookupInput) (*ReferenceLookupOutput, error) {
	topic := normalizeTopic(input.Topic)
	if topic == "" {
		return nil, fmt.Errorf("topic is required; allowed topics: %s", strings.Join(t.allowedTopics(), ", "))
	}

	output, ok := t.references[topic]
	if !ok {
		return nil, fmt.Errorf("unsupported reference topic %q; allowed topics: %s", input.Topic, strings.Join(t.allowedTopics(), ", "))
	}
	return &output, nil
}

// Info returns the Eino tool definition for reference lookup.
func (t *ReferenceLookup) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: referenceLookupToolName,
		Desc: "Look up bounded Phase 1 reference information such as BMI formula and interpretation limitations. Use this instead of inventing medical reference facts.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"topic": {
				Type:     schema.String,
				Desc:     "Reference topic to retrieve.",
				Enum:     t.allowedTopics(),
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun executes the reference lookup tool from JSON arguments.
func (t *ReferenceLookup) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input ReferenceLookupInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", referenceLookupToolName, err)
	}

	output, err := t.Lookup(ctx, input)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", referenceLookupToolName, err)
	}
	return string(data), nil
}

func (t *ReferenceLookup) allowedTopics() []string {
	topics := make([]string, 0, len(t.references))
	for topic := range t.references {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}

func normalizeTopic(topic string) string {
	switch strings.ToLower(strings.TrimSpace(topic)) {
	case "bmi", "bmi公式", "bmi_formula":
		return "bmi_formula"
	case "bmi_interpretation", "bmi_interpretation_limitations", "bmi解读":
		return "bmi_interpretation_limitations"
	case "growth", "growth_curve", "growth_reference", "growth_reference_limitations", "生长曲线":
		return "growth_reference_limitations"
	case "sleep", "adolescent_sleep", "睡眠", "青少年睡眠":
		return "adolescent_sleep"
	case "safety", "youth_safety", "安全", "未成年人安全":
		return "youth_safety"
	default:
		return strings.ToLower(strings.TrimSpace(topic))
	}
}
