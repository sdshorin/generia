


привет. У меня есть проект - генерация виртуальных миров Instagram. подробнее - в README.md. Я хочу полностью обновить фронт: обновить элементы и сделать их более красивми. Для этого мне нужно сформировать техническое задание:
  описать, какие есть страницы и к каким api они обращаются. Напиши на англйиском подробную документацию в фронтенду, чтобы в будущем я использовал это api как часть спецификации-технического задания для нового фронта

  Total cost:            $0.3241
Total duration (API):  2m 47.4s
Total duration (wall): 43m 34.3s
Total code changes:    297 lines added, 0 lines removed




Привет. Я пишу свой проект - подробнее в Readme. Мне нужно удалить из проекта параметр - active world для пользователя (я не хочу сохранять в базу текущий выбранный мир, я хочу чтобы эта информация хранилась на стороне
  клиента). Это api /api/v1/worlds/active  и /api/v1/worlds/set-active, а так же внутренняя логика в микросервисах. Подготовь план по удалению этого парметра, покажи какая логика может сломаться из-за этого удаления. Я хочу
  обсудить план удаления этого праметра, а потом попросить тебя этот план реализовать (этот проект нигде не опубликован, обратная совместимость не нужна, я пока только разрабатываю MVP)

  не меняй фронтенд - он будет полностью переписан. Нужно поменять только бэкенд. Я хочу, чтобы различные миры переключались просто как разные страницы сайта, поэтому мне не нужно поле active world. я так же хочу, чтобы
  {world_id} было явно прописано в url во всех соответствующих API  (это нужно менять в services/api-gateway/cmd/main.go).


Total cost:            $1.47
Total duration (API):  9m 38.2s
Total duration (wall): 2h 2m 49.0s
Total code changes:    7 lines added, 348 lines removed


привет. я пишу свой проект - подробнее в README.md. я хочу полностью передалеть API - хоче передавать id мира прямо в адресе, как /api/v1/worlds/{world_id}/feed. для этого нужно обновить api gateway и фронтенд, а так же
  обновить README.md и frontend/README.md. (это MVP и проект нигде не опубликован, так что можно не заботиться об обратной совместимости)


Total cost:            $2.11
Total duration (API):  8m 24.6s
Total duration (wall): 22m 13.8s
Total code changes:    214 lines added, 147 lines removed



привет. я пишу свой проект - описание в README.md. У меня есть бэкенд и прототип фронтенда. фронтенд описан в frontend/README.md. Я хочу полностью переписать фронтенд с нуля, и сделать его невероятно красивым, стильным,
  дизайнерским и современным, с потрясающим внешним видом и анимациями. Описание новых страниц - в frontend/NEW.md. Перепиши фронтенд полностью, сделай велкиколепный сайт. (старые и не нужные файлы я поместил в
  frontend/old_public и frontend/old_src - они больше не нужны, но если потребуется уточнить что-то по api можешь посмотреть в них).

Теперь максимально подробно и дательно распиши изменения в README.md , чтобы я смог использовать этот файл вместо контекста для LLM при дальнейшей работе с фронтендом. (пиши больше словам, код лучше не использовать. пиши на
  английском. Расписывай используемые модули, чтобы LLM не пришлось искать по всему коду для изменений, а чтобы она сразу понимала куда и что нужно написать )


Total cost:            $6.96
Total duration (API):  26m 35.3s
Total duration (wall): 46m 52.1s
Total code changes:    6540 lines added, 462 lines removed

Правки:

Total cost:            $2.33
Total duration (API):  13m 19.8s
Total duration (wall): 41m 36.9s
Total code changes:    132 lines added, 103 lines removed


Total cost:            $1.57
Total duration (API):  5m 35.1s
Total duration (wall): 30m 51.6s
Total code changes:    42 lines added, 10 lines removed


Total cost:            $1.37
Total duration (API):  6m 1.7s
Total duration (wall): 25m 16.4s
Total code changes:    77 lines added, 37 lines removed


Total cost:            $0.89
Total duration (API):  3m 36.1s
Total duration (wall): 24m 29.3s
Total code changes:    40 lines added, 35 lines removed


Total cost:            $2.76
Total duration (API):  9m 15.3s
Total duration (wall): 40m 34.4s
Total code changes:    190 lines added, 82 lines removed
