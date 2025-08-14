#!/bin/bash

set -e

# Frontend Docker 이미지 빌드 및 배포 스크립트

echo "🔨 Frontend Docker 이미지 빌드 시작..."

# 변수 설정
IMAGE_NAME="logout-token-sample-ui"
REGISTRY="registry-dev2.k-paas.org/kpaas"
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:latest"

# Docker 이미지 빌드
echo "📦 Docker 이미지 빌드 중..."
docker build -t ${IMAGE_NAME}:latest .

# 레지스트리 태그 추가
echo "🏷️  레지스트리 태그 추가..."
docker tag ${IMAGE_NAME}:latest ${FULL_IMAGE_NAME}

# 레지스트리에 푸시
echo "📤 레지스트리에 이미지 푸시 중..."
docker push ${FULL_IMAGE_NAME}

echo "✅ Frontend 이미지 빌드 및 푸시 완료!"
echo "📍 이미지: ${FULL_IMAGE_NAME}"