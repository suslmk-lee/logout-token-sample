# Keycloak OIDC & Backchannel Logout Test Sample

Keycloak과 OIDC 연동 및 Backchannel Logout 기능을 테스트할 수 있는 샘플 애플리케이션입니다.

## 프로젝트 구조

```
logout-token-sample/
├── frontend/           # React 기반 프론트엔드
│   ├── src/
│   │   ├── App.js
│   │   ├── App.css
│   │   ├── index.js
│   │   └── index.css
│   ├── public/
│   ├── Dockerfile
│   ├── k8s/           # Kubernetes 매니페스트
│   └── package.json
├── backend/            # Node.js Express 기반 백엔드 (레거시)
│   ├── server.js
│   ├── package.json
│   └── .env
├── backend-go/         # Go 기반 백엔드 (현재 사용)
│   ├── main.go
│   ├── config/
│   ├── handlers/
│   ├── models/
│   ├── services/
│   ├── middleware/
│   ├── Dockerfile
│   ├── k8s/           # Kubernetes 매니페스트
│   ├── README.md
│   ├── DEPLOY.md
│   └── go.mod
└── README.md
```

## 기능

### Backend (Go + Gin) - **현재 사용**
- **OIDC 기반 Keycloak 인증**: OpenID Connect 프로토콜 구현
- **메모리 기반 세션 관리**: 고루틴 안전 세션 스토리지
- **Backchannel Logout 엔드포인트**: JWT 토큰 파싱 및 세션 무효화
- **SSE (Server-Sent Events)**: 실시간 세션 무효화 알림
- **환경별 설정**: 로컬/프로덕션 자동 구성
- **Kubernetes 지원**: Docker 이미지 및 K8s 매니페스트 제공

### Backend (Node.js + Express) - **레거시**
- 폴링 기반 세션 상태 확인
- 기본적인 Backchannel Logout 구현

### Frontend (React)
- **Keycloak 로그인/로그아웃 UI**: 원클릭 인증
- **실시간 SSE 연결**: 세션 무효화 즉시 감지
- **사용자 정보 표시**: 프로필 정보 및 세션 상태
- **활성 세션 목록**: 현재 로그인된 사용자 목록
- **Docker/Kubernetes 지원**: 컨테이너 배포 가능

## 빠른 시작 가이드

> **권장**: Go 백엔드를 사용하세요. 더 빠르고 안정적이며 SSE를 지원합니다.

### 방법 1: Go 백엔드 사용 (권장)

```bash
# 1. Go 백엔드 실행
cd backend-go
go run .

# 2. 프론트엔드 실행 (별도 터미널)
cd frontend
npm install && npm start
```

### 방법 2: Kubernetes 배포

```bash
# Go 백엔드 배포
cd backend-go
kubectl apply -f k8s/all-in-one.yaml

# 프론트엔드 배포 
cd frontend
kubectl apply -f k8s/all-in-one.yaml
```

자세한 내용은 다음 문서를 참조하세요:
- [Go 백엔드 가이드](./backend-go/README.md)
- [Kubernetes 배포 가이드](./backend-go/DEPLOY.md)

## 설치 및 실행

### 1. Keycloak 설정

먼저 Keycloak에서 클라이언트를 설정해야 합니다.

1. Keycloak Admin Console에 로그인
2. `cp-realm` 렐름 생성 (또는 기존 렐름 사용)
3. 새 클라이언트 생성:
   - Client ID: `your-client-id`
   - Client Type: `OpenID Connect`
   - Standard Flow Enabled: `true`
   - Valid Redirect URIs: `http://localhost:3001/auth/callback` (Go 백엔드용)
   - Backchannel Logout URL: `http://localhost:3001/auth/backchannel-logout` (Go 백엔드용)
   - Backchannel Logout Session Required: `true`

### 2. Go 백엔드 설정 및 실행 (권장)

```bash
cd backend-go

# 의존성 설치
go mod tidy

# 환경변수 설정 (.env 파일 수정)
KEYCLOAK_URL=https://registry-keycloak2.k-paas.org
KEYCLOAK_REALM=cp-realm
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret
SESSION_SECRET=your-session-secret-key-here
PORT=3001
FRONTEND_URL=http://localhost:3000

# 서버 실행
go run .
```

### 2-1. Node.js 백엔드 (레거시)

```bash
cd backend

# 의존성 설치 및 실행
npm install && npm start
```

### 3. Frontend 설정 및 실행

