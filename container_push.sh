#!/bin/bash

AWS_ACOUNT_ID=`aws sts get-caller-identity --query 'Account' --output text`

aws ecr get-login-password | \
  docker login --username AWS --password-stdin ${AWS_ACOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com

docker tag web-page-summarizer-task:latest ${AWS_ACOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/web-page-summarizer-task:latest

docker push ${AWS_ACOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/web-page-summarizer-task:latest
