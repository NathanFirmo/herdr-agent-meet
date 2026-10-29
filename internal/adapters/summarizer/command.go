package summarizer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const prompt = `Você recebe trechos que um agente de programação acabou de escrever no terminal.
Resuma numa única frase curta (no máximo 25 palavras), em português do Brasil, em primeira pessoa,
como um colega numa chamada de vídeo reportando ao chefe o que está fazendo.
Sem markdown, sem caminhos de arquivo, sem comandos. Responda só a frase.`

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
