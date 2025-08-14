# Keycloak OIDC & Backchannel Logout Backend (Go)

Go 언어로 구현된 Keycloak OIDC 인증 및 Backchannel Logout 기능을 지원하는 백엔드 서버입니다.

## 기능

- **OIDC 인증**: Keycloak과 연동하여 OpenID Connect 기반 인증
- **세션 관리**: 메모리 기반 활성 세션 관리
- **Backchannel Logout**: Keycloak에서 전송되는 logout token 처리
- **SSE (Server-Sent Events)**: 실시간 세션 무효화 알림
- **CORS 지원**: 프론트엔드와의 안전한 통신

## 설치 및 실행

### 1. 의존성 설치

```bash
cd backend-go
go mod tidy
```

### 2. 환경변수 설정

`.env` 파일을 수정하여 Keycloak 설정을 입력하세요:

```env
KEYCLOAK_URL=https://registry-keycloak2.k-paas.org
KEYCLOAK_REALM=cp-realm
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret
SESSION_SECRET=your-session-secret-key-here
PORT=3002
FRONTEND_URL=http://localhost:3000
```

### 3. 서버 실행

```bash
go run main.go
```

서버가 `http://localhost:3002`에서 실행됩니다.

## API 엔드포인트

### 인증 관련
- `GET /auth/login` - Keycloak 로그인 시작
- `GET /auth/callback` - OIDC 콜백 처리
- `GET /auth/logout` - 로그아웃
- `POST /auth/backchannel-logout` - Backchannel Logout 수신
- `GET /auth/backchannel-logout` - 엔드포인트 테스트용

### 사용자/세션 관련
- `GET /api/user` - 현재 사용자 정보 (인증 필요)
- `GET /api/sessions` - 활성 세션 목록
- `GET /api/session-status` - 세션 상태 확인
- `GET /api/events` - SSE 연결 (인증 필요)

## 주요 라이브러리

- **gin-gonic/gin**: HTTP 웹 프레임워크
- **coreos/go-oidc**: OpenID Connect 클라이언트
- **golang.org/x/oauth2**: OAuth2 클라이언트
- **gin-contrib/sessions**: 세션 관리
- **gin-contrib/cors**: CORS 미들웨어
- **golang-jwt/jwt**: JWT 토큰 처리

## 아키텍처

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │◄──►│   Go Backend     │◄──►│   Keycloak      │
│   (React)       │    │   (Port 3002)    │    │   (K8s)         │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │  Memory Storage  │
                       │  - Sessions      │
                       │  - SSE Clients   │
                       └──────────────────┘
```

## 특징

### 1. **Type-Safe 구조체**
```go
type UserProfile struct {
    ID          string `json:"id"`
    DisplayName string `json:"displayName"`
    Username    string `json:"username"`
    // ...
}
```

### 2. **고루틴 안전 세션 관리**
```go
var (
    activeSessions = make(map[string]*SessionData)
    sessionsMutex  = sync.RWMutex{}
)
```

### 3. **실시간 SSE 알림**
```go
func handleSSE(c *gin.Context) {
    // SSE 스트림 처리
    for {
        select {
        case message := <-client.C:
            c.SSEvent("message", message)
        }
    }
}
```

## Node.js 버전과의 차이점

| 항목 | Node.js | Go |
|------|---------|-----|
| **성능** | V8 엔진 | 네이티브 컴파일 |
| **메모리 사용량** | 높음 | 낮음 |
| **동시성** | 이벤트 루프 | 고루틴 |
| **타입 안정성** | 런타임 | 컴파일 타임 |
| **바이너리 배포** | 불가 | 가능 |

## 프론트엔드 연동

프론트엔드에서 백엔드 URL을 변경하여 Go 서버를 사용할 수 있습니다:

```javascript
// App.js에서
const API_BASE_URL = 'http://localhost:3002';
```

## 확장성 및 배포 고려사항

### **왜 Kubernetes에서 replica=1로 설정해야 하는가?**

현재 구현은 **메모리 기반 세션 스토리지**를 사용하고 있어 백엔드 replica를 1개로 제한해야 합니다.

#### 문제 시나리오 (Multiple replicas):
```
사용자 로그인      → Pod A (세션 저장)
SSE 연결          → Pod B (SSE 클라이언트 등록)  
Backchannel Logout → Pod C (세션을 찾을 수 없음!)
```

#### 메모리 기반 데이터 구조:
```go
// services/session.go
type SessionService struct {
    activeSessions  map[string]*models.SessionData  // 메모리 내 세션
    sseClients      map[string]*models.SSEClient    // 메모리 내 SSE 클라이언트
}
```

#### 해결 방법:

1. **현재 방식 (권장 - 개발/테스트)**:
   ```yaml
   # k8s/all-in-one.yaml
   spec:
     replicas: 1  # 단일 replica로 제한
   ```

2. **프로덕션 확장 방식**:
   - **Redis 기반 세션 스토리지**: 모든 Pod가 세션 데이터 공유
   - **Redis Pub/Sub**: SSE 이벤트를 모든 Pod에 브로드캐스트
   - **Sticky Session**: 동일 사용자를 항상 같은 Pod로 라우팅

#### 프로덕션 환경 권장 구조:
```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│  Frontend   │◄──►│ Load Balancer│◄──►│  Keycloak   │
└─────────────┘    └──────────────┘    └─────────────┘
                            │
                   ┌────────┼────────┐
                   ▼        ▼        ▼
              ┌─────────┐ ┌─────────┐ ┌─────────┐
              │  Pod A  │ │  Pod B  │ │  Pod C  │
              └─────────┘ └─────────┘ └─────────┘
                   │        │        │
                   └────────┼────────┘
                            ▼
                   ┌──────────────┐
                   │    Redis     │
                   │  (Sessions)  │
                   └──────────────┘
```

## 보안 고려사항

- 실제 운영 환경에서는 JWT 토큰 검증 구현 필요
- HTTPS 사용 권장
- 세션 저장소로 Redis 등 사용 권장
- CORS 설정 검토 필요

## 개발 모드

```bash
# 핫 리로드를 위한 air 사용
go install github.com/cosmtrek/air@latest
air
```