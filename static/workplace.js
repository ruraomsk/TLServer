"use strict";

function login() {
	$.ajax({
		type: 'POST',
		beforeSend: function (request) {
			alert(sessionStorage.getItem("token"));
			request.setRequestHeader("Authorization", "Bearer" + " " + sessionStorage.getItem("token"));
		},
		//            headers: {"Authorization" : "Bearer" + " " + document.cookie},
		url: window.location.href,
		//            data: JSON.stringify(tokenObj),
		//            dataType: 'json',
		success: function (data) {
		doAll(data);
//			console.log(data.point.Point0.X + "," + data.point.Point0.Y + "\n" + data.point.Point1.X + "," + data.point.Point1.Y);
//			goToPage();
//          alert(window.location.href);
		},
		error: function (request, errorMsg) {
			console.log(errorMsg);
		}
	});
}

function doAll(var data) {

    var map = new ymaps.Map('map', {
        center: [54.9912, 73.3685],
        zoom: 15,
    });

    map.setBounds([[data.point.Point0.X,  data.point.Point0.Y],[data.point.Point1.X, data.point.Point1.Y]]);

    for (var trafficLight in data.tflight){
            map.geoObjects.add(new ymaps.Placemark([trafficLight.points.X, trafficLight.points.Y], {
                balloonContent: 'Важные данные с контроллера',
                hintContent: trafficLight.description
            }, {
                iconLayout: createChipsLayout(function (zoom) {
                    // Размер метки будет определяться функией с оператором switch.
                    return calculate(zoom);
                })
            }));
    }

};

function getRandomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);
  return Math.floor(Math.random() * (max - min)) + min; //Максимум не включается, минимум включается
}

var createChipsLayout = function (calculateSize) {
// Создадим макет метки.
    var Chips = ymaps.templateLayoutFactory.createClass(
        '<div class="placemark"  style="background-image:url(\'img/'+ getRandomInt(1, 34) + '.svg\'); background-size: 100%"></div>',
        {
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
                    circleShape = {type: 'Circle', coordinates: [0, 0], radius: size / 2};
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

var calculate = function (zoom) {
    switch (zoom) {
//          case 11:
//            return 5;
//            break;
//          case 12:
//            return 10;
//            break;
//          case 13:
//            return 20;
//            break;
          case 14:
            return 30;
            break;
          case 15:
            return 35;
            break;
          case 16:
            return 50;
            break;
          case 17:
            return 60;
            break;
          case 18:
            return 80;
            break;
          case 19:
            return 130;
            break;
          default:
            return 20;
    }
}

//ymaps.ready(function () {
//    var map = new ymaps.Map('map', {
//        center: [54.9912, 73.3685],
//        zoom: 15,
//    });
//
////    map.setBounds([[55.000001,36],[56.3,36.5]]);
//
//    map.geoObjects.add(new ymaps.Placemark([54.9912, 73.3685], {
//        balloonContent: 'Важные данные с контроллера',
//        hintContent: 'Контроллер №42'
//    }, {
//        iconLayout: createChipsLayout(function (zoom) {
//            // Размер метки будет определяться функией с оператором switch.
//            return calculate(zoom);
//        })
//    }));
//
//});