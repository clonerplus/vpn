#!/bin/sh
# Cron job: deactivate expired configs and subscriptions
# Run via Kubernetes CronJob or system crontab
curl -s -X POST http://vpn-manager:8080/api/cleanup