```bash
cd frontend

# 의존성 설치
npm install

# 개발 서버 실행
npm start
```

## 사용법

1. **로그인 테스트**
   - 브라우저에서 `http://localhost:3000` 접속
   - "로그인" 버튼 클릭
   - Keycloak 로그인 페이지에서 인증
   - 로그인 성공 후 사용자 정보 확인

2. **SSE 연결 확인** (Go 백엔드만)
   - 로그인 후 개발자 도구 → Console 확인
   - "SSE 연결됨" 메시지 확인
   - keepalive 메시지가 3초마다 수신되는지 확인

3. **Backchannel Logout 테스트**
   - 로그인 상태에서 Keycloak Admin Console 접속
   - Users → 해당 사용자 → Sessions 탭
   - "Sign out" 버튼 클릭하여 강제 로그아웃
   - **Go 백엔드**: SSE로 즉시 세션 무효화 알림 수신
   - **Node.js 백엔드**: 5초 후 폴링으로 세션 상태 변경 감지

## API 엔드포인트

### 인증 관련
- `GET /auth/login` - Keycloak 로그인 시작
- `GET /auth/callback` - OIDC 콜백
- `GET /auth/logout` - 로그아웃
- `POST /auth/backchannel-logout` - Backchannel Logout 수신

### 사용자/세션 관련
- `GET /api/user` - 현재 사용자 정보 (인증 필요)
- `GET /api/session-status` - 세션 상태 확인
- `GET /api/sessions` - 활성 세션 목록
- `GET /api/events` - SSE 연결 (Go 백엔드만, 인증 필요)

## 주요 특징

### Go vs Node.js 백엔드 비교

| 기능 | Go 백엔드 | Node.js 백엔드 |
|------|-----------|----------------|
| **성능** | ⚡ 네이티브 컴파일 | 🟡 V8 엔진 |
| **메모리 사용량** | ✅ 낮음 | ⚠️ 높음 |
| **실시간 알림** | ✅ SSE 지원 | ❌ 폴링 (5초) |
| **타입 안정성** | ✅ 컴파일 타임 | ⚠️ 런타임 |
| **Docker 이미지** | 📦 ~15MB | 📦 ~150MB |
| **Kubernetes** | ✅ 완전 지원 | ⚠️ 기본 지원 |
| **확장성** | ✅ 고루틴 | 🟡 이벤트루프 |

### Go 백엔드 - SSE 기반 실시간 알림
```go
// handlers/api.go
func (h *APIHandler) HandleSSE(c *gin.Context) {
    // SSE 스트림 생성
    client := &models.SSEClient{
        UserID: userID,
        C:      make(chan string, 10),
        Done:   make(chan bool),
    }
    
    // keepalive로 연결 유지 (3초마다)
    keepalive := time.NewTicker(3 * time.Second)
    
    // 실시간 이벤트 전송
    for {
        select {
        case message := <-client.C:
            c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", message))
        }
    }
}
```

### Backchannel Logout 구현 (Go)
```go
// handlers/auth.go 
func (h *AuthHandler) HandleBackchannelLogout(c *gin.Context) {
    // JWT logout_token 파싱
    claims, err := h.authService.ParseLogoutToken(logoutToken)
    userID := claims["sub"].(string)
    
    // 세션 제거 및 SSE 알림
    h.sessionService.RemoveSession(userID)
    h.sessionService.NotifySessionInvalidated(userID) // 즉시 SSE 전송
}
```

## 아키텍처

### 현재 구조 (Go 백엔드)
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   React         │◄──►│   Go Backend     │◄──►│   Keycloak      │
│   Frontend      │    │   (Port 3001)    │    │   (K8s/Cloud)   │
│   (Port 3000)   │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                        │
        ▼                        ▼
┌─────────────────┐    ┌──────────────────┐
│  SSE Connection │    │  Memory Storage  │
│  (Real-time)    │    │  - Sessions      │
│                 │    │  - SSE Clients   │
└─────────────────┘    └──────────────────┘
```

### 메시지 흐름
1. **로그인**: React → Go → Keycloak → Go → React
2. **SSE 연결**: React ↔ Go (persistent connection)
3. **Backchannel Logout**: Keycloak → Go → SSE → React
4. **즉시 UI 업데이트**: 사용자가 강제 로그아웃되면 3초 내 감지

## 배포 옵션

### 1. 로컬 개발
```bash
# Go 백엔드 (터미널 1)
cd backend-go && go run .

