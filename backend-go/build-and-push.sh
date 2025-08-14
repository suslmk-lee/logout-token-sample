#!/bin/bash

set -e

# 변수 설정
IMAGE_NAME="registry-dev2.k-paas.org/kpaas/logout-token-sample"
TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"

echo "🐳 Docker 이미지 빌드 및 푸시 스크립트"
echo "이미지: ${FULL_IMAGE_NAME}"
echo ""

# 1. Docker 이미지 빌드
echo "📦 Docker 이미지 빌드 중..."
docker build -t ${FULL_IMAGE_NAME} .

if [ $? -eq 0 ]; then
    echo "✅ Docker 이미지 빌드 성공!"
else
    echo "❌ Docker 이미지 빌드 실패!"
    exit 1
fi

# 2. 레지스트리 로그인 (필요한 경우)
echo ""
echo "🔐 레지스트리 로그인 확인..."
# docker login registry-dev2.k-paas.org  # 필요시 주석 해제

# 3. 이미지 푸시
echo "📤 Docker 이미지 푸시 중..."
docker push ${FULL_IMAGE_NAME}

if [ $? -eq 0 ]; then
    echo "✅ Docker 이미지 푸시 성공!"
    echo "이미지: ${FULL_IMAGE_NAME}"
else
    echo "❌ Docker 이미지 푸시 실패!"
    exit 1
fi

echo ""
echo "🎉 Docker 이미지 빌드 및 푸시 완료!"
echo ""
echo "다음 단계:"
echo "1. kubectl apply -f k8s/all-in-one.yaml"
echo "2. kubectl get pods -l app=logout-token-backend"
echo "3. kubectl logs -f deployment/logout-token-backend"