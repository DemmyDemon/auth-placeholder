package authplaceholder

import (
	"html/template"
	"net/http"
)

// AuthPage is a variable that holds the full HTML that will be presented as the authentication form.
// It is done this way so it is trivially overridable, should you need to.
var AuthPage = `<!DOCTYPE html>
<html>
	<head>
		<title>{{.AuthTitle}}</title>
		{{ if .Stylesheet}}<link rel="stylesheet" href="{{.Stylesheet}}">{{ end }}
	</head>
	<body>
		<dialog id="authdialog" open>
			<h2>{{.AuthTitle}}</h2>
			<form id="authform" action="{{ .ValidatePath }}" method="post">
				<label for="username">Username</label>
				<input type="text" id="username" name="username">
				<label for="password">Password</label>
				<input type="password" id="password" name="password">
				<input type="submit" value="Authenticate">
			</form>
		</dialog>
	</body>
</html>
`

// SendAuthForm sends the auth form to the specified http.ResponseWriter.
// It parses the AuthForm template each time, which is less efficient, but makes it much easier to override if you need it.
func (pa PlaceholderAuth) SendAuthForm(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnauthorized)
	templ := template.Must(template.New("authpage").Parse(AuthPage))
	templ.Execute(w, pa)
}
