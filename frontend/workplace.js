'use strict';

var currnum = -1;
var IDs = [];
var statuses = [];
let regionInfo;
let areaInfo;

//Функция для открытия вкладки
function openPage(url) {
	$.ajax({
		type: 'GET',
		beforeSend: function (request) {
			request.setRequestHeader('Authorization', 'Bearer ' + sessionStorage.getItem('token'));
		},
		url: window.location.href + url,
		success: function (data) {
		    window.open(window.location.href + url);
		},
		error: function (request) {
			console.log(request.status + ' ' + request.responseText);
//			location.href = window.location.origin;
		}
	});
}

function sleep (time) {
  return new Promise((resolve) => setTimeout(resolve, time));
}

ymaps.ready(function () {

    //Для управления закладками
    localStorage.setItem("maintab", "open");

    window.onbeforeunload = function() {
        localStorage.setItem("maintab", "closed");
        sleep(500);
    }

    $('#changeDialog').show();

    //Открытие личного кабинета
    $('#manageButton').on('click', function() {
        openPage('/manage');
    });

    //Смена пароля текущего аккаунта
	$('#changeButton').on('click', function () {
	    $('#oldPassword').val('');
	    $('#newPassword').val('');
	    $('#repPassword').val('');
		$('#changeDialog').dialog('open');
	});

    //Выбор места для открытия на карте
	$('#locationButton').on('click', function () {
		$('#locationDialog').dialog('open');
	});

    //Выход из аккаунта
    $('#logoutButton').on('click', function() {
        $.ajax({
            type: 'GET',
            beforeSend: function (request) {
                request.setRequestHeader('Authorization', 'Bearer ' + sessionStorage.getItem('token'));
            },
            url: window.location.href + '/logOut',
            success: function (data) {
               location.href = window.location.origin;
            },
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    //Проверка валидности пароля
    $('#newPassword').bind('input', function () {
        $('#newPasswordMsg').remove();
        if ($('#newPassword').val().length < 6){
            $('#newPasswordForm').append('<div style="color: red;" id="newPasswordMsg"><h5>Пароль слишком короткий</h5></div>');
        }
    });

    $('#repPassword').bind('input', function () {
        $('#repPasswordMsg').remove();
        if(($('#newPassword').val() !== $('#repPassword').val()) && ($('#repPassword') !== '')){
            $('#repPasswordForm').append('<div style="color: red;" id="repPasswordMsg"><h5>Пароли не совпадают</h5></div>');
        }
    });

    //Окно изменения пароля
    $('#changeDialog').dialog({
        autoOpen: false,
    	buttons: {
    	    'Подтвердить': function () {
                if (($('#oldPassword').val() === '') || ($('#newPassword').val() === '') || ($('#repPassword').val() === '')) {
    			    if (!($('#oldPasswordMsg').length) && ($('#oldPassword').val() === '')){
    				    $('#oldPasswordForm').append('<div style="color: red;" id="oldPasswordMsg"><h5>Введите старый пароль</h5></div>');
                    }
    				if (!($('#newPasswordMsg').length) && ($('#newPassword').val() === '')){
    					$('#newPasswordForm').append('<div style="color: red;" id="newPasswordMsg"><h5>Введите новый пароль</h5></div>');
                    }
    				if (!($('#repPasswordMsg').length) && ($('#repPassword').val() === '')){
    					$('#repPasswordForm').append('<div style="color: red;" id="repPasswordMsg"><h5>Повторите пароль</h5></div>');
                    }
    				return;
    			}

                var toSend = {
                    oldPW: $('#oldPassword').val(),
                    newPW: $('#newPassword').val()
                };

                $.ajax({
                    url: window.location.href + '/changepw',
                    type: 'post',
                    dataType: 'json',
                    contentType: 'application/json',
                    success: function (data) {
                        console.log(data.msg);
                        $('#changeDialog').dialog('close');
                    },
                    data: JSON.stringify(toSend),
                    error: function (request) {
                        if(request.responseText.message === 'Invalid login credentials') {
        				    $('#oldPasswordForm').append('<div style="color: red;" id="oldPasswordMsg"><h5>Неверный пароль</h5></div>');
                        }
                        if(request.responseText.message === 'Password contains invalid characters') {
    					    $('#newPasswordForm').append('<div style="color: red;" id="newPasswordMsg"><h5>Пароль содержит недопустимые символы</h5></div>');
                        }
                        console.log(request.status + ' ' + request.responseText);
                    }
                });
    		},
    		'Отмена': function () {
    		    $(this).dialog('close');
    		}
    	},
    	modal: true,
    	resizable: false,
    	close: function() {
    	    $('#oldPasswordMsg').remove();
            $('#newPasswordMsg').remove();
            $('#repPasswordMsg').remove();
            $('#repPasswordMsgNotMatch').remove();
    	}
    });

	$.ajax({
		type: 'POST',
		url: window.location.href,
		success: function (data) {
            regionInfo = data.regionInfo;
            areaInfo = data.areaInfo;

            //Заполнение поля выбора регионов для создания пользователя
            for (let reg in regionInfo) {
                $('#region').append(new Option(regionInfo[reg], reg));
            };
            fillAreas();
		    if(data.manageFlag) $('#manageButton').show();

		    //Создание и первичная настройка карты
			var map = new ymaps.Map('map', {
				center: [54.9912, 73.3685],
				zoom: 19,
			});

            map.controls.remove('searchControl');

			map.setBounds([
				[data.boxPoint.point0.Y, data.boxPoint.point0.X],
				[data.boxPoint.point1.Y, data.boxPoint.point1.X]
			]);

			//Разбор полученной от сервера информации
			data.tflight.forEach(trafficLight => {
				currnum = trafficLight.tlsost.num;
				IDs.push(trafficLight.region.num + '-' + trafficLight.area.num + '-' + trafficLight.ID);
				statuses.push(currnum);
				//Создание меток контроллеров на карте
				var placemark = new ymaps.Placemark([trafficLight.points.Y, trafficLight.points.X], {
					hintContent: trafficLight.description
				}, {
					iconLayout: createChipsLayout(function (zoom) {
						// Размер метки будет определяться функией с оператором switch.
						return calculate(zoom);
					}),
				});
				//Функция для вызова АРМ через клик по контроллеру
                placemark.events.add('click', function() {
                    window.open(window.location.href + '/cross?Region=' + trafficLight.region.num + '&Area='+ trafficLight.area.num + '&ID=' + trafficLight.ID);
                })
                //Добавление метки контроллера на карту
				map.geoObjects.add(placemark);
			});
			//Функция для обновления статусов контроллеров в реальном времени
			//Отрисовываются только контроллеры, попадающие в область видимости
			window.setInterval(function () {

    			if (!document.hidden) {
                    var currPoint = map.getBounds(false);
                    var Point00 = {
                        Y: currPoint["0"]["0"],
                        X: currPoint["0"]["1"]
                    };
                    var Point01 = {
                        Y: currPoint["1"]["0"],
                        X: currPoint["1"]["1"]
                    };
                    var point = {
                        Point0: Point00,
                        Point1: Point01
                    };
                    $.ajax({
                        type: 'POST',
                        url: window.location.href + '/update',
                        data: JSON.stringify(point),
                        dataType: 'json',
                        success: function (data) {
                            if (data.tflight === null) {
                                console.log('null');
                            } else {
                                console.log('Обновление');
                                //Обновление статуса контроллера происходит только при его изменении
                                data.tflight.forEach(trafficLight => {
                                    let id = trafficLight.ID;
                                    let index = IDs.indexOf(trafficLight.region.num + '-' + trafficLight.area.num + '-' + id);
                                    let num = trafficLight.tlsost.num;

                                    if (index === -1) {
                                        currnum = trafficLight.tlsost.num;
                                        IDs.push(trafficLight.region.num + '-' + trafficLight.area.num + '-' + id);
                                        statuses.push(currnum);
                                        var placemark = new ymaps.Placemark([trafficLight.points.Y, trafficLight.points.X], {
                                            hintContent: trafficLight.description
                                        }, {
                                            iconLayout: createChipsLayout(function (zoom) {
                                                // Размер метки будет определяться функией с оператором switch.
                                                return calculate(zoom);
                                            })
                                        });
                                        placemark.events.add('click', function() {
                                            window.open(window.location.href + '/cross?Region=' + trafficLight.region.num + '&Area='+ trafficLight.area.num + '&ID=' + trafficLight.ID);
                                        })
                                        //Добавление контроллеров, которые ранее не попадали в область видимости
                                        map.geoObjects.add(placemark);
                                    } else if(statuses[index] != num) {
                                        statuses[index] = num;
                                        currnum = num;
                                        var placemark = new ymaps.Placemark([trafficLight.points.Y, trafficLight.points.X], {
                                            hintContent: trafficLight.description
                                        }, {
                                            iconLayout: createChipsLayout(function (zoom) {
                                                // Размер метки будет определяться функией с оператором switch.
                                                return calculate(zoom);
                                            })
                                        });
                                        placemark.events.add('click', function() {
                                            window.open(window.location.href + '/cross?Region=' + trafficLight.region.num + '&Area='+ trafficLight.area.num + '&ID=' + trafficLight.ID);
                                        })
                                        //Замена метки контроллера со старым состоянием на метку с новым
                                        map.geoObjects.splice(index, 1, placemark);
                                    }
                                })
                            }
                        },
                        error: function (request, errorMsg) {
                            console.log(errorMsg);
                            alert('Пожалуйста, войдите в систему снова.');
                            location.href = window.location.origin;
                        }
                    });
				}
			}, 5000);
            //Всплывающее окно для создания пользователя /locationButton
            $('#locationDialog').dialog({
                autoOpen: false,
                buttons: {
                    'Подтвердить': function () {
                        //Проверка корректности введённых данных
                            if (($('#area option:selected').text() === '')) {
                                if (!($('#areasMsg').length) && ($('#area option:selected').text() === '')){
                                    $('#areasForm').append('<div style="color: red;" id="areasMsg"><h5>Выберите районы</h5></div>');
                                }
                                return;
                            }
                            var selectedAreas = $('#area option:selected').toArray().map(item => item.value);
                            var areas = [];

                            //Сбор данных для отправки на сервер
                            var toSend = {
                                region: $('#region option:selected').val(),
                                area: selectedAreas
                            };

                            //Отправка данных на сервер
                            $.ajax({
                                url: window.location.href + '/locationButton',
                                type: 'post',
                                dataType: 'json',
                                contentType: 'application/json',
                                success: function (data) {
                                    console.log(data.msg);
                                    map.setBounds([
                                        [data.boxPoint.point0.Y, data.boxPoint.point0.X],
                                        [data.boxPoint.point1.Y, data.boxPoint.point1.X]
                                    ]);
                                },
                                data: JSON.stringify(toSend),
                                error: function (request) {
                                    console.log(request.status + ' ' + request.responseText);
                                }
                            });
                            $(this).dialog('close');
                    },
                    'Отмена': function () {
                        $(this).dialog('close');
                    }
                },
                modal: true,
                resizable: false,
                close: function() {
                    $('#areasMsg').remove();
                }
            });
		},
		error: function (request, errorMsg) {
			console.log(errorMsg);
			alert('Пожалуйста, войдите в систему снова.');
			location.href = window.location.origin;
		}
	})

});

//Заполнение поля выбора районов для создания или изменения пользователя
function fillAreas() {
	$('#area').empty();
//	$('#updateArea').empty();
	for (var regAreaJson in areaInfo) {
		for (var areaJson in areaInfo[regAreaJson]) {
			if (regAreaJson === $('#region').find(':selected').text()) {
				$('#area').append(new Option(areaInfo[regAreaJson][areaJson], areaJson));
			}
		};
	};
}

var createChipsLayout = function (calculateSize) {
    if(currnum === 0) {
        console.log('Возвращен несуществующий статус');
        return null;
    }
	// Создадим макет метки.
	var Chips = ymaps.templateLayoutFactory.createClass(
		'<div class="placemark"  style="background-image:url(\'' + window.location.origin + '/file/img/trafficLights/' + currnum + '.svg\'); background-size: 100%"></div>', {
			build: function () {
				Chips.superclass.build.call(this);
				var map = this.getData().geoObject.getMap();
				if (!this.inited) {
					this.inited = true;
					// Получим текущий уровень зума.
					var zoom = map.getZoom();
					// Подпишемся на событие изменения области просмотра карты.
					map.events.add('boundschange', function () {
						// Запустим перестраивание макета при изменении уровня зума.
						var currentZoom = map.getZoom();
						if (currentZoom != zoom) {
							zoom = currentZoom;
							this.rebuild();
						}
					}, this);
				}
				var options = this.getData().options,
					// Получим размер метки в зависимости от уровня зума.
					size = calculateSize(map.getZoom()),
					element = this.getParentElement().getElementsByClassName('placemark')[0],
					// По умолчанию при задании своего HTML макета фигура активной области не задается,
					// и её нужно задать самостоятельно.
					// Создадим фигуру активной области "Круг".
					circleShape = {
						type: 'Circle',
						coordinates: [0, 0],
						radius: size / 2
					};
				// Зададим высоту и ширину метки.
				element.style.width = element.style.height = size + 'px';
				// Зададим смещение.
				element.style.marginLeft = element.style.marginTop = -size / 2 + 'px';
				// Зададим фигуру активной области.
				options.set('shape', circleShape);
			}
		}
	);
	return Chips;
};

//Мастшабирование иконов светофороф на карте
var calculate = function (zoom) {
	switch (zoom) {
//		          case 11:
//		            return 5;
//		            break;
//		          case 12:
//		            return 10;
//		            break;
//		          case 13:
//		            return 20;
//		            break;
		case 14:
			return 30;
			break;
		case 15:
			return 35;
			break;
		case 16:
			return 50;
			break;
		case 17:;
			return 60;
			break;
		case 18:
			return 80;
			break;
		case 19:
			return 130;
			break;
		default:
			return 25;
	}
}

