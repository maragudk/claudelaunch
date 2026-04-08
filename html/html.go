package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// LaunchResult holds the info needed to render the success page.
type LaunchResult struct {
	Session string
	URL     string
}

func page(title string, body ...Node) Node {
	return HTML5(HTML5Props{
		Title:    title,
		Language: "en",
		Head: []Node{
			Script(Src("https://cdn.tailwindcss.com")),
		},
		Body: []Node{
			Div(Class("min-h-screen bg-gray-950 flex items-center justify-center"),
				Div(Class("bg-gray-900 rounded-lg shadow-lg p-8 w-full max-w-md"),
					Group(body),
				),
			),
		},
	})
}

// IndexPage renders the launcher form.
func IndexPage() Node {
	return page("claudelaunch",
		H1(Class("text-2xl font-bold text-gray-100 mb-6"), Text("claudelaunch")),
		P(Class("text-gray-400 mb-6"), Text("Launch a persistent Claude Code session inside tmux.")),
		FormEl(Method("POST"), Action("/"),
			Label(Class("block text-sm font-medium text-gray-300 mb-2"), For("name"),
				Text("Session name"),
			),
			Input(
				Type("text"),
				Name("name"),
				ID("name"),
				Required(),
				Pattern("[a-zA-Z0-9._-]+"),
				Placeholder("my-session"),
				Class("w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"),
			),
			Button(Type("submit"),
				Class("mt-4 w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md cursor-pointer transition-colors"),
				Text("Launch"),
			),
		),
	)
}

// SuccessPage renders a success message after launching a session.
func SuccessPage(result LaunchResult) Node {
	return page("launched - claudelaunch",
		H1(Class("text-2xl font-bold text-green-400 mb-4"), Text("Session launched")),
		P(Class("text-gray-300 mb-2"), Text("Started tmux session: "),
			Code(Class("bg-gray-800 px-2 py-1 rounded text-blue-400"), Text(result.Session)),
		),
		Iff(result.URL != "", func() Node {
			return Div(Class("mt-4 mb-4"),
				A(Href(result.URL), Target("_blank"),
					Class("block w-full bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded-md cursor-pointer transition-colors text-center"),
					Text("Open Session"),
				),
			)
		}),
		P(Class("text-gray-400 mb-6 text-sm"),
			Text("Or attach locally: "),
			Code(Class("bg-gray-800 px-2 py-1 rounded text-gray-300"), Textf("tmux attach -t %v", result.Session)),
		),
		A(Href("/"), Class("text-blue-400 hover:text-blue-300 underline"), Text("Launch another")),
	)
}

// ErrorPage renders an error message.
func ErrorPage(msg string) Node {
	return page("error - claudelaunch",
		H1(Class("text-2xl font-bold text-red-400 mb-4"), Text("Error")),
		P(Class("text-gray-300 mb-6"), Text(msg)),
		A(Href("/"), Class("text-blue-400 hover:text-blue-300 underline"), Text("Go back")),
	)
}
