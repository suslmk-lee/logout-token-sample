# ArgoCD GitOps 설정 가이드

이 문서는 logout-token-sample 프로젝트의 ArgoCD 기반 GitOps 구성을 설명합니다.

## 🏗 아키텍처

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GitHub        │    │     ArgoCD      │    │   Kubernetes    │
│   Repository    │◄──►│                 │◄──►│    Cluster      │
│                 │    │                 │    │                 │
├─ dev branch     │    ├─ backend-dev    │    ├─ logout-token-  │
├─ feature/*      │    ├─ frontend-dev   │    │   dev namespace │
└─ main branch    │    └─ ...            │    └─ ...           │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 프로젝트 구조

```
argocd/
├── applications/           # ArgoCD Application 매니페스트
│   ├── backend-dev.yaml
│   └── frontend-dev.yaml
├── projects/              # ArgoCD AppProject 설정
│   └── logout-token-project.yaml
└── README.md

backend-go/k8s/
├── base/                  # Kustomize 기본 매니페스트
│   ├── kustomization.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/              # 환경별 Kustomize 오버레이
    └── dev/
        ├── kustomization.yaml
        ├── deployment-patch.yaml
        └── configmap-patch.yaml

frontend/k8s/
├── base/                  # Kustomize 기본 매니페스트
│   ├── kustomization.yaml
│   ├── configmap.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/              # 환경별 Kustomize 오버레이
    └── dev/
        ├── kustomization.yaml
        └── deployment-patch.yaml
```

## 🚀 배포 설정

### 1. ArgoCD AppProject 생성

```bash
kubectl apply -f argocd/projects/logout-token-project.yaml
```

### 2. ArgoCD Applications 생성

```bash
# Backend 애플리케이션
kubectl apply -f argocd/applications/backend-dev.yaml

# Frontend 애플리케이션  
kubectl apply -f argocd/applications/frontend-dev.yaml
```

### 3. GitHub Secrets 설정

GitHub Repository → Settings → Secrets and variables → Actions에서 다음 설정:

```
HARBOR_USERNAME: [Harbor 레지스트리 사용자명]
HARBOR_PASSWORD: [Harbor 레지스트리 비밀번호]
```

## 🔄 CI/CD 워크플로우

### 트리거 조건

- **Backend**: `backend-go/**` 경로 변경 시
- **Frontend**: `frontend/**` 경로 변경 시  
- **Branch**: `dev`, `main`, `feature/*`

### 워크플로우 단계

1. **코드 푸시** → GitHub Repository (`dev` 또는 `feature/*`)
2. **GitHub Actions** → 빌드, 테스트, 이미지 생성
3. **이미지 태그 업데이트** → Kustomization 파일 자동 수정
4. **ArgoCD 감지** → 변경사항 자동 동기화
5. **Kubernetes 배포** → 새 버전 롤아웃

### 이미지 태깅 전략

```
dev 브랜치:        dev-{sha7}     (예: dev-a1b2c3d)
feature 브랜치:    feature-{sha7} (예: feature-x1y2z3a)  
main 브랜치:       latest, {sha7}
```

## 🎯 환경별 설정

### Development 환경

- **Namespace**: `logout-token-dev`
- **Replicas**: Backend(1), Frontend(1)
- **Resources**: 낮은 리소스 할당
- **Auto-sync**: 활성화
- **Self-heal**: 활성화

### 환경 추가 방법

1. **Kustomize Overlay 생성**:
   ```bash
   mkdir -p backend-go/k8s/overlays/staging
   mkdir -p frontend/k8s/overlays/staging
   ```

2. **ArgoCD Application 생성**:
   ```yaml
   # argocd/applications/backend-staging.yaml
   metadata:
     name: logout-token-backend-staging
   spec:
     source:
       path: backend-go/k8s/overlays/staging
     destination:
       namespace: logout-token-staging
   ```

## 🔐 보안 및 권한

### AppProject 권한

- **허용 리소스**: ConfigMap, Secret, Deployment, Service 등
- **금지 리소스**: RBAC, ResourceQuota, LimitRange
- **클러스터 리소스**: Namespace만 허용

### 사용자 역할

| 역할 | 권한 | 대상 환경 |
|------|------|-----------|
| **developer** | 읽기/쓰기 | dev만 |
| **operator** | 읽기/쓰기 | 모든 환경 |
| **readonly** | 읽기 전용 | 모든 환경 |

## 🛠 운영 가이드

### 수동 동기화

```bash
# ArgoCD CLI 사용
argocd app sync logout-token-backend-dev
argocd app sync logout-token-frontend-dev

# UI에서 "SYNC" 버튼 클릭
```

### 롤백

```bash
# 이전 버전으로 롤백
argocd app rollback logout-token-backend-dev

# 특정 Git 커밋으로 롤백
argocd app set logout-token-backend-dev --revision {commit-hash}
```

### 로그 확인

```bash
# ArgoCD 애플리케이션 상태
argocd app get logout-token-backend-dev

# Kubernetes 파드 로그
kubectl logs -f deployment/logout-token-backend -n logout-token-dev
```

### 트러블슈팅

#### 1. 동기화 실패
```bash
# OutOfSync 상태 확인
argocd app get logout-token-backend-dev

# 수동 동기화 시도
argocd app sync logout-token-backend-dev --prune
```

#### 2. 이미지 Pull 실패
```bash
# Secret 확인
kubectl get secret -n logout-token-dev
kubectl describe secret logout-token-backend-secret -n logout-token-dev

# 이미지 태그 확인
kubectl describe deployment logout-token-backend -n logout-token-dev
```

#### 3. 헬스체크 실패
```bash
# Pod 상태 확인
kubectl get pods -n logout-token-dev
kubectl describe pod {pod-name} -n logout-token-dev

# 서비스 엔드포인트 확인
kubectl get endpoints -n logout-token-dev
```

## 📊 모니터링

### ArgoCD 메트릭

- **Sync Status**: OutOfSync, Synced, Unknown
- **Health Status**: Healthy, Progressing, Degraded, Suspended
- **Sync History**: 동기화 이력 및 소요 시간

### 알림 설정

ArgoCD Notifications Controller 설정:

```yaml
# config/argocd-notifications-cm.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.slack: |
    token: $slack-token
  template.app-deployed: |
    message: |
      {{.app.metadata.name}} deployed successfully!
  trigger.on-deployed: |
    - when: app.status.operationState.phase in ['Succeeded']
      send: [app-deployed]
```

## 🔗 참고 자료

- [ArgoCD 공식 문서](https://argo-cd.readthedocs.io/)
- [Kustomize 문서](https://kustomize.io/)
- [GitHub Actions 문서](https://docs.github.com/actions)
- [Harbor 레지스트리 가이드](https://goharbor.io/docs/)

## 🆘 지원

문제가 발생하면 다음을 확인하세요:

1. [ArgoCD UI](https://argocd.example.com) 상태 확인
2. GitHub Actions 빌드 로그 확인  
3. Kubernetes 클러스터 리소스 상태 확인
4. 이 문서의 트러블슈팅 섹션 참조