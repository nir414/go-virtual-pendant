// ============================================================================
// internal/web/handlers.go - 웹 서버 핸들러 및 정적 파일 처리
// ============================================================================
// 웹 인터페이스와 정적 파일 서빙을 담당하는 핸들러들입니다.
// 템플릿 렌더링과 CSS, JS 파일 제공 기능이 포함됩니다.
// ============================================================================

package web

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// ============================================================================
// 웹 인터페이스 핸들러 (Web Interface Handlers)
// ============================================================================

// InterfaceHandler 웹 인터페이스 템플릿 서빙 (외부 호출용)
func InterfaceHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Template loading error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ============================================================================
// 정적 파일 핸들러 (Static File Handlers)
// ============================================================================

// StaticFileHandler 정적 파일 서빙 (외부 호출용)
func StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	// URL에서 /static/ 제거
	file := r.URL.Path[len("/static/"):]
	filePath := filepath.Join("web/static", file)

	// 보안을 위해 경로 검증
	if filepath.Dir(filePath) != "web/static" && filepath.Dir(filePath) != "web\\static" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, filePath)
}
