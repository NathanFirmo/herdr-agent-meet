package meet

import (
	"fmt"
	"strings"
)

const DefaultLanguage = "pt-BR"

// Locale is what the call says and writes on its own, in the language the
// user picked. Language is a BCP 47 tag such as "pt-BR" or "en-US".
type Locale struct {
	Language string `json:"language"`
	SelfName string `json:"-"`
	// MentionAll is the word after "@" that addresses every agent at once.
	MentionAll  string `json:"mentionAll"`
	BlockedLine string `json:"-"`
	TestLine    string `json:"-"`
	// VoiceBriefing goes before the first message of an agent started from
	// the call, so it writes for the speech synthesizer reading it aloud.
	VoiceBriefing string `json:"-"`
	// Introduction is the first message when the user did not write one.
	Introduction string `json:"-"`
}

var portuguese = Locale{
	SelfName:    "Você",
	MentionAll:  "todos",
	BlockedLine: "Chefe, preciso da sua aprovação aqui!",
	TestLine:    "Oi, chefe! Este é um teste de áudio. Se você está me ouvindo, está tudo certo.",
	VoiceBriefing: "Você está participando de uma chamada de voz: tudo o que você escrever como texto será lido em voz alta por uma ferramenta de TTS (text-to-speech). " +
		"Escreva pensando em quem vai ouvir: frases curtas e naturais, em prosa corrida, como se estivesse falando. " +
		"Evite Markdown, tabelas, listas longas, emojis, URLs e blocos de código, porque eles não são lidos ou soam mal; descreva em palavras o que fez em vez de colar código ou caminhos extensos. " +
		"Seja direto e resuma o essencial, deixando detalhes para quando forem pedidos. " +
		"Fale sempre em português do Brasil. Isso vale só para o texto das suas respostas; use as ferramentas normalmente.",
	Introduction: "Por enquanto, apenas se apresente em uma frase e aguarde instruções.",
}

const englishBriefing = "You are in a voice call: everything you write as text will be read aloud by a TTS (text-to-speech) tool. " +
	"Write for someone listening: short, natural sentences in flowing prose, as if you were talking. " +
	"Avoid Markdown, tables, long lists, emojis, URLs and code blocks, because they are skipped or sound bad; describe in words what you did instead of pasting code or long paths. " +
	"Be direct and stick to the essentials, leaving details for when they are asked for."

var english = Locale{
	SelfName:    "You",
	MentionAll:  "all",
	BlockedLine: "Boss, I need your approval here!",
	TestLine:    "Hi, boss! This is an audio test. If you can hear me, everything is working.",
	VoiceBriefing: englishBriefing + " The descriptions you give your tools, like a command's description field, are read aloud too: " +
		"write them as a short sentence saying what you are about to do. Other than that, use your tools as usual.",
	Introduction: "For now, just introduce yourself in one sentence and wait for instructions.",
}

// NewLocale returns the built-in texts for the language. Languages without
// them get the English texts, and agents are told to reply in the language.
// Tool descriptions are only spoken in English, so only English briefs
// agents about them.
func NewLocale(language string) Locale {
	language = strings.ReplaceAll(strings.TrimSpace(language), "_", "-")
	if language == "" {
		language = DefaultLanguage
	}

	var locale Locale
	switch BaseLanguage(language) {
	case "pt":
		locale = portuguese
	case "en":
		locale = english
	default:
		locale = english
		locale.VoiceBriefing = englishBriefing + fmt.Sprintf(" Always reply in the language with the BCP 47 tag %q instead of English. "+
			"This applies only to the text of your replies; use your tools as usual.", language)
	}
	locale.Language = language
	return locale
}

// BaseLanguage is the primary subtag of a language tag: "pt" for "pt-BR".
func BaseLanguage(tag string) string {
	base, _, _ := strings.Cut(strings.ReplaceAll(tag, "_", "-"), "-")
	return strings.ToLower(strings.TrimSpace(base))
}
