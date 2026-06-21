package render

import "github.com/charmbracelet/glamour"

func Markdown(input string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = r.Close()
	}()
	return r.Render(input)
}
