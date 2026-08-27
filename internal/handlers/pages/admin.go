package pages

import (
	"net/http"

	"github.com/a-h/templ"
	admin "github.com/bata94/RegattaApi/internal/templates/pages/admin"
)

func InternalAdmin(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.Dashboard()
}

func InternalAdminUsers(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.Users()
}

func InternalAdminUserGroups(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.UserGroups()
}

func InternalAdminEmailQueue(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.EmailQueue()
}
