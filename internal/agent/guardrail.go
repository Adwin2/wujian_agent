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
	if containsGuardrailAny(text, []string{"快速瘦", "暴瘦", "绝食", "催吐", "一天不吃", "极端减肥", "厌食", "体像", "身材焦虑", "太胖了不想吃"}) {
		return guardrailDecision{Blocked: true, Reason: "minor body-image or extreme weight-loss advice", Message: guardrailBodyImageMessage}
	}
	if containsGuardrailAny(text, []string{"色情", "性暗示", "成人内容", "约炮"}) {
		return guardrailDecision{Blocked: true, Reason: "age-inappropriate content", Message: "我不能提供不适合未成年人的内容。可以继续讨论健康、睡眠、饮食、运动或情绪支持相关问题。"}
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
