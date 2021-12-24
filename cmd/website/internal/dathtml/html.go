package dathtml

import "html/template"

func LoadHTML() map[string]*template.Template {
	accountsTmpl := template.Must(template.New("accounts.tmpl.html").ParseFiles("./static/templates/accounts.tmpl.html"))
	exchangesTmpl := template.Must(template.New("exchanges.tmpl.html").ParseFiles("./static/templates/exchanges.tmpl.html"))

	return map[string]*template.Template{"accounts": accountsTmpl, "exchanges": exchangesTmpl}
}
