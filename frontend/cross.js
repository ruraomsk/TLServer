'use strict';

let ID = 0;
let loopFunc;
let osFlag = false;

$(document).ready(function() {

    //Закрытие вкладки при закрытии карты
    window.setInterval(function () {
        if(localStorage.getItem('maintab') === 'closed') window.close();
    }, 250);

    var $table = $('#table');
    $table.bootstrapTable();

    $('#img1').on('click', function() {
        $('#check').trigger('click');
    })

    //Получение информации о перекрёстке
  	$.ajax({
   		type: 'POST',
   		url: window.location.href,
   		success: function (data) {
   		    let region = data.cross.region.num;
   		    let area = data.cross.area.num;
   		    ID = data.cross.ID;
   		    let idevice = data.state.idevice;
   		    document.title = 'ДК-' + ID;
            controlSend({id : data.state.idevice, cmd : 4, param : 1});

            $(window).on("beforeunload", function() {
                controlSend({id : data.state.idevice, cmd : 4, param : 0});
            })

   	        if(data.controlCrossFlag) {
   	            $('#controlButton').show();
   	            $('#p1').show();
   	            $('#p2').show();
   	            $('#jm').show();
   	            $('#os').show();
   	            $('#lr').show();
   	            $('#ky').show();
   	        }

            console.log(data);

            //Отображение полученных данных на экране АРМа
            $('#description').html(data.cross.description);

            $('a').each(function() {
                let id = $(this).attr('id');
                this.className = checkButton(this.className.toString(), data.controlCrossFlag);
                if(id !== 'os'){
                    $('#' + id).on('click', function() {
                        buttonClick(id, data.state.idevice);
                    })
                }
            });

            //OS just because
            $('#os').on('click', function() {
                osFlag = osFlag ? false : true;
                osFlag ? $(this).attr('style', ' background-color: #cccccc;') : $(this).attr('style', ' background-color: #f8f9fa;');
                if (osFlag) {
                    buttonClick('os', data.state.idevice);
                    loopFunc = window.setInterval(function () {
                        buttonClick('os', data.state.idevice);
                    }, 60000);
                } else {
                    clearInterval(loopFunc);
                    loopFunc = undefined;
                }
            })


            $('select').each(function() {
                checkSelect($(this), data.controlCrossFlag);
            });

            //Добавление режима движения и подложки в виде участка карты
            $('#img').attr('src', window.location.origin + '/file/cross/' + region + '/' + area + '/' + ID + '/cross.svg');
            $('#img').attr('style', 'background-size: cover; background-image: url('+ window.location.origin + '/file/cross/' + region + '/' + area + '/' + ID + '/map.png' +'); background-repeat: no-repeat;');

            $('#status').html('Статус: ' + data.cross.tlsost.description);

            $('#controlButton').on('click', function () {
                openPage(window.location.origin + window.location.pathname + '/control' + window.location.search, idevice);
            });

            //Проверка существования карт и добавление их выбора
            let counter = 0;
            data.state.arrays.SetDK.dk.forEach(tab => {
                if (tab.sts[0].stop !== 0) {
                    $('#pk').append(new Option('ПК ' + (counter + 1), counter + 1));
                }
                counter++;
            });
            $('#pk option[value=' + data.state.pk + ']').attr('selected', 'selected');

            counter = 0;
            data.state.arrays.DaySets.daysets.forEach(rec => {
                if (rec.lines[0].npk !== 0) {
                    $('#sk').append(new Option('CК ' + (counter + 1), counter + 1));
                }
                counter++;
            })
            $('#sk option[value=' + data.state.ck + ']').attr('selected', 'selected');

            counter = 0;
            data.state.arrays.WeekSets.wsets.forEach(rec => {
                let flag = true;
                rec.days.forEach(day => {
                    if(rec.days[day] === 0) flag = false;
                })
                if (flag) $('#nk').append(new Option('НК ' + (counter + 1), counter + 1));
                counter++;
            })
            $('#nk option[value=' + data.state.nk + ']').attr('selected', 'selected');

            $('#pk').on('change keyup', function(){
                selectChange('#pk', data.state.idevice);
            });
            $('#sk').on('change keyup', function(){
                selectChange('#sk', data.state.idevice);
            });
            $('#nk').on('change keyup', function(){
                selectChange('#nk', data.state.idevice);
            });
  		},
   		error: function (request, errorMsg) {
            console.log(request.status + ' ' + request.responseText);
   		}
	});
    window.setInterval(function () {
        reload();
    }, 1000);
})

