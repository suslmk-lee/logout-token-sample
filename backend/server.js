const express = require('express');
const session = require('express-session');
const passport = require('passport');
const OpenIDConnectStrategy = require('passport-openidconnect').Strategy;
const cors = require('cors');
const { v4: uuidv4 } = require('uuid');
require('dotenv').config();

const app = express();

// 세션 저장소 (실제 운영에서는 Redis 등 사용)
const activeSessions = new Map();

// SSE 클라이언트 저장소
const sseClients = new Map(); // userId -> response 객체

// Middleware
app.use(cors({
  origin: process.env.FRONTEND_URL,
  credentials: true
}));
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// 세션 설정
app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: false, // HTTPS 환경에서는 true로 설정
    httpOnly: true,
    maxAge: 24 * 60 * 60 * 1000 // 24시간
  }
}));

app.use(passport.initialize());
app.use(passport.session());

// Passport 설정
passport.use('oidc', new OpenIDConnectStrategy({
  issuer: `${process.env.KEYCLOAK_URL}/realms/${process.env.KEYCLOAK_REALM}`,
  authorizationURL: `${process.env.KEYCLOAK_URL}/realms/${process.env.KEYCLOAK_REALM}/protocol/openid-connect/auth`,
  tokenURL: `${process.env.KEYCLOAK_URL}/realms/${process.env.KEYCLOAK_REALM}/protocol/openid-connect/token`,
  userInfoURL: `${process.env.KEYCLOAK_URL}/realms/${process.env.KEYCLOAK_REALM}/protocol/openid-connect/userinfo`,
  clientID: process.env.CLIENT_ID,
  clientSecret: process.env.CLIENT_SECRET,
  callbackURL: 'http://localhost:3001/auth/callback',
  scope: 'openid profile email'
}, (iss, userProfile, profile, jwtClaims, accessToken, refreshToken, params, done) => {
  console.log('Callback parameters:');
  console.log('iss:', iss);
  console.log('userProfile:', userProfile);
  console.log('profile:', profile);
  
  // userProfile이 실제 사용자 정보이고, profile은 메타데이터
  const userId = userProfile.id;
  
  // 사용자 정보 처리
  const user = {
    id: userId,
    profile: userProfile,
    accessToken: accessToken,
    refreshToken: refreshToken
  };
  
  // 활성 세션에 사용자 추가
  const sessionId = uuidv4();
  activeSessions.set(userId, {
    sessionId: sessionId,
    user: user,
    loginTime: new Date()
  });
  
  console.log(`User logged in: ${userId}`);
  return done(null, user);
}));

passport.serializeUser((user, done) => {
  done(null, user.id);
});

passport.deserializeUser((id, done) => {
  const session = activeSessions.get(id);
  if (session) {
    done(null, session.user);
  } else {
    done(null, false);
  }
});

// 라우트
app.get('/auth/login', passport.authenticate('oidc'));

app.get('/auth/callback', 
  passport.authenticate('oidc', { failureRedirect: '/' }),
  (req, res) => {
    res.redirect(process.env.FRONTEND_URL);
  }
);

app.get('/auth/logout', (req, res) => {
  const userId = req.user?.id;
  
  req.logout((err) => {
    if (err) {
      console.error('Logout error:', err);
      return res.status(500).json({ error: 'Logout failed' });
    }
    
    // 활성 세션에서 제거
    if (userId) {
      activeSessions.delete(userId);
      console.log(`User logged out: ${userId}`);
    }
    
    // Keycloak 로그아웃 URL로 리다이렉트
    const logoutUrl = `${process.env.KEYCLOAK_URL}/realms/${process.env.KEYCLOAK_REALM}/protocol/openid-connect/logout?redirect_uri=${encodeURIComponent(process.env.FRONTEND_URL)}`;
    res.json({ logoutUrl: logoutUrl });
  });
});

// 사용자 정보 조회
app.get('/api/user', (req, res) => {
  if (req.isAuthenticated()) {
    console.log('User profile:', JSON.stringify(req.user.profile, null, 2));
    
    const profile = req.user.profile;
    let name = 'Unknown';
    
    if (profile.displayName) {
      name = profile.displayName;
    } else if (profile.name && typeof profile.name === 'object') {
      const firstName = profile.name.givenName || '';
      const lastName = profile.name.familyName || '';
      name = `${firstName} ${lastName}`.trim() || profile.username || 'Unknown';
    } else if (profile.username) {
      name = profile.username;
    }
    
    let email = 'No email';
    if (profile.emails && Array.isArray(profile.emails) && profile.emails[0]) {
      email = profile.emails[0].value || profile.emails[0];
    } else if (profile.email) {
      email = profile.email;
    }
    
    res.json({
      user: {
        id: req.user.id,
        name: name,
        email: email
      }
    });
  } else {
    res.status(401).json({ error: 'Not authenticated' });
  }
});

// 활성 세션 조회 (관리용)
app.get('/api/sessions', (req, res) => {
  const sessions = Array.from(activeSessions.entries()).map(([userId, session]) => {
    const profile = session.user.profile;
    let userName = 'Unknown';
    
    if (profile.displayName) {
      userName = profile.displayName;
    } else if (profile.name && typeof profile.name === 'object') {
      const firstName = profile.name.givenName || '';
      const lastName = profile.name.familyName || '';
      userName = `${firstName} ${lastName}`.trim() || profile.username || 'Unknown';
    } else if (profile.username) {
      userName = profile.username;
    }
    
    return {
      userId: userId,
      sessionId: session.sessionId,
      loginTime: session.loginTime,
      userName: userName
    };
  });
  
  res.json({ sessions: sessions });
});

// Backchannel Logout 테스트 엔드포인트 (GET 방식으로 테스트용)
app.get('/auth/backchannel-logout', (req, res) => {
  console.log('=== Backchannel Logout Test (GET) Called ===');
  res.json({ message: 'Backchannel logout endpoint is accessible', timestamp: new Date() });
});

// Backchannel Logout 엔드포인트
app.post('/auth/backchannel-logout', (req, res) => {
  console.log('=== Backchannel Logout Called ===');
  console.log('Request body:', req.body);
  console.log('Content-Type:', req.headers['content-type']);
  
  try {
    const { logout_token } = req.body;
    
    if (!logout_token) {
      console.log('ERROR: Missing logout_token');
      return res.status(400).json({ error: 'Missing logout_token' });
    }
    
    console.log('Logout token received:', logout_token);
    
    // JWT 토큰을 간단히 디코딩 (실제 운영에서는 검증 필요)
    const tokenParts = logout_token.split('.');
    if (tokenParts.length !== 3) {
      console.log('ERROR: Invalid token format');
      return res.status(400).json({ error: 'Invalid token format' });
    }
    
    const payload = JSON.parse(Buffer.from(tokenParts[1], 'base64').toString());
    console.log('Decoded payload:', payload);
    
    // sub 클레임에서 사용자 ID 추출
    const userId = payload.sub;
    console.log('Extracted user ID:', userId);
    console.log('Active sessions before:', Array.from(activeSessions.keys()));
    
    if (userId && activeSessions.has(userId)) {
      activeSessions.delete(userId);
      console.log(`✓ User session invalidated via backchannel logout: ${userId}`);
      
      // SSE로 클라이언트에 즉시 알림
      const sseClient = sseClients.get(userId);
      if (sseClient) {
        sseClient.write('data: session_invalidated\n\n');
        console.log(`📡 SSE notification sent to user: ${userId}`);
      }
    } else {
      console.log(`⚠ User session not found in active sessions: ${userId}`);
    }
    
    console.log('Active sessions after:', Array.from(activeSessions.keys()));
    
    res.status(200).send('OK');
  } catch (error) {
    console.error('Backchannel logout error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// 세션 상태 확인
app.get('/api/session-status', (req, res) => {
  console.log('Session status check called');
  console.log('Is authenticated:', req.isAuthenticated());
  
  if (req.isAuthenticated()) {
    const userId = req.user.id;
    const isSessionActive = activeSessions.has(userId);
    
    console.log('User ID:', userId);
    console.log('Session active:', isSessionActive);
    console.log('Active sessions:', Array.from(activeSessions.keys()));
    
    res.json({ 
      authenticated: true, 
      sessionActive: isSessionActive,
      userId: userId
    });
  } else {
    res.json({ 
      authenticated: false, 
      sessionActive: false 
    });
  }
});

// SSE 엔드포인트
app.get('/api/events', (req, res) => {
  if (!req.isAuthenticated()) {
    return res.status(401).json({ error: 'Not authenticated' });
  }

  // SSE 헤더 설정
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'Access-Control-Allow-Origin': process.env.FRONTEND_URL,
    'Access-Control-Allow-Credentials': 'true'
  });

  const userId = req.user.id;
  
  // 클라이언트 저장
  sseClients.set(userId, res);
  console.log(`SSE client connected: ${userId}`);

  // 연결 확인 메시지
  res.write('data: connected\n\n');

  // 클라이언트 연결 해제 시 정리
  req.on('close', () => {
    sseClients.delete(userId);
    console.log(`SSE client disconnected: ${userId}`);
  });
});

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`Backend server running on http://localhost:${PORT}`);
  console.log(`Keycloak URL: ${process.env.KEYCLOAK_URL}`);
  console.log(`Realm: ${process.env.KEYCLOAK_REALM}`);
});