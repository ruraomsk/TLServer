## Сервер управления режимами светофоров
#### Функции сервера
+ Производить аутентификацию пользователей
+ Отображение светофоров на карте
+ Управление режимами работы светофоров
+ Управление планами координации светофоров
+ Отпображение логов устройств/сервера
+ Управления каталогами с картами
+ Котроль БД на целостность заполнения планов координации


##### Ресурсы:
  + **`/cross`** описание пути и скриптов для получения файлов для перекрестка
  + **`/img`** описание пути и скриптов для получения файлов для основных картинок
  + **`/icons`** описание пути и скриптов для получения файлов для иконок
  + **`/map`** запрос информации для заполнения странички с картой
  + **`/map/serverLogOut`** обработчик выхода из системы
  + **`/map/update`** обновление странички с данными которые попали в область пользователя
  + **`/map/locationButton`** обработчик для формирования новых координат отображения карты
  + **`/cross`** информация о состоянии перекрестка
  + **`/cross/dev`** информация о состоянии перекрестка (информация о дейвайсе)
  + **`/cross/DispatchControlButtons`** обработчик диспетчерского управления (отправка команд управления)
  + **`/cross/control`** расширеная страничка настройки перекрестка
  + **`/cross/control/close`** обработчик закрытия перекрестка
  + **`/cross/control/editable`** обработчик контроля управления перекрестка
  + **`/cross/control/sendButton`** обработчик приема данных от пользователя для отправки на устройство
  + **`/cross/control/checkButton`** обработчик проверки данных
  + **`/cross/control/createButton`** обработчик создания перекрестка
  + **`/cross/control/deleteButton`** обработчик обработчик удаления перекрсетка
  + **`/manage`** обработка создание и редактирования пользователя
  + **`/manage/changepw`** обработчик для изменения пароля
  + **`/manage/delete`** обработчик для удаления аккаунтов
  + **`/manage/add`** обработчик для добавления аккаунтов
  + **`/manage/update`** обработчик для редактирования данных аккаунта
  + **`/manage/crossEditControl`** обработчик по управлению занятых перекрестков
  + **`/manage/crossEditControl/free`** обработчик по управлению освобождению перекрестка
  + **`/manage/stateTest`** обработчик проверки структуры State
  + **`/manage/serverLog`** обработчик по выгрузке лог файлов сервера
  + **`/manage/serverLog/info`** обработчик выбранного лог файла сервера
  + **`/manage/crossCreator`** обработка проверки/создания каталога карты перекрестков
  + **`/manage/crossCreator/checkAllCross`** обработка проверки наличия всех каталогов и файлов необходимых для построения перекрестков
  + **`/manage/crossCreator/checkSelected`** обработка проверки наличия выбранных каталогов и файлов необходимых для построения перекрестков
  + **`/manage/crossCreator/makeSelected`** обработка создания каталога карты перекрестков
  + **`/map/deviceLog`** обработка лога от устройства
  + **`/map/deviceLog/info`** обработка лога устройства по выбранному интеревалу времени
