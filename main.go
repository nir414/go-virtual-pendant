// go-virtual-pendant/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// jogHandler handles jog command requests
func jogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd JogCommand
	err := json.NewDecoder(r.Body).Decode(&cmd)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 로봇에 JOG 명령 전송
	response, err := sendJogCommand(cmd)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if response.Success {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// jogStateHandler handles jog state requests
func jogStateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getRobotData()
	if err != nil {
		http.Error(w, "Failed to fetch jog state", http.StatusBadGateway)
		return
	}

	// 로봇 데이터를 서버 로그에 출력
	log.Printf("🤖 API 요청 - 조인트: %v, 카르테시안: %v", data.Joint, data.Cartesian)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// webInterfaceHandler는 이제 web.go에서 처리됩니다 (이 함수 삭제)
// func webInterfaceHandler...

// setJogModeHandler handles jog mode change requests
func setJogModeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Mode string `json:"mode"` // "computer", "joint", "world", "tool", "free"
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := setRobotJogMode(req.Mode)
	if err != nil {
		http.Error(w, "Failed to set jog mode", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// setAxisHandler handles axis selection requests
func setAxisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Axis  int `json:"axis"`  // 1-6 for joints, 1-6 for cartesian
		Robot int `json:"robot"` // robot number (usually 1)
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := setRobotAxis(req.Axis, req.Robot)
	if err != nil {
		http.Error(w, "Failed to set axis", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// main function - server entry point
func main() {
	// 정적 파일 서빙 (CSS, JS)
	http.HandleFunc("/static/", staticFileHandler)

	// API 엔드포인트 등록
	http.HandleFunc("/api/jog", jogHandler)
	http.HandleFunc("/api/jog/state", jogStateHandler)
	http.HandleFunc("/api/jog/mode", setJogModeHandler)
	http.HandleFunc("/api/jog/axis", setAxisHandler)

	// 웹 인터페이스 (템플릿 사용)
	http.HandleFunc("/", webInterfaceHandler)

	log.Println("🚀 Virtual Pendant API running on http://localhost:8082")
	log.Println("🌐 웹 인터페이스: http://localhost:8082")
	log.Println("📍 로봇 위치 모니터링 시작 (5초 간격)")

	// 로봇 위치 모니터링 고루틴 시작
	go monitorRobotPosition()

	log.Fatal(http.ListenAndServe(":8082", nil))
}
