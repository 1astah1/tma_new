package middleware

import (
	"net/http"

	"tma-backend/internal/handler"
)

// SuperAdminOnly закрывает управление составом администраторов от остальных
// менеджеров: раздавать доступ к чужим заказам и деньгам может только владелец.
func SuperAdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, role := range handler.GetRoles(r.Context()) {
			if role == "super_admin" {
				next.ServeHTTP(w, r)
				return
			}
		}
		handler.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Управлять админами может только владелец")
	})
}
