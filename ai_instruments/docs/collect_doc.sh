#!/bin/bash

OUTPUT_FILE="full_doc.md"

# Создаем или очищаем выходной файл
echo "# Полная документация проекта Generia" > "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "*Автоматически сгенерированный документ*" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "## Содержание" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "1. [Общая архитектура](#общая-архитектура)" >> "$OUTPUT_FILE"
echo "2. [Микросервисы](#микросервисы)" >> "$OUTPUT_FILE"
echo "3. [API](#api)" >> "$OUTPUT_FILE"
echo "4. [Фронтенд](#фронтенд)" >> "$OUTPUT_FILE"
echo "5. [Базы данных](#базы-данных)" >> "$OUTPUT_FILE"
echo "6. [Инфраструктура](#инфраструктура)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Массив с именами файлов в порядке добавления
FILES=(
  "general_architecture.md"
  "microservices.md"
  "api.md"
  "frontend.md"
  "db.md"
  "infrastructure.md"
)

# Добавляем содержимое каждого файла
for file in "${FILES[@]}"; do
  # Проверяем существование файла
  if [ -f "${file}" ]; then
    # Извлекаем заголовок первого уровня для создания якоря
    SECTION_TITLE=$(grep "^# " "${file}" | head -n 1 | sed 's/^# //g')
    ANCHOR_NAME=$(echo "${SECTION_TITLE}" | cut -d ' ' -f1-2 | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
    
    # Добавляем заголовок с нужным якорем
    echo "" >> "$OUTPUT_FILE"
    echo "## ${SECTION_TITLE} {#${ANCHOR_NAME}}" >> "$OUTPUT_FILE"
    
    # Пропускаем первую строку (заголовок) и добавляем остальное содержимое
    tail -n +2 "${file}" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "---" >> "$OUTPUT_FILE"
  else
    echo "Предупреждение: Файл ${file} не найден" >&2
  fi
done

echo "Полная документация собрана в файле: $OUTPUT_FILE"
echo "Размер файла: $(du -h "$OUTPUT_FILE" | cut -f1)"