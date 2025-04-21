

docker cp /Users/sergejsorin/study/diploma/generia/services/ai-worker/send_message.py generia-ai-worker:/app/send_message.py

docker exec generia-ai-worker python /app/send_message.py
