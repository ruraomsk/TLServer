## Сервер управления режимами светофоров
#### Функции сервера
+ Производить аутентификацию пользователей
+ Отображение положения светофоров на карте
+ Управление режимами работы светофоров

##### Функции:
+ **Функции без префикса**
  + **`/login`** метод **`POST`** принимает поля login и password 
  + (должна быть в префиксе /user)**`/create`** метод **`POST`** принимает поля login, password, boxpoint[point0,point1], wtime  _(wtime - время заданное для работы создаваемого пользователя)_
  
+ **Функции с префиксом /user**
  + **`/create`** метод **`POST`** принимает поля login, password, boxpoint[point0,point1], wtime  |(wtime - время заданное для работы создаваемого пользователя)
  + **`{slug}`** - пользователь который прошел авторизацию в системе
  + **`/{slug}`** метод **`GET`** возврящает для браузера workplace.html
  + **`/{slug}`** метод **`POST`**  возвращает поля boxpoint[point0,point1], tflight[ID,region[num,name],idevice,tlsost[num,description],points,state[ck,nk,pk,arrays,status,statistics]] _(список светофоров)_, ya_map _(ключ для yandexmap)_
  + **`/{slug}/update`** метод **`POST`** принимает boxpoint[point0,point1] возвращает информацию о светофорах которые попали в полученную облатсь tflight[ID,region[num,name],idevice,tlsost[num,description],points,state[ck,nk,pk,arrays,status,statistics]]
  + **`/{slug}/cross`** метод **`GET`** возвращает для браузара cross.html
  + **`/{slug}/cross`** метод **`POST`** забирает из URL значение Region и ID, и возвращает информацию о данном перекрестке cross[ID,region[num,name],idevice,tlsost[num,description],points,state[ck,nk,pk,arrays,status,statistics]]
  
+ **Функции с префиксом /file**
  + **`/cross`** метод **`GET`** файловый сервер для перекрество 
  + **`/img`** метод **`GET`** файловый сервер для состояний светофоров
