package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
)

const guardrailBodyImageMessage = "我不能提供未成年人节食、极端减重或体重羞辱建议。可以一起关注规律饮食、充足睡眠、适量运动和必要时咨询医生或营养师。"

type guardrailDecision struct {
	Blocked bool
	Reason  string
	Message string
}

func evaluateGuardrail(message string) guardrailDecision {
	text := strings.TrimSpace(message)
	if text == "" {
		return guardrailDecision{}
	}

	// Body image / extreme weight loss
	if containsGuardrailAny(text, []string{"快速瘦", "暴瘦", "绝食", "催吐", "一天不吃", "极端减肥", "厌食", "体像", "身材焦虑", "太胖了不想吃", "瘦成闪电", "一个月瘦20斤", "断食减肥"}) {
		return guardrailDecision{Blocked: true, Reason: "minor body-image or extreme weight-loss advice", Message: guardrailBodyImageMessage}
	}

	// Age-inappropriate content
	if containsGuardrailAny(text, []string{"色情", "性暗示", "成人内容", "约炮", "裸聊", "援交"}) {
		return guardrailDecision{Blocked: true, Reason: "age-inappropriate content", Message: "我不能提供不适合未成年人的内容。可以继续讨论健康、睡眠、饮食、运动或情绪支持相关问题。"}
	}

	// Substance abuse
	if containsGuardrailAny(text, []string{"吸毒", "嗑药", "毒品", "大麻", "冰毒", "海洛因", "摇头丸", "K粉", "可卡因", "止咳水滥用", "镇定剂滥用", "安眠药滥用", "滥用安眠药", "滥用镇定剂", "滥用止咳水", "笑气", "一氧化二氮", "怎么买到药", "哪里买药", "处方药滥用", "药物依赖", "药物成瘾"}) {
		return guardrailDecision{Blocked: true, Reason: "substance abuse or illegal drug advice", Message: "我不能提供任何关于毒品或药物滥用的建议。如果你或身边的人正在经历药物困扰，请立即联系家长、老师、医生或拨打当地禁毒热线寻求帮助。"}
	}

	// Violence / self-harm (immediate danger triggers screening pipeline, but guardrail adds extra layer)
	if containsGuardrailAny(text, []string{"怎么杀人", "杀人方法", "制造炸弹", "制造武器", "恐怖袭击", "报复社会", "伤害别人", "暴力解决", "校园霸凌", "怎么打人", "虐待动物", "霸凌别人", "欺负同学"}) {
		return guardrailDecision{Blocked: true, Reason: "violence or harm-to-others content", Message: "我不能提供任何关于伤害他人或暴力的建议。如果你正在经历冲突或愤怒情绪，可以和信任的成年人聊聊，或拨打心理援助热线寻求帮助。"}
	}

	return guardrailDecision{}
}


func guardrailToolMiddleware() compose.ToolMiddleware {
	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
				if input != nil {
					decision := evaluateGuardrail(input.Arguments)
					if decision.Blocked {
						return nil, fmt.Errorf("guardrail blocked %s: %s", input.Name, decision.Reason)
					}
				}
				return next(ctx, input)
			}
		},
	}
}

func containsGuardrailAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
