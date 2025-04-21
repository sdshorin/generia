#!/usr/bin/env python3
"""
Скрипт для отправки сообщения в Kafka изнутри контейнера.
"""

import asyncio
import json
import uuid
from datetime import datetime
from aiokafka import AIOKafkaProducer
from motor.motor_asyncio import AsyncIOMotorClient
from pymongo import MongoClient
from bson import ObjectId
import os
from dotenv import load_dotenv

# Загрузка переменных окружения
load_dotenv()

KAFKA_BROKERS = "kafka:9092"
KAFKA_TOPIC = "generia-tasks"
MONGODB_URI = os.getenv("MONGODB_URI", "mongodb://mongodb:27017")
MONGODB_DB = os.getenv("MONGODB_DB", "generia")



async def create_mongodb_task(task_id, world_id, parameters):
    """
    Создает запись задачи в MongoDB.
    """
    # Создаем клиент MongoDB
    client = AsyncIOMotorClient(MONGODB_URI)
    db = client["generia_ai_worker"]
    
    # Создаем документ задачи с полным набором полей
    task_document = {
        "_id": task_id,
        "world_id": world_id,
        "type": "init_world_creation",
        "status": "pending",
        "parameters": parameters,
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow(),
        "attempt_count": 0,
        "worker_id": None,
        "result": None,
        "error": None
    }
    
    # Вставляем документ в коллекцию
    await db.tasks.insert_one(task_document)
    print(f"Задача создана в MongoDB с ID: {task_id}")

async def send_test_message(prompt, users_count=10, posts_count=50):
    """
    Отправляет тестовое сообщение в Kafka для запуска генерации.
    """
    # Создание идентификаторов для задачи и мира
    task_id = str(uuid.uuid4())
    world_id = str(uuid.uuid4())
    
    # Параметры задачи
    parameters = {
        "user_prompt": prompt,
        "users_count": users_count,
        "posts_count": posts_count,
        "test_mode": True,
        "created_at": datetime.now().isoformat()
    }
    
    # Создаем запись в MongoDB
    await create_mongodb_task(task_id, world_id, parameters)
    
    # # Ждем 5 секунд перед отправкой в Kafka
    # print("Ожидание 5 секунд перед отправкой в Kafka...")
    # await asyncio.sleep(5)
    
    # Создание продюсера Kafka
    producer = AIOKafkaProducer(
        bootstrap_servers=KAFKA_BROKERS,
        value_serializer=lambda v: json.dumps(v, ensure_ascii=False).encode('utf-8')  # Отключаем ASCII-экранирование
    )
    
    try:
        # Запуск продюсера
        await producer.start()
        
        # Формирование сообщения
        message = {
            "event_type": "task_created",
            "task_id": task_id,
            "task_type": "init_world_creation",
            "world_id": world_id,
            "parameters": parameters
        }
        
        # Отправка сообщения
        await producer.send_and_wait(KAFKA_TOPIC, message)
        
        print(f"Сообщение успешно отправлено в Kafka:")
        print(f"  Тема: {KAFKA_TOPIC}")
        print(f"  ID задачи: {task_id}")
        print(f"  ID мира: {world_id}")
        print(f"  Промпт: {prompt}")
        print(f"  Пользователей: {users_count}")
        print(f"  Постов: {posts_count}")
        
    finally:
        # Закрытие продюсера
        await producer.stop()

# Параметры тестовой задачи
async def main():
    prompt = "Реальность, где сны материализуются в физические объекты, которые исчезают на рассвете."
    # prompt = "Мир фентези, в котором маги умеют перемещаться во времени"
    await send_test_message(prompt, 2, 4)

if __name__ == "__main__":
    asyncio.run(main())