[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/uvnTmvcw)

## Conference Link Service

При запросе `createConferenceLink: true` генерируется тестовая ссылка.
В реальном сценарии здесь был бы вызов внешнего API.

Обработка сбоев:
- При недоступности сервиса возвращается ошибка 500
- При таймауте бронь не создаётся
- Реализована идемпотентность (повторные вызовы не создают дубликатов)