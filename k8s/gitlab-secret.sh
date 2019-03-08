#!/bin/bash
kubectl create secret docker-registry gitlab-auth --docker-server=https://registry.gitlab.com --docker-username=$USERNAME --docker-password=$PASSWORD --docker-email=$EMAIL
