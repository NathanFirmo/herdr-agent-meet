package summarizer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const prompt = `Você recebe trechos que um agente de programação acabou de escrever no terminal.
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

// Command runs a user-provided shell command that reads the prompt on stdin
// and writes the summary on stdout, so any CLI or alias wrapper works.
type Command struct {
	command string
}

func NewCommand(command string) *Command {
	return &Command{command: command}
}

func (c *Command) Summarize(ctx context.Context, paragraphs []string) (string, error) {
	input := prompt + "\n\n<trechos>\n" + strings.Join(paragraphs, "\n\n") + "\n</trechos>"
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
