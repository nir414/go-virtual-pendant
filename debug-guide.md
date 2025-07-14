# 🐛 JavaScript 디버깅 가이드

## 📋 브레이크포인트가 작동하지 않는 이유와 해결방법

### 🔍 1. 일반적인 문제점들

#### ❌ 서버가 실행되지 않은 상태
- **문제**: Go 서버가 실행되지 않으면 JavaScript 파일에 접근할 수 없음
- **해결**: VS Code에서 `🚀 Go Run` 태스크 실행 또는 터미널에서 `go run .`

#### ❌ 브라우저 캐시 문제
- **문제**: 이전 버전의 JavaScript 파일이 캐시되어 있음
- **해결**: 브라우저에서 Ctrl+Shift+R (강제 새로고침)

#### ❌ 잘못된 디버깅 설정
- **문제**: Chrome 디버거가 올바른 파일 경로를 찾지 못함
- **해결**: launch.json의 webRoot 경로 확인

### 🛠️ 2. VS Code에서 JavaScript 디버깅 설정

#### Step 1: Go 서버 실행
```bash
# 터미널에서 실행
go run .
```

#### Step 2: 브라우저에서 개발자 도구 사용
1. 크롬 브라우저에서 `http://localhost:8080` 접속
2. `F12` 키 또는 `Ctrl+Shift+I`로 개발자 도구 열기
3. **Sources** 탭 클릭
4. `static/app.js` 파일 찾기
5. 원하는 줄 번호 클릭하여 브레이크포인트 설정

#### Step 3: VS Code에서 Chrome 디버거 연결 (선택사항)
1. VS Code에서 `Ctrl+Shift+D` (디버그 패널 열기)
2. 상단 드롭다운에서 `🌐 Debug JavaScript in Chrome` 선택
3. `F5` 키 또는 초록색 재생 버튼 클릭

### 🎯 3. 효과적인 디버깅 방법들

#### A. Console.log 디버깅 (가장 쉬운 방법)
```javascript
function getSelectedAxis() {
    const select = document.getElementById('axisSelect');
    console.log('🔍 select 요소:', select);        // 요소가 제대로 찾아졌는지 확인
    console.log('🔍 선택된 값:', select.value);    // 현재 값 확인
    return select.value;
}
```

#### B. 브라우저 개발자 도구 활용
```javascript
// 브라우저 콘솔에서 직접 테스트
document.getElementById('axisSelect')           // 요소 존재 확인
document.getElementById('axisSelect').value     // 현재 값 확인
```

#### C. VS Code 디버거 사용
- `F9`: 브레이크포인트 토글
- `F5`: 디버깅 시작
- `F10`: 다음 줄로 이동 (Step Over)
- `F11`: 함수 내부로 들어가기 (Step Into)

### 🚨 4. 자주 발생하는 오류들

#### getElementById가 null을 반환하는 경우
```javascript
function getSelectedAxis() {
    const select = document.getElementById('axisSelect');
    
    // 🛡️ 안전한 코딩: null 체크 추가
    if (!select) {
        console.error('❌ axisSelect 요소를 찾을 수 없습니다!');
        return 'joint1'; // 기본값 반환
    }
    
    return select.value;
}
```

#### HTML 요소 ID 확인하기
```html
<!-- index.html에서 확인해야 할 요소들 -->
<select id="axisSelect">...</select>           <!-- ✅ ID가 정확한지 확인 -->
<span id="selectedAxis"></span>                <!-- ✅ 대소문자 구분 -->
<div id="status"></div>                        <!-- ✅ 오타 없는지 확인 -->
```

### 💡 5. 디버깅 모드 확인하기

#### 현재 디버깅 상태 체크
```javascript
// 브라우저 콘솔에서 실행해보기
console.log('📊 디버깅 정보:');
console.log('- 현재 URL:', window.location.href);
console.log('- DOM 로딩 완료:', document.readyState);
console.log('- axisSelect 요소:', document.getElementById('axisSelect'));
console.log('- selectedAxis 요소:', document.getElementById('selectedAxis'));
```

### 🔧 6. 추천 디버깅 워크플로우

1. **서버 실행 확인** → `go run .`
2. **브라우저 접속** → `http://localhost:8080`
3. **개발자 도구 열기** → `F12`
4. **콘솔에서 테스트** → `document.getElementById('axisSelect')`
5. **브레이크포인트 설정** → Sources 탭에서 app.js 찾기
6. **함수 실행** → 웹페이지에서 실제 동작 테스트

### 🎓 7. getElementById 심화 이해

```javascript
// getElementById의 내부 동작 원리
const element = document.getElementById('axisSelect');

// 이는 다음과 같은 의미:
// 1. document: HTML 문서 전체를 나타내는 객체
// 2. getElementById: ID로 요소를 찾는 메소드
// 3. 'axisSelect': 찾고자 하는 요소의 ID 속성값
// 4. 반환값: HTMLElement 객체 또는 null

// 다른 요소 찾기 방법들과 비교:
document.querySelector('#axisSelect')          // CSS 선택자 사용
document.querySelector('select[id="axisSelect"]') // 속성 선택자
document.getElementsByTagName('select')[0]     // 태그명으로 찾기 (첫 번째)
```

## 🎯 결론

브레이크포인트가 작동하지 않는 가장 흔한 이유는:
1. **서버가 실행되지 않음** (가장 빈번)
2. **브라우저 캐시 문제**
3. **잘못된 파일 경로**

**가장 확실한 디버깅 방법**: 브라우저 개발자 도구의 Console 탭에서 `console.log()` 사용하기!
