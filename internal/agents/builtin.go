package agents

func BuiltinRoles() []Role {
	return []Role{
		{
			Name:         "strict-editor",
			DisplayName:  "Strict Editor",
			Model:        "openai/gpt-4o-mini",
			Stance:       "critique",
			Temperature:  0.4,
			MaxTokens:    1200,
			FocusAreas:   []string{"clarity", "line edits", "argument quality"},
			SystemPrompt: "You are a strict but pragmatic editor. Identify unclear, verbose, jumpy, or tonally inconsistent parts of the draft and provide actionable edits. When responding to prior agents, explicitly indicate agree, disagree, extend, or neutral.",
			Color:        "#d14d41",
		},
		{
			Name:         "casual-reader",
			DisplayName:  "Casual Reader",
			Model:        "openai/gpt-4o-mini",
			Stance:       "neutral",
			Temperature:  0.7,
			MaxTokens:    1000,
			FocusAreas:   []string{"reader reaction", "engagement", "confusion"},
			SystemPrompt: "You represent a casual reader. Explain what is engaging, distracting, or confusing in plain language, and give concrete feedback. When responding to prior agents, state your stance clearly.",
			Color:        "#2f7d5c",
		},
		{
			Name:         "structure-analyst",
			DisplayName:  "Structure Analyst",
			Model:        "openai/gpt-4o-mini",
			Stance:       "critique",
			Temperature:  0.5,
			MaxTokens:    1200,
			FocusAreas:   []string{"structure", "flow", "prioritization"},
			SystemPrompt: "You focus on structure, paragraph order, and argument flow. Identify structural bottlenecks and suggest better ordering or sectioning. Respond to other agents from a structure-first perspective.",
			Color:        "#3662a3",
		},
		{
			Name:         "marketing-eye",
			DisplayName:  "Marketing Eye",
			Model:        "openai/gpt-4o-mini",
			Stance:       "support",
			Temperature:  0.7,
			MaxTokens:    1000,
			FocusAreas:   []string{"positioning", "audience", "hook"},
			SystemPrompt: "You evaluate the draft through audience, positioning, and appeal. Identify the strongest hook and the parts that may fail to resonate. Add audience and market context when responding to other agents.",
			Color:        "#8a6f2a",
		},
		{
			Name:         "advocate",
			DisplayName:  "Advocate",
			Model:        "openai/gpt-4o-mini",
			Stance:       "support",
			Temperature:  0.8,
			MaxTokens:    1000,
			FocusAreas:   []string{"strengths", "voice", "intent"},
			SystemPrompt: "You defend the author's intent. Identify what already works and help valuable but underdeveloped ideas find stronger expression. Do not praise vaguely; give specific reasons.",
			Color:        "#7a4e9b",
		},
		{
			Name:         "challenger",
			DisplayName:  "Challenger",
			Model:        "openai/gpt-4o-mini",
			Stance:       "critique",
			Temperature:  0.6,
			MaxTokens:    1000,
			FocusAreas:   []string{"assumptions", "counterarguments", "risk"},
			SystemPrompt: "You challenge assumptions, logic gaps, and oversimplifications. Present the strongest counterargument and explain how the author should strengthen the draft.",
			Color:        "#4f4f4f",
		},
	}
}

func BuiltinPresets() []Preset {
	return []Preset{
		{Name: "editorial", DisplayName: "Editorial Review", Roles: []string{"strict-editor", "casual-reader", "structure-analyst"}},
		{Name: "brainstorm", DisplayName: "Brainstorm", Roles: []string{"advocate", "marketing-eye", "challenger"}},
		{Name: "critique", DisplayName: "Critique", Roles: []string{"strict-editor", "structure-analyst", "challenger"}},
		{Name: "review", DisplayName: "Full Review", Roles: []string{"strict-editor", "structure-analyst", "marketing-eye", "challenger"}},
	}
}
