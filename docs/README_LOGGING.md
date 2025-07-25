# 로깅 시스템 개선

## 🎯 개선 사항

### 1. 환경변수 기반 로그 레벨 설정

이제 로그 출력을 환경변수로 제어할 수 있습니다:

```bash
# 기본 모드 (필수 정보만)
go run .

# 디버그 모드 (개발자용)
set LOG_LEVEL=DEBUG
go run .

# 상세 모드 (전문가용)
set LOG_LEVEL=VERBOSE
go run .
```

### 2. 로그 레벨별 출력

- **INFO (기본)**: 사용자에게 필요한 핵심 정보만
  - ℹ️ 로봇 명령 실행
  - ℹ️ 모드 변경
  - ℹ️ 성공/실패 메시지

- **DEBUG**: 개발자용 상세 정보
  - 🔍 명령 전송 내용
  - 🔍 통신 에러
  - 🔍 데이터 파싱 상태

- **VERBOSE**: 전문가용 모든 정보
  - 🔧 내부 처리 과정
  - 🔧 메모리 상태
  - 🔧 스레드 정보

### 3. 깔끔한 로그 형태

**이전**: 복잡한 시간 정보와 이모지
```
[15:04:05.000] 🕹️ JOG 명령 수신: 모드=joint, 축=j1, 방향=positive, 스텝=1.000
```

**개선**: 간단하고 직관적
```
ℹ️ JOG 명령 수신: 모드=joint, 축=j1, 방향=positive, 스텝=1.000
```

### 4. 사용자 친화적

- **일반 사용자**: 꼭 필요한 정보만 보임
- **개발자**: `LOG_LEVEL=DEBUG`로 문제 해결 가능
- **전문가**: `LOG_LEVEL=VERBOSE`로 모든 내부 정보 확인

## 🚀 실행 방법

```bash
# 일반 사용
go run .

# 개발/디버깅
set LOG_LEVEL=DEBUG && go run .

# 전문가 모드
set LOG_LEVEL=VERBOSE && go run .
```

이제 더 이상 VS Code 디버거의 복잡한 내부 로그에 혼란스러워하지 않아도 됩니다!