# React 프론트엔드 (터미널 2) 
cd frontend && npm start
```

### 2. Docker 컨테이너
```bash
# 백엔드 빌드
cd backend-go && docker build -t logout-backend .

# 프론트엔드 빌드
cd frontend && docker build -t logout-frontend .
```

### 3. Kubernetes 클러스터
```bash
# 전체 스택 배포
kubectl apply -f backend-go/k8s/all-in-one.yaml
kubectl apply -f frontend/k8s/all-in-one.yaml
```

## 중요한 제약사항

⚠️ **Go 백엔드는 Kubernetes에서 replica=1로만 실행 가능**
- 메모리 기반 세션 스토리지 사용
- Multiple replicas 시 세션 데이터 분산 문제
- 프로덕션에서는 Redis 기반 세션 스토리지 권장

자세한 내용: [확장성 가이드](./backend-go/README.md#확장성-및-배포-고려사항)

## 보안 고려사항

- **JWT 토큰 검증**: 실제 운영 환경에서는 완전한 JWT 검증 구현 필요
- **HTTPS 통신**: 프로덕션에서는 TLS 암호화 필수
- **세션 스토리지**: Redis/PostgreSQL 등 영구 저장소 사용 권장
- **CORS 정책**: 적절한 도메인 제한 설정
- **Security Context**: Kubernetes에서 비특권 컨테이너 실행

## 트러블슈팅

### 1. 로그인이 되지 않는 경우
- Keycloak 클라이언트 설정 확인
- Redirect URI 정확성 확인
- 클라이언트 시크릿 설정 확인

### 2. SSE 연결이 되지 않는 경우 (Go 백엔드)
- 브라우저 개발자 도구에서 CORS 오류 확인
- `withCredentials: true` 설정 확인
- EventSource readyState 확인 (0: CONNECTING, 1: OPEN, 2: CLOSED)

### 3. Backchannel Logout이 작동하지 않는 경우
- Keycloak 클라이언트 설정에서 Backchannel Logout URL 확인
- 백엔드 서버가 Keycloak에서 접근 가능한지 확인
- 로그를 통해 logout_token 수신 여부 확인
- Go 백엔드: SSE 클라이언트 연결 상태 확인

### 4. CORS 오류가 발생하는 경우
- 백엔드의 CORS 설정 확인 (`gin-contrib/cors` 미들웨어)
- SSE 사용 시 `Access-Control-Allow-Credentials: true` 설정
- 프론트엔드 proxy 설정 확인

### 5. Kubernetes 배포 문제
- Pod 상태: `kubectl get pods -l app=logout-token-backend`
- 서비스 확인: `kubectl get svc logout-token-backend`
- 로그 확인: `kubectl logs -f deployment/logout-token-backend`

## 개발 환경

### Go 백엔드 (권장)
- **Go 1.19+**
- **Gin Web Framework**
- **go-oidc v3**
- **golang-jwt/jwt v4**

### Node.js 백엔드 (레거시)
- **Node.js 16+**
- **Express.js**

### 공통
- **React 18+**
- **Keycloak 20+**
- **Docker 20.10+** (컨테이너 배포 시)
- **Kubernetes 1.24+** (K8s 배포 시)

## 성능 벤치마크

| 메트릭 | Go 백엔드 | Node.js 백엔드 |
|--------|-----------|----------------|
| 시작 시간 | ~100ms | ~2000ms |
| 메모리 사용량 | ~15MB | ~80MB |
| 로그인 응답시간 | ~50ms | ~150ms |
| Docker 이미지 크기 | 15MB | 150MB |
| 세션 무효화 지연 | 즉시 (SSE) | ~5초 (폴링) |

## 다음 단계

### 프로덕션 준비사항
1. **Redis 세션 스토리지** 구현 → Multiple replica 지원
2. **JWT 토큰 완전 검증** → 보안 강화  
3. **Health Check 엔드포인트** → 모니터링 개선
4. **Prometheus 메트릭** → 관찰가능성 확보
5. **Helm Chart** → 배포 자동화

### 확장 아이디어
1. **WebSocket 지원** → 양방향 실시간 통신
2. **Multi-tenant 지원** → 여러 조직 관리
3. **Audit Log** → 사용자 활동 추적
4. **Rate Limiting** → API 보호

## 관련 문서

- [Go 백엔드 상세 가이드](./backend-go/README.md)
- [Kubernetes 배포 가이드](./backend-go/DEPLOY.md)
- [프론트엔드 개발 가이드](./frontend/README.md)

## 라이선스

MIT License