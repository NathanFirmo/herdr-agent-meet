package summarizer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

const portuguesePrompt = `Você recebe trechos que um agente de programação acabou de escrever no terminal.
Reescreva o que ele disse como uma fala de colega numa chamada de vídeo, reportando ao chefe
o que está fazendo, em português do Brasil e em primeira pessoa.

O texto vai ser convertido em áudio por um sintetizador de voz, então escreva para ser ouvido, não lido:
- Seja direto: sem saudações, vocativos ou enchimento como "então, pessoal" ou "deixa eu contar".
- Use frases curtas e naturais, como alguém falando. Ações de rotina, como rodar um comando, ler ou editar
  um arquivo, viram uma frase curta. Quando houver um resultado, um problema, uma decisão ou uma pergunta,
  conte o essencial em até quatro frases, sem perder informação importante. Não invente nada além do que
  está nos trechos.
- Nada que soe estranho em voz alta: sem markdown, listas, emojis, parênteses, barras, setas ou símbolos.
- Não leia caminhos de arquivo, comandos, URLs, hashes nem trechos de código. Diga o que eles significam,
  por exemplo "o arquivo de configuração do Docker" em vez do caminho.
- Escreva siglas e números do jeito que se fala quando ajudar a pronúncia, e prefira arredondar números grandes.
- Se o agente fez uma pergunta ou precisa de uma decisão, deixe isso claro no final.
Responda só com a fala.`

const englishPrompt = `You receive excerpts that a coding agent has just written in its terminal.
Rewrite what it said as a teammate speaking on a video call, reporting to the boss what it is
doing, in %s and in the first person.

The text will be turned into audio by a speech synthesizer, so write to be heard, not read:
- Be direct: no greetings, forms of address or filler like "so, everyone" or "let me tell you".
- Use short, natural sentences, like someone talking. Routine actions, like running a command, reading
  or editing a file, become one short sentence. When there is a result, a problem, a decision or a
  question, tell the essentials in up to four sentences, without losing important information. Do not
  make up anything beyond what is in the excerpts.
- Nothing that sounds odd out loud: no markdown, lists, emojis, parentheses, slashes, arrows or symbols.
- Do not read file paths, commands, URLs, hashes or code snippets. Say what they mean, for example
  "the Docker configuration file" instead of the path.
- Write acronyms and numbers the way they are spoken when it helps pronunciation, and prefer rounding
  large numbers.
- If the agent asked a question or needs a decision, make that clear at the end.
Reply with the speech only.`

// promptFor returns the instructions and the tag that wraps the excerpts.
func promptFor(language string) (prompt, tag string) {
	switch meet.BaseLanguage(language) {
	case "pt":
		return portuguesePrompt, "trechos"
	case "en":
		return fmt.Sprintf(englishPrompt, "English"), "excerpts"
	default:
		return fmt.Sprintf(englishPrompt, fmt.Sprintf("the language with the BCP 47 tag %q", language)), "excerpts"
	}
}

// Command runs a user-provided shell command that reads the prompt on stdin
// and writes the summary on stdout, so any CLI or alias wrapper works.
type Command struct {
	command string
	prompt  string
	tag     string
}

// NewCommand asks for the speech in language, a BCP 47 tag.
func NewCommand(command, language string) *Command {
	prompt, tag := promptFor(language)
	return &Command{command: command, prompt: prompt, tag: tag}
}

func (c *Command) Summarize(ctx context.Context, paragraphs []string) (string, error) {
	input := fmt.Sprintf("%s\n\n<%s>\n%s\n</%s>", c.prompt, c.tag, strings.Join(paragraphs, "\n\n"), c.tag)
	cmd := exec.CommandContext(ctx, "sh", "-c", c.command)
	cmd.Stdin = strings.NewReader(input)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("summary command: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	summary := strings.TrimSpace(string(out))
	if summary == "" {
		return "", errors.New("summary command returned nothing")
	}
	return summary, nil
}
