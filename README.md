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
│   └── package.json
├── backend/            # Node.js Express 기반 백엔드
│   ├── server.js
│   ├── package.json
│   └── .env
└── README.md
```

## 기능

### Backend (Node.js + Express)
- OIDC 기반 Keycloak 인증
- 세션 관리 및 사용자 정보 저장
- Backchannel Logout 엔드포인트 구현
- 활성 세션 조회 API
- 세션 상태 확인 API

### Frontend (React)
- Keycloak 로그인/로그아웃 UI
- 사용자 정보 표시
- 실시간 세션 상태 모니터링
- 활성 세션 목록 표시
- Backchannel Logout 테스트 가이드

## 설치 및 실행

### 1. Keycloak 설정

먼저 Keycloak에서 클라이언트를 설정해야 합니다.

1. Keycloak Admin Console에 로그인
2. `cp-realm` 렐름 생성 (또는 기존 렐름 사용)
3. 새 클라이언트 생성:
   - Client ID: `your-client-id`
   - Client Type: `OpenID Connect`
   - Standard Flow Enabled: `true`
   - Valid Redirect URIs: `http://localhost:3001/auth/callback`
   - Backchannel Logout URL: `http://localhost:3001/auth/backchannel-logout`
   - Backchannel Logout Session Required: `true`

### 2. Backend 설정 및 실행

```bash
cd backend

# 의존성 설치
npm install

# 환경변수 설정 (.env 파일 수정)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=cp-realm
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret
SESSION_SECRET=your-session-secret-key-here
PORT=3001
FRONTEND_URL=http://localhost:3000

# 서버 실행
npm start
# 또는 개발 모드
npm run dev
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

2. **Backchannel Logout 테스트**
   - 로그인 상태에서 Keycloak Admin Console 접속
   - Users → 해당 사용자 → Sessions 탭
   - "Sign out" 버튼 클릭하여 강제 로그아웃
   - 프론트엔드에서 세션 상태가 자동으로 "비활성"으로 변경되는지 확인

## API 엔드포인트

### 인증 관련
- `GET /auth/login` - Keycloak 로그인 시작
- `GET /auth/callback` - OIDC 콜백
- `GET /auth/logout` - 로그아웃
- `POST /auth/backchannel-logout` - Backchannel Logout 수신

### 사용자/세션 관련
- `GET /api/user` - 현재 사용자 정보
- `GET /api/session-status` - 세션 상태 확인
- `GET /api/sessions` - 활성 세션 목록

## 주요 특징

### Backchannel Logout 구현
```javascript
app.post('/auth/backchannel-logout', (req, res) => {
  const { logout_token } = req.body;
  
  // JWT 토큰 디코딩 및 사용자 식별
  const payload = JSON.parse(Buffer.from(logout_token.split('.')[1], 'base64').toString());
  const userId = payload.sub;
  
  // 활성 세션에서 제거
  if (userId && activeSessions.has(userId)) {
    activeSessions.delete(userId);
    console.log(`User session invalidated: ${userId}`);
  }
  
  res.status(200).send('OK');
});
```

### 실시간 세션 상태 모니터링
프론트엔드에서 5초마다 세션 상태를 확인하여 Backchannel Logout에 의한 세션 무효화를 즉시 감지합니다.

## 보안 고려사항

- 실제 운영 환경에서는 JWT 토큰 검증 구현 필요
- HTTPS 사용 권장
- 세션 저장소로 Redis 등 사용 권장
- CORS 설정 검토 필요

## 트러블슈팅

### 1. 로그인이 되지 않는 경우
- Keycloak 클라이언트 설정 확인
- Redirect URI 정확성 확인
- 클라이언트 시크릿 설정 확인

### 2. Backchannel Logout이 작동하지 않는 경우
- Keycloak 클라이언트 설정에서 Backchannel Logout URL 확인
- 백엔드 서버가 Keycloak에서 접근 가능한지 확인
- 로그를 통해 logout_token 수신 여부 확인

### 3. CORS 오류가 발생하는 경우
- 백엔드의 CORS 설정 확인
- 프론트엔드 proxy 설정 확인

## 개발 환경

- Node.js 16+
- React 18+
- Keycloak 20+

## 라이선스

MIT License