//Функция для обновления данных на странице
function reload() {
    if (!document.hidden) {
        $.ajax({
            type: 'POST',
            url: window.location.href,
            success: function (data) {
                $('#description').html(data.cross.description);
                $('#status').html('Статус: ' + data.cross.tlsost.description);
                $('#pk').find('option').each(function() {
                    $(this).removeAttr('selected');
                });
                $('#sk').find('option').each(function() {
                    $(this).removeAttr('selected');
                });
                $('#nk').find('option').each(function() {
                    $(this).removeAttr('selected');
                });
                $('#pk option[value=' + data.state.pk + ']').attr('selected', 'selected');
                $('#sk option[value=' + data.state.ck + ']').attr('selected', 'selected');
                $('#nk option[value=' + data.state.nk + ']').attr('selected', 'selected');

                if (data.device === undefined) return;
                if (data.device.DK.fdk === 0) return;

                //Обработка таблицы
                let $table = $('#table');
                let dataArr = $table.bootstrapTable('getData');
                let toWrite = {phaseNum : data.device.DK.fdk, tPr : '', tMain : '', duration : ''};
                let checkDup = false;
                let index = 0;
                (data.device.DK.pdk) ? toWrite.tPr = data.device.DK.tdk : toWrite.tMain = data.device.DK.tdk;
                dataArr.forEach(rec => {
                    (rec.phaseNum === data.device.DK.fdk) ? checkDup = true : index++;
                })
                if(!checkDup) {
                    toWrite.duration = toWrite.tMain + toWrite.tPr;
                    dataArr.push(toWrite);
                    dataArr.sort(compare);
                } else {
                    toWrite.phaseNum = dataArr[index].phaseNum;
                    (data.device.DK.pdk) ? toWrite.tMain = dataArr[index].tMain : toWrite.tPr = dataArr[index].tPr;
                    toWrite.duration = toWrite.tMain + toWrite.tPr;
                    $table.bootstrapTable('updateRow', {index: index, row: toWrite});
                }
            },
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    }
}

function compare(a,b) {
    if (a.phaseNum > b.phaseNum) return 1;
    if (b.phaseNum > a.phaseNum) return -1;
    return 0;
}

function buttonClick(button, id){
    let toSend = {id : id, cmd : 9, param : 0};
    switch (button) {
        case 'lr':
            toSend.param = 0;
            break;
        case 'p1':
            toSend.param = 1;
            break;
        case 'p2':
            toSend.param = 2;
            break;
        case 'p3':
            toSend.param = 3;
            break;
        case 'p4':
            toSend.param = 4;
            break;
        case 'p5':
            toSend.param = 5;
            break;
        case 'p6':
            toSend.param = 6;
            break;
        case 'p7':
            toSend.param = 7;
            break;
        case 'p8':
            toSend.param = 8;
            break;
        case 'ky':
            toSend.param = 9;
            break;
        case 'jm':
            toSend.param = 10;
            break;
        case 'os':
            toSend.param = 11;
            break;
    }
    controlSend(toSend);
}

function selectChange(select, id) {
    let toSend = {id : id, cmd : 0, param : 0};
    switch (select) {
        case '#pk':
            toSend.cmd = 5;
            break;
        case '#sk':
            toSend.cmd = 6;
            break;
        case '#nk':
            toSend.cmd = 7;
            break;
    };
    toSend.param = Number($(select).val());
    controlSend(toSend);
}

//Отправка выбранной команды на сервер
function controlSend(toSend) {
    $.ajax({
        url: window.location.origin + window.location.pathname + '/DispatchControlButtons',
        type: 'post',
        dataType: 'json',
        contentType: 'application/json',
        success: function (data) {
            console.log(data);
        },
        data: JSON.stringify(toSend),
        error: function (request) {
            console.log(request.status + ' ' + request.responseText);
        }
    });
}

function checkButton(buttonClass, rights) {
    if(rights) {
        if(buttonClass.indexOf('disabled') !== -1) return buttonClass.substring(0, buttonClass.length-9);
    } else {
        if(buttonClass.indexOf('disabled') === -1) return buttonClass.concat(' disabled');
    }
    return buttonClass;
}

function checkSelect($select, rights) {
    if(rights) {
        $select.prop('disabled', false);
    } else {
        $select.prop('disabled', true);
    }
}

//Функция для открытия новой вкладки
function openPage(url, idevice) {
    controlSend({id : idevice, cmd : 4, param : 0});
	$.ajax({
		url: url,
		type: 'GET',
		success: function (data) {
//		    location.href = url;
            window.open(url);
            window.close()
		},
		error: function (request) {
			console.log(request.status + ' ' + request.responseText);
//			location.href = window.location.origin;
		}
	});
}