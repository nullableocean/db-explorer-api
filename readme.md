#### Программа db_explorer

##### Описание

Это простой веб-сервис представляет собой менеджер MySQL-базы данных, который позволяет осуществлять CRUD-запросы (create, read, update, delete) к ней по HTTP

В данном задании идёт отрабатывание навыков работы с HTTP и взаимодействия с базой данных.

АПИ для пользователя:
* `GET /` - возвращает список все таблиц (которые мы можем использовать в дальнейших запросах)
* `GET /{table}?limit=5&offset=7` - возвращает список из 5 записей (limit) начиная с 7-й (offset) из таблицы $table. limit по-умолчанию 5, offset 0
* `GET /{table}/{id}` - возвращает информацию о самой записи или 404
* `PUT /{table}` - создаёт новую запись, данный по записи в теле запроса (POST-параметры)
* `POST /{table}/{id}` - обновляет запись, данные приходят в теле запроса (POST-параметры)
* `DELETE /{table}/{id}` - удаляет запись

Особенности задачи:
* Роутинг запросов - руками, никаких внешних библиотек использовать нельзя.
* Полная динамика. при инициализации в NewDbExplorer считываем из базы список таблиц, полей, далее работаем с ними при валидации. Если добавить третью таблицу - всё должно работать для неё.
* Считаем что во время работы программы список таблиц не меняется
* Валидация на уровне "string - int - float - null", без заморочек.
* Вся работа происходит через database/sql.
* Все имена полей так как они в записаны базе.
* Не забывать про SQL-инъекции

##### Запуск
- `cp .env.example .env`
- `docker compose up --build` - поднять БД для теста
- `make test` - тесты
- `make build && ./main` - запустить сервер


