package middleware

import (
    "net/http"
)

// Middleware function to log incoming requests
func LogRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Log request details
        println("Request:", r.Method, r.URL.Path)

        // Call the next handler
        next.ServeHTTP(w, r)
    })
}
