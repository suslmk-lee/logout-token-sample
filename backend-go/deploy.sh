#!/bin/bash

set -e

echo "🚀 Kubernetes 배포 스크립트"
echo ""

# 1. 현재 배포 상태 확인
echo "📋 현재 배포 상태 확인..."
kubectl get deployments -l app=logout-token-backend 2>/dev/null || echo "기존 배포 없음"
echo ""

# 2. Kubernetes 리소스 배포
echo "🔄 Kubernetes 리소스 배포 중..."
kubectl apply -f k8s/all-in-one.yaml

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

# 6. Ingress 상태 확인
echo ""
echo "🌍 Ingress 상태:"
kubectl get ingress logout-token-ingress

echo ""
echo "🎉 배포 완료!"
echo ""
echo "유용한 명령어들:"
echo "📜 로그 확인: kubectl logs -f deployment/logout-token-backend"
echo "🔍 Pod 상세: kubectl describe pods -l app=logout-token-backend"
echo "🗑️  삭제: kubectl delete -f k8s/all-in-one.yaml"
echo ""

# 7. 접속 정보 출력
INGRESS_IP=$(kubectl get ingress logout-token-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")
INGRESS_HOST=$(kubectl get ingress logout-token-ingress -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "logout-token-backend.k-paas.org")

echo "📡 접속 정보:"
echo "호스트: ${INGRESS_HOST}"
echo "IP: ${INGRESS_IP}"
echo "테스트 URL: http://${INGRESS_HOST}/auth/backchannel-logout"