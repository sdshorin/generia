

привет. У меня есть проект - прототип instagram (подробнее - в README.md). Сейчас я использую Jaeger. Но он deprecated, поэтому мне нужно обновиться на
OpenTelemetry (все пакеты я уже скачал с помощью go get). 
Jagger используется во всех микросервисах (например, services/api-gateway/cmd/main.go), 
так же его параметры указаны в конфиге - pkg/config/config.go

так как Jaeger используется во всех микросервисах, наверное, лучше вынести общий с ним код в pkg?
Jaeger поднимается с помощью docker-compose.yml
