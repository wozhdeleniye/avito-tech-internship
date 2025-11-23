[avito-tech.api.testing.postman_collection.json](https://github.com/user-attachments/files/23696774/avito-tech.api.testing.postman_collection.json)## PullRequestController

Приложение по тестовому заданию avito-tech: хранит команды, пользователей и pull requests.

**Билд проекта**

1. В корне репозитория запустите:

```powershell
docker-compose up
```

2. Сервис по умолчанию будет доступен на `http://localhost:8085`.

**Где смотреть конфигурацию**

- Переменные хранятся в `docker-compose.yml` в app:environment.
- Миграции выполняются при старте сервера и дропает все ранее существовавшие таблицы в БД.

**Моя коллекция с постмана для тестирования**

[Uploading avito-tech.{
	"info": {
		"_postman_id": "882bb52f-66b1-4bd5-a9a2-279ab1d7ae87",
		"name": "avito-tech.api.testing",
		"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		"_exporter_id": "32583795"
	},
	"item": [
		{
			"name": "main",
			"item": [
				{
					"name": "create_team",
					"request": {
						"method": "POST",
						"header": [],
						"body": {
							"mode": "raw",
							"raw": "{\r\n  \"team_name\": \"backend2\",\r\n  \"members\": [\r\n    { \"user_id\": \"a11\", \"username\": \"Alice\", \"is_active\": true },\r\n    { \"user_id\": \"a12\", \"username\": \"Bob\", \"is_active\": true },\r\n    { \"user_id\": \"a13\", \"username\": \"Bob\", \"is_active\": true },\r\n    { \"user_id\": \"a14\", \"username\": \"Bob\", \"is_active\": true },\r\n    { \"user_id\": \"a15\", \"username\": \"kendrik_lamar\", \"is_active\": true }\r\n  ]\r\n}"
						},
						"url": {
							"raw": "{{BASE_URL}}team/add",
							"host": [
								"{{BASE_URL}}team"
							],
							"path": [
								"add"
							]
						}
					},
					"response": []
				},
				{
					"name": "create_pr",
					"request": {
						"method": "POST",
						"header": [],
						"body": {
							"mode": "raw",
							"raw": "{\r\n  \"pull_request_id\": \"pr-1191\",\r\n  \"pull_request_name\": \"Add search\",\r\n  \"author_id\": \"a11\"\r\n}"
						},
						"url": {
							"raw": "{{BASE_URL}}pullRequest/create",
							"host": [
								"{{BASE_URL}}pullRequest"
							],
							"path": [
								"create"
							]
						}
					},
					"response": []
				},
				{
					"name": "merge_pr",
					"request": {
						"method": "POST",
						"header": [],
						"body": {
							"mode": "raw",
							"raw": "{\r\n  \"pull_request_id\": \"pr-1191\"\r\n}"
						},
						"url": {
							"raw": "{{BASE_URL}}pullRequest/merge",
							"host": [
								"{{BASE_URL}}pullRequest"
							],
							"path": [
								"merge"
							]
						}
					},
					"response": []
				},
				{
					"name": "get_team",
					"protocolProfileBehavior": {
						"disableBodyPruning": true
					},
					"request": {
						"method": "GET",
						"header": [],
						"body": {
							"mode": "raw",
							"raw": "{\r\n  \"team_name\": \"backend\",\r\n  \"members\": [\r\n    { \"user_id\": \"u1\", \"username\": \"Alice\", \"is_active\": true },\r\n    { \"user_id\": \"u2\", \"username\": \"Bob\", \"is_active\": true }\r\n  ]\r\n}"
						},
						"url": {
							"raw": "{{BASE_URL}}team/get?team_name=backend111",
							"host": [
								"{{BASE_URL}}team"
							],
							"path": [
								"get"
							],
							"query": [
								{
									"key": "team_name",
									"value": "backend111"
								}
							]
						}
					},
					"response": []
				},
				{
					"name": "reassign_pr",
					"request": {
						"method": "POST",
						"header": [],
						"body": {
							"mode": "raw",
							"raw": "{\r\n  \"old_user_id\": \"a12\",\r\n  \"pull_request_id\": \"pr-1191\"\r\n}",
							"options": {
								"raw": {
									"language": "json"
								}
							}
						},
						"url": {
							"raw": "{{BASE_URL}}pullRequest/reassign",
							"host": [
								"{{BASE_URL}}pullRequest"
							],
							"path": [
								"reassign"
							]
						}
					},
					"response": []
				},
				{
					"name": "get_assigned",
					"request": {
						"method": "GET",
						"header": [],
						"url": {
							"raw": "{{BASE_URL}}users/getReview?user_id=a12",
							"host": [
								"{{BASE_URL}}users"
							],
							"path": [
								"getReview"
							],
							"query": [
								{
									"key": "user_id",
									"value": "a12"
								}
							]
						}
					},
					"response": []
				},
				{
					"name": "set_active",
					"request": {
						"method": "POST",
						"header": [],
						"body": {
							"mode": "raw",
							"raw": "{\r\n  \"user_id\": \"a14\",\r\n  \"is_active\": false \r\n}",
							"options": {
								"raw": {
									"language": "json"
								}
							}
						},
						"url": {
							"raw": "{{BASE_URL}}users/setIsActive",
							"host": [
								"{{BASE_URL}}users"
							],
							"path": [
								"setIsActive"
							]
						}
					},
					"response": []
				}
			]
		}
	],
	"event": [
		{
			"listen": "prerequest",
			"script": {
				"type": "text/javascript",
				"packages": {},
				"requests": {},
				"exec": [
					""
				]
			}
		},
		{
			"listen": "test",
			"script": {
				"type": "text/javascript",
				"packages": {},
				"requests": {},
				"exec": [
					""
				]
			}
		}
	],
	"variable": [
		{
			"key": "BASE_URL",
			"value": ""
		},
		{
			"key": "AUTH_PATH",
			"value": ""
		}
	]
}api.testing.postman_collection.json…]()



# Возникшие трудности

1. странное описание примеров id в openapi файле - сделал custom id, который возможно хотели использовать как уникальный неймтег как в телеграмме(зачем такая придумка если можно было б сделать полноценную апку с авторизацией). имею ввиду id у pull_request и user, которые приходят на вход в апи

2. не до конца понимал и понимаю предметную обасть. по сути пользователь не может существовать вне группы. по сути это должен быть грубый инструмент для группировки людей. зачем тогда для этого делать сервис - непонятно. изначально не так понял задание, возможно можно было сделать куда более простое приложение. почему-то думал про гипотетическую масштабируемость и независимость от других сервисов.<br /><br />в конечном итоге решил отказаться от возможности авторизации в угоду соответствия заданию. все-таки ломать не делать

3. некоторые операции требовали создания единых тразакций, пришлось получше разобраться с их устройством в gorm
