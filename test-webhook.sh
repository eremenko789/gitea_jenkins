#!/bin/bash

# Скрипт для тестирования webhook

echo "Testing Gitea-Jenkins Webhook Server"
echo "===================================="

# URL сервера
SERVER_URL="http://localhost:8080"

# Проверка здоровья сервиса
echo "1. Checking health endpoint..."
curl -s "$SERVER_URL/health" | jq '.' || echo "Health check failed"

echo -e "\n2. Testing webhook with PR #453 for 'scada' repository..."

# Тестовый webhook payload для PR #453 в репозитории "scada"
curl -X POST "$SERVER_URL/webhook/gitea" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "opened",
    "number": 453,
    "pull_request": {
      "id": 453,
      "number": 453,
      "title": "Add new SCADA feature",
      "body": "This PR adds a new feature to the SCADA system",
      "state": "open",
      "html_url": "http://gitea:3000/redkitlab/scada/pulls/453",
      "head": {
        "ref": "feature/scada-enhancement",
        "sha": "abc123def456",
        "repo": {
          "name": "scada",
          "full_name": "redkitlab/scada",
          "html_url": "http://gitea:3000/redkitlab/scada",
          "clone_url": "http://gitea:3000/redkitlab/scada.git",
          "owner": {
            "login": "redkitlab",
            "id": 1
          }
        }
      },
      "base": {
        "ref": "main",
        "repo": {
          "name": "scada",
          "full_name": "redkitlab/scada",
          "owner": {
            "login": "redkitlab"
          }
        }
      },
      "user": {
        "login": "developer",
        "id": 2
      }
    },
    "repository": {
      "id": 1,
      "name": "scada",
      "full_name": "redkitlab/scada",
      "html_url": "http://gitea:3000/redkitlab/scada",
      "clone_url": "http://gitea:3000/redkitlab/scada.git",
      "owner": {
        "login": "redkitlab",
        "id": 1
      }
    },
    "sender": {
      "login": "developer",
      "id": 2
    }
  }' | jq '.' || echo "Webhook test failed"

echo -e "\n3. Testing webhook with PR #100 for 'monitoring' repository..."

# Тестовый webhook payload для PR #100 в репозитории "monitoring"
curl -X POST "$SERVER_URL/webhook/gitea" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "opened",
    "number": 100,
    "pull_request": {
      "id": 100,
      "number": 100,
      "title": "Update monitoring dashboard",
      "body": "This PR updates the monitoring dashboard with new metrics",
      "state": "open",
      "html_url": "http://gitea:3000/redkitlab/monitoring/pulls/100",
      "head": {
        "ref": "feature/monitoring-update",
        "sha": "def456ghi789",
        "repo": {
          "name": "monitoring",
          "full_name": "redkitlab/monitoring",
          "html_url": "http://gitea:3000/redkitlab/monitoring",
          "clone_url": "http://gitea:3000/redkitlab/monitoring.git",
          "owner": {
            "login": "redkitlab",
            "id": 1
          }
        }
      },
      "base": {
        "ref": "main",
        "repo": {
          "name": "monitoring",
          "full_name": "redkitlab/monitoring",
          "owner": {
            "login": "redkitlab"
          }
        }
      },
      "user": {
        "login": "developer",
        "id": 2
      }
    },
    "repository": {
      "id": 2,
      "name": "monitoring",
      "full_name": "redkitlab/monitoring",
      "html_url": "http://gitea:3000/redkitlab/monitoring",
      "clone_url": "http://gitea:3000/redkitlab/monitoring.git",
      "owner": {
        "login": "redkitlab",
        "id": 1
      }
    },
    "sender": {
      "login": "developer",
      "id": 2
    }
  }' | jq '.' || echo "Webhook test failed"

echo -e "\n4. Testing webhook with PR #50 for 'unknown' repository (should not trigger any jobs)..."

# Тестовый webhook payload для репозитория, который не отслеживается
curl -X POST "$SERVER_URL/webhook/gitea" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "opened",
    "number": 50,
    "pull_request": {
      "id": 50,
      "number": 50,
      "title": "Test PR for unknown repo",
      "body": "This PR is for a repository that is not tracked",
      "state": "open",
      "html_url": "http://gitea:3000/redkitlab/unknown/pulls/50",
      "head": {
        "ref": "feature/test",
        "sha": "ghi789jkl012",
        "repo": {
          "name": "unknown",
          "full_name": "redkitlab/unknown",
          "html_url": "http://gitea:3000/redkitlab/unknown",
          "clone_url": "http://gitea:3000/redkitlab/unknown.git",
          "owner": {
            "login": "redkitlab",
            "id": 1
          }
        }
      },
      "base": {
        "ref": "main",
        "repo": {
          "name": "unknown",
          "full_name": "redkitlab/unknown",
          "owner": {
            "login": "redkitlab"
          }
        }
      },
      "user": {
        "login": "developer",
        "id": 2
      }
    },
    "repository": {
      "id": 3,
      "name": "unknown",
      "full_name": "redkitlab/unknown",
      "html_url": "http://gitea:3000/redkitlab/unknown",
      "clone_url": "http://gitea:3000/redkitlab/unknown.git",
      "owner": {
        "login": "redkitlab",
        "id": 1
      }
    },
    "sender": {
      "login": "developer",
      "id": 2
    }
  }' | jq '.' || echo "Webhook test failed"

echo -e "\nTest completed!"