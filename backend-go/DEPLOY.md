# Kubernetes 내부 배포 가이드

## 개요

백엔드 애플리케이션을 Kubernetes 클러스터 내부에만 배포하여, 동일 클러스터의 Keycloak에서 Backchannel Logout 요청을 받을 수 있도록 구성합니다.

## 배포 아키텍처

```
┌─────────────────────┐    ┌─────────────────────┐
│    Keycloak         │───▶│  logout-token-      │
│  (Same Cluster)     │    │  backend Service    │
│                     │    │  (ClusterIP)        │
└─────────────────────┘    └─────────────────────┘
                                      │
                                      ▼
                            ┌─────────────────────┐
                            │  logout-token-      │
                            │  backend Pods       │
                            │  (1 replica)        │
                            └─────────────────────┘
```

## 배포 방법

### 1단계: Docker 이미지 빌드 및 푸시
```bash
cd backend-go
./build-and-push.sh
```

### 2단계: Kubernetes 내부 배포
```bash
# 내부 전용 배포 (Ingress 없음)
./deploy-internal.sh

# 또는 수동으로
kubectl apply -f k8s/all-in-one.yaml
```

## Keycloak 설정

배포 완료 후 다음 URL을 Keycloak 클라이언트의 Backchannel Logout URL로 설정:

```
http://logout-token-backend.default.svc.cluster.local:3001/auth/backchannel-logout
```

### Keycloak 클라이언트 설정 단계:
1. Keycloak Admin Console 접속
2. Clients → cp-client → Settings
3. **Backchannel logout URL** 설정:
   ```
   http://logout-token-backend.default.svc.cluster.local:3001/auth/backchannel-logout
   ```
4. **Backchannel logout session required**: `On`
5. 저장

## ⚠️ 중요한 제약사항

### Replica 1개 제한
현재 구현은 **메모리 기반 세션 스토리지**를 사용하므로 반드시 `replicas: 1`로 설정해야 합니다.

**Multiple replicas 시 발생하는 문제**:
```
사용자 로그인      → Pod A (세션 저장)
SSE 연결          → Pod B (SSE 클라이언트 등록)  
Backchannel Logout → Pod C (세션을 찾을 수 없음!)
```

**프로덕션 환경 확장 방법**:
- Redis 기반 세션 스토리지 구현
- Redis Pub/Sub를 통한 SSE 이벤트 브로드캐스트
- Sticky Session 설정

자세한 내용은 [README.md의 확장성 섹션](./README.md#확장성-및-배포-고려사항)을 참조하세요.

## 리소스 구성

### ConfigMap
- `KEYCLOAK_URL`: Keycloak 서버 URL
- `KEYCLOAK_REALM`: Keycloak 렐름
- `CLIENT_ID`: 클라이언트 ID
- `PORT`: 애플리케이션 포트
- `FRONTEND_URL`: 프론트엔드 URL

### Secret
- `CLIENT_SECRET`: Keycloak 클라이언트 시크릿
- `SESSION_SECRET`: 세션 암호화 키

### Deployment
- **Replicas**: 1개 (**메모리 기반 세션 스토리지로 인한 제약**)
- **Resources**: 최소/최대 리소스 제한
- **Health Checks**: Liveness/Readiness 프로브
- **Security Context**: 비특권 컨테이너, 보안 프로필 적용

### Service
- **Type**: ClusterIP (내부 접근만)
- **Port**: 3001
- **DNS**: `logout-token-backend.default.svc.cluster.local`

## 모니터링 및 관리

### 배포 상태 확인
```bash
# Pod 상태
kubectl get pods -l app=logout-token-backend

# Service 상태
kubectl get svc logout-token-backend

# 로그 확인
kubectl logs -f deployment/logout-token-backend
```

### 헬스체크
```bash
# 클러스터 내부에서 테스트
kubectl run test-pod --image=curlimages/curl --rm -it --restart=Never \
  -- curl http://logout-token-backend.default.svc.cluster.local:3001/auth/backchannel-logout
```

### 환경변수 업데이트
```bash
# ConfigMap 수정
kubectl edit configmap logout-token-configmap

# Secret 수정
kubectl edit secret logout-token-secret

# Pod 재시작 (환경변수 적용)
kubectl rollout restart deployment/logout-token-backend
```

## 테스트

### 1. Backchannel Logout 테스트
1. 애플리케이션에 로그인
2. Keycloak Admin Console에서 사용자 세션 종료
3. 애플리케이션에서 세션 무효화 확인

### 2. 로그 확인
```bash
# Backchannel Logout 호출 로그 확인
kubectl logs -f deployment/logout-token-backend | grep "Backchannel"
```

## 트러블슈팅

### Service 연결 문제
```bash
# Service 엔드포인트 확인
kubectl get endpoints logout-token-backend

# DNS 해결 테스트
kubectl run test-dns --image=busybox --rm -it --restart=Never \
  -- nslookup logout-token-backend.default.svc.cluster.local
```

### Keycloak 연결 확인
```bash
# Keycloak에서 백엔드로 접근 가능한지 확인
# (Keycloak Pod에서 실행)
curl -v http://logout-token-backend.default.svc.cluster.local:3001/auth/backchannel-logout
```

## 네임스페이스 변경

다른 네임스페이스에 배포하려면:

1. YAML 파일의 `namespace` 변경
2. Service URL 업데이트: `logout-token-backend.<NAMESPACE>.svc.cluster.local`
3. Keycloak 설정 업데이트

## 삭제

```bash
# 전체 리소스 삭제
kubectl delete -f k8s/all-in-one.yaml
```

## 보안 고려사항

- 클러스터 내부 통신만 허용
- 네트워크 정책으로 접근 제한 가능
- RBAC 설정으로 권한 관리
- Secret 암호화 저장