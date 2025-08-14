import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './App.css';

const API_BASE_URL = 'http://localhost:3001'; // 로컬 개발환경

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [sessions, setSessions] = useState([]);
  const [sessionStatus, setSessionStatus] = useState(null);
  const [sseConnected, setSseConnected] = useState(false);

  // 사용자 정보 조회
  const checkAuthStatus = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`${API_BASE_URL}/api/user`, {
        withCredentials: true
      });
      console.log('User response:', response.data);
      
      // 사용자 데이터를 안전하게 처리
      const userData = response.data.user;
      const safeUser = {
        id: String(userData.id || 'Unknown'),
        name: String(userData.name || 'Unknown'),
        email: String(userData.email || 'No email')
      };
      
      setUser(safeUser);
    } catch (error) {
      console.error('Auth check error:', error);
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  // 세션 상태 확인
  const checkSessionStatus = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/session-status`, {
        withCredentials: true
      });
      setSessionStatus(response.data);
    } catch (error) {
      console.error('세션 상태 확인 실패:', error);
    }
  };

  // 활성 세션 조회
  const fetchActiveSessions = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/sessions`, {
        withCredentials: true
      });
      setSessions(response.data.sessions);
    } catch (error) {
      console.error('세션 조회 실패:', error);
    }
  };

  // 로그인
  const handleLogin = () => {
    window.location.href = `${API_BASE_URL}/auth/login`;
  };

  // 로그아웃
  const handleLogout = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/auth/logout`, {
        withCredentials: true
      });
      
      if (response.data.logoutUrl) {
        window.location.href = response.data.logoutUrl;
      }
    } catch (error) {
      console.error('로그아웃 실패:', error);
    }
  };

  // 페이지 로드 시 인증 상태 확인
  useEffect(() => {
    checkAuthStatus();
    checkSessionStatus();
    fetchActiveSessions();
  }, []);

  // SSE 연결 관리 (사용자 로그인 후)
  useEffect(() => {
    if (!user) return;

    console.log('🔗 Starting SSE connection for user:', user.id);
    
    const eventSource = new EventSource(`${API_BASE_URL}/api/events`, {
      withCredentials: true
    });

    eventSource.onopen = (event) => {
      console.log('✅ SSE connection opened:', event);
      setSseConnected(true);
    };

    eventSource.onmessage = (event) => {
      console.log('📨 SSE message received:', event.data);
      
      if (event.data === 'connected') {
        console.log('🎉 SSE initial connection confirmed');
      } else if (event.data === 'session_invalidated') {
        console.log('🚨 Session invalidated via SSE!');
        checkSessionStatus();
      }
    };

    eventSource.onerror = (error) => {
      console.error('❌ SSE error:', error);
      console.log('EventSource readyState:', eventSource.readyState);
      setSseConnected(false);
      
      // EventSource will automatically try to reconnect unless we close it
      if (eventSource.readyState === EventSource.CLOSED) {
        console.log('🔄 SSE connection was closed, cleaning up');
      }
    };

    return () => {
      console.log('🔚 Cleaning up SSE connection');
      eventSource.close();
      setSseConnected(false);
    };
  }, [user]);

  if (loading) {
    return (
      <div className="app">
        <div className="container">
          <h1>Keycloak Logout Test</h1>
          <div className="loading">로딩 중...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <div className="container">
        <h1>Keycloak OIDC & Backchannel Logout Test</h1>
        
        {user ? (
          <div className="user-section">
            <div className="user-info">
              <h2>로그인된 사용자</h2>
              <p><strong>ID:</strong> {user.id}</p>
              <p><strong>이름:</strong> {user.name}</p>
              <p><strong>이메일:</strong> {user.email}</p>
            </div>
            
            <div className="session-info">
              <h3>세션 상태</h3>
              {sessionStatus && (
                <div className="status-card">
                  <p><strong>인증 상태:</strong> 
                    <span className={sessionStatus.authenticated ? 'status-active' : 'status-inactive'}>
                      {sessionStatus.authenticated ? '인증됨' : '미인증'}
                    </span>
                  </p>
                  <p><strong>세션 상태:</strong> 
                    <span className={sessionStatus.sessionActive ? 'status-active' : 'status-inactive'}>
                      {sessionStatus.sessionActive ? '활성' : '비활성'}
                    </span>
                  </p>
                  <p><strong>실시간 알림:</strong> 
                    <span className={sseConnected ? 'status-active' : 'status-inactive'}>
                      {sseConnected ? 'SSE 연결됨' : 'SSE 연결 안됨'}
                    </span>
                  </p>
                  {!sessionStatus.sessionActive && sessionStatus.authenticated && (
                    <div className="warning">
                      ⚠️ 세션이 Backchannel Logout에 의해 무효화되었습니다!
                    </div>
                  )}
                </div>
              )}
            </div>
            
            <button className="logout-btn" onClick={handleLogout}>
              로그아웃
            </button>
          </div>
        ) : (
          <div className="login-section">
            <p>Keycloak을 통해 로그인하세요.</p>
            <button className="login-btn" onClick={handleLogin}>
              로그인
            </button>
          </div>
        )}
        
        <div className="sessions-section">
          <h3>활성 세션 목록</h3>
          {sessions.length > 0 ? (
            <div className="sessions-list">
              {sessions.map((session) => (
                <div key={session.sessionId} className="session-card">
                  <p><strong>사용자:</strong> {session.userName}</p>
                  <p><strong>사용자 ID:</strong> {session.userId}</p>
                  <p><strong>로그인 시간:</strong> {new Date(session.loginTime).toLocaleString()}</p>
                </div>
              ))}
            </div>
          ) : (
            <p>활성 세션이 없습니다.</p>
          )}
        </div>
        
        <div className="info-section">
          <h3>테스트 방법</h3>
          <ol>
            <li>위의 로그인 버튼을 클릭하여 Keycloak으로 로그인</li>
            <li>Keycloak Admin Console에서 해당 사용자의 세션을 강제로 로그아웃</li>
            <li>이 페이지에서 세션 상태가 자동으로 업데이트되는지 확인</li>
            <li>Backchannel Logout이 정상 작동하면 세션이 비활성 상태로 변경됨</li>
          </ol>
          
          <div className="endpoint-info">
            <h4>백엔드 엔드포인트</h4>
            <p><code>POST /auth/backchannel-logout</code> - Backchannel Logout 수신</p>
            <p><code>GET /api/session-status</code> - 현재 세션 상태 확인</p>
            <p><code>GET /api/sessions</code> - 활성 세션 목록</p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;