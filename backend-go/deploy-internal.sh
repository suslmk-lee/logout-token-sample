#!/bin/bash

set -e

echo "🚀 Kubernetes 내부 배포 스크립트 (Ingress 없음)"
echo ""

# 1. 현재 배포 상태 확인
echo "📋 현재 배포 상태 확인..."
kubectl get deployments -l app=logout-token-backend 2>/dev/null || echo "기존 배포 없음"
echo ""

# 2. Kubernetes 리소스 배포 (Ingress 없음)
echo "🔄 Kubernetes 리소스 배포 중..."
kubectl apply -f k8s/all-in-one-no-ingress.yaml

if [ $? -eq 0 ]; then
    echo "✅ Kubernetes 리소스 배포 성공!"
else
    echo "❌ Kubernetes 리소스 배포 실패!"
    exit 1
fi

echo ""

# 3. 배포 상태 확인
echo "⏳ 배포 상태 확인 중..."
kubectl rollout status deployment/logout-token-backend --timeout=300s

# 4. Pod 상태 확인
echo ""
echo "📊 Pod 상태:"
kubectl get pods -l app=logout-token-backend

# 5. Service 상태 확인
echo ""
echo "🌐 Service 상태:"
kubectl get svc logout-token-service

echo ""
echo "🎉 내부 배포 완료!"
echo ""
echo "유용한 명령어들:"
echo "📜 로그 확인: kubectl logs -f deployment/logout-token-backend"
echo "🔍 Pod 상세: kubectl describe pods -l app=logout-token-backend"
echo "🗑️  삭제: kubectl delete -f k8s/all-in-one-no-ingress.yaml"
echo ""

# 6. 클러스터 내부 접속 정보 출력
SERVICE_IP=$(kubectl get svc logout-token-service -o jsonpath='{.spec.clusterIP}')
SERVICE_PORT=$(kubectl get svc logout-token-service -o jsonpath='{.spec.ports[0].port}')

echo "📡 클러스터 내부 접속 정보:"
echo "Service Name: logout-token-service.default.svc.cluster.local"
echo "Service IP: ${SERVICE_IP}"
echo "Service Port: ${SERVICE_PORT}"
echo ""
echo "🔗 Keycloak Backchannel Logout URL:"
echo "http://logout-token-service.default.svc.cluster.local:${SERVICE_PORT}/auth/backchannel-logout"
echo ""
echo "📝 Keycloak 설정에서 위 URL을 Backchannel Logout URL로 사용하세요."