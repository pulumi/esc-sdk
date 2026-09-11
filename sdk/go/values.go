// Copyright 2026, Pulumi Corporation.  All rights reserved.

package esc_sdk

func plainValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		plain := make(map[string]any, len(value))
		for key, envelope := range value {
			plain[key] = plainValue(unwrapEnvelope(envelope))
		}
		return plain
	case []any:
		plain := make([]any, len(value))
		for i, envelope := range value {
			plain[i] = plainValue(unwrapEnvelope(envelope))
		}
		return plain
	default:
		return value
	}
}

func unwrapEnvelope(envelope any) any {
	if fields, ok := envelope.(map[string]any); ok {
		return fields["value"]
	}
	return envelope
}
