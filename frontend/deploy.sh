#!/bin/bash

set -e

# Frontend Kubernetes 배포 스크립트

echo "🚀 Frontend Kubernetes 배포 시작..."

# 현재 디렉토리 확인
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
K8S_DIR="${SCRIPT_DIR}/k8s"

# k8s 매니페스트 파일 확인
if [ ! -d "${K8S_DIR}" ]; then
    echo "❌ k8s 디렉토리를 찾을 수 없습니다: ${K8S_DIR}"
    exit 1
fi

echo "📁 Kubernetes 매니페스트 디렉토리: ${K8S_DIR}"

# 배포 순서대로 적용
echo "📦 Deployment 적용 중..."
kubectl apply -f "${K8S_DIR}/deployment.yaml"

echo "🔧 Service 적용 중..."
kubectl apply -f "${K8S_DIR}/service.yaml"

echo "🌐 Ingress 적용 중..."
kubectl apply -f "${K8S_DIR}/ingress.yaml"

# 배포 상태 확인
echo "⏳ 배포 상태 확인 중..."
kubectl rollout status deployment/logout-token-frontend -n default

# 서비스 정보 출력
echo "📋 배포된 리소스 정보:"
echo "====================================="
kubectl get deployment logout-token-frontend -n default
echo "-------------------------------------"
kubectl get service logout-token-frontend-service -n default
echo "-------------------------------------"
kubectl get ingress logout-token-frontend-ingress -n default
echo "====================================="

echo "✅ Frontend 배포 완료!"
echo "🌐 접속 URL: https://logout-ui.180.210.83.161.nip.io"