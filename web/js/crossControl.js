'use strict';

let data;
let unmodifiedData;
let setDK;
let daySets;
let weekSets;
let monthSets;
let stageSets;
let recentlyClicked;
let pkFlag = true;
let skFlag = true;
let kvFlag = true;
let firstLoad = true;

let numberFlag = true;
let longPathFlag = true;

let mainTableFlag = true;
let skTableFlag = true;
let nkTableFlag = false;
let gkTableFlag = false;
let vvTableFlag = false;
let vv2TableFlag = true;
let kvTableFlag = false;

let tempIndex;
let copyArray = [];
let points = {
    Y: 0,
    X: 0
};

let editFlag = false;
let login = '';

//Получение информации из выбранной строки
function getSelectedRowData(table, fullPath, force) {
//    let forceRow = force;
    let index = (force === undefined) ? $('#' + table).find('tr.success').data('index') : force;
    if (force === undefined) tempIndex = index;
    let path = (fullPath !== undefined) ? fullPath.split('.') : undefined;
    let rowData = [];
    if (index === undefined) return;
    if (table === 'pkTable') {
        let selected = $('#pkSelect').val();
        rowData = JSON.parse(JSON.stringify(setDK[selected][path[0]][index]));
    }
    if (table === 'skTable') {
        let selected = $('#mapNum').val();
        rowData = JSON.parse(JSON.stringify(daySets[selected][path[0]][index]));
    }
    if (table === 'nkTable') {
        rowData = JSON.parse(JSON.stringify(weekSets[index][path[0]]));
    }
    if (table === 'gkTable') {
        rowData = JSON.parse(JSON.stringify(monthSets[index][path[0]]));
    }
    if (table === 'kvTable') {
        rowData = JSON.parse(JSON.stringify(stageSets[index]));
    }
    return rowData;
}

function renewal(table, incFlag) {
    let counter = 0;
    if ((incFlag !== undefined) && (incFlag)) tempIndex += 1;
    $('#' + table + ' tbody tr').each(function () {
        if (counter++ === tempIndex) $(this).find('td').trigger('click');
    });
    tempIndex = undefined;
}

//Выделение выбранной строки
function colorizeSelectedRow(table) {
    let index = $('#' + table).find('tr.success').data('index');

    let counter = 0;
    $('#' + table).find('tbody').find('tr').each(function () {
        if (counter === index) {
            $(this).find('td').each(function () {
//                if($(this).attr('style') ===  'background-color: #cccccc') {
//                    $(this).attr('style', '');
//                } else {
                $(this).attr('style', 'background-color: #cccccc')
//                }
            })
        }
        if (counter++ !== index) {
            $(this).find('td').each(function () {
                $(this).attr('style', '')
            })
        }
    });
}

//Функция для проверки возможности редактирования
function checkEdit() {
//    if(localStorage.getItem('maintab') === 'closed') window.close();
    $.ajax({
        url: window.location.origin + window.location.pathname + '/editable' + window.location.search,
        type: 'GET',
        success: function (data) {
//            console.log(data.editInfo);
            login = data.editInfo.login;
            editFlag = data.editInfo.editFlag;
            if (data.editInfo.kick) window.close();
        },
        error: function (request) {
            console.log(request.status + ' ' + request.responseText);
            alert(request.status + ' ' + request.responseText);
            window.location.href = window.location.origin;
        }
    });
}

$(function () {

    checkEdit();
    //Закрытие вкладки при закрытии карты
    window.setInterval(function () {
        checkEdit();
    }, 5000);

//Функционал кнопок управления
    //Кнопка для отпрвления данных на сервер
    $('#sendButton').on('click', function () {
        $.ajax({
            url: window.location.origin + window.location.pathname + '/sendButton',
            type: 'post',
            dataType: 'json',
            contentType: 'application/json',
            success: function (data) {
                console.log(data.message);
                console.log(data.result);
                disableControl('forceSendButton', true);
                disableControl('sendButton', false);
            },
            data: JSON.stringify(data.state),
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    //Кнопка для создания нового перекрёстка
    $('#addButton').on('click', function () {
        data.state.dgis = points.Y + ',' + points.X;
        $.ajax({
            url: window.location.origin + window.location.pathname + '/createButton',
            type: 'post',
            dataType: 'json',
            contentType: 'application/json',
            success: function (data) {
                console.log(data.message);
                console.log(data.result);
//                console.log(data);
            },
            data: JSON.stringify(data.state),
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    //Кнопка для обновления данных на АРМе
    $('#reloadButton').on('click', function () {
        $.ajax({
            url: window.location.href,
            type: 'post',
            contentType: 'application/json',
            success: function (data) {
//                console.log(data);
                loadData(data, firstLoad);
                document.title = 'АРМ ДК-' + data.cross.ID;
                firstLoad = false;
            },
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    //Кнопка для удаления перекрёстка
    $('#deleteButton').on('click', function () {
        if (confirm('Вы уверены? Перекрёсток будет безвозвратно удалён.')) {
            $.ajax({
                url: window.location.origin + window.location.pathname + '/deleteButton',
                type: 'post',
                dataType: 'json',
                contentType: 'application/json',
                success: function (data) {
                    console.log(data);
                    if (data.status) {
                        alert('Перекрёсток удалён, вкладка будет закрыта');
                        window.close();
                    } else {
                        alert('Ожидание ответа сервера, попробуйте еще раз');
                    }
                },
                data: JSON.stringify(data.state),
                error: function (request) {
                    console.log(request.status + ' ' + request.responseText);
                }
            });
        }
    });

    //Кнопка для проверки редактирования перекрестка
    $('#checkEdit').on('click', function () {
        if (editFlag) {
            alert('На данном ДК нет других пользователей');
        } else {
            alert('На даннок ДК работает пользователь ' + login);
        }
    });

    //Кнопка для проверки валидности заполненных данных
    $('#checkButton').on('click', function () {
        $.ajax({
            url: window.location.origin + window.location.pathname + '/checkButton',
            type: 'post',
            dataType: 'json',
            contentType: 'application/json',
            success: function (data) {
                let counter = 0;
                let flag = false;
                disableControl('sendButton', data.status);
                $('#verification').bootstrapTable('removeAll');
                data.result.forEach(function () {
                    if (data.result[counter].includes('Проверка')) {
                        $('#verification').bootstrapTable('append', {
                            left: data.result[counter],
                            right: data.result[counter + 1]
                        });
                        flag = true;
                    } else {
                        (!flag) ? $('#verification').bootstrapTable('append', {
                            left: '',
                            right: data.result[counter]
                        }) : flag = false;
                    }
                    counter++;
                });
                $('#trigger').show();
                if ($('#panel').attr('style') !== 'display: block;') $('#trigger').trigger('click');
                $('th[data-field="left"]').attr('style', 'min-width: 346px;');
                $('th[data-field="right"]').attr('style', 'min-width: 276px;');
            },
            data: JSON.stringify(data.state),
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    $('#questionTrigger').on('click', function () {
        // $('#questionPanel').show();
        if ($('#questionPanel').attr('style') !== 'display: block;') {
            $('#questionPanel').show();
        } else {
            $('#questionPanel').hide();
        }
    });

    $('#markdownTry').on('click', function () {
        $.ajax({
            url: 'https://192.168.115.120:8082/file/markdown/crossControl.md',
            type: 'GET',
            success: function (data) {
//            console.log(data.editInfo);
//                 if (data.editInfo.kick) window.close();
                let converter = new showdown.Converter(),
                    text = data,
                    html = converter.makeHtml(text);
                html = '<div class="row"><div class="col-md-10">' + html + '</div></div>';
                $('#questionTrigger').show();
                if ($('#questionPanel').attr('style') !== 'display: block;') $('#questionTrigger').trigger('click');
                $('#questionPanel').append(html);
            },
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
                alert(request.status + ' ' + request.responseText);
                window.location.href = window.location.origin;
            }
        });
    });
    //Функционирование кнопки с выводом информации о проверке
    $(".trigger").on('click', function () {
        $(".panel").toggle("fast");
        $(this).toggleClass("active");
        return false;
    });

    //Первая загрузка страницы
    $('#reloadButton').trigger('click');

//Функционал кнопок на вкладке "Основные"
    //Кнопка для возвращения исходных данных
    $('#mainReloadButton').on('click', function () {
        mainTabFill(unmodifiedData, false);
    });

//Функционал кнопок на вкладке "ПК"
    //Кнопка для копирования строки
    $('#pkCopyButton').on('click', function () {
        let selected = $('#pkSelect').val();
        copyArray = Object.assign({}, setDK[selected]);
    });

    //Кнопка для перезаписи строки
    $('#pkPasteButton').on('click', function () {
        let selected = $('#pkSelect').val();
        setDK[selected] = Object.assign({}, copyArray);
        pkTabFill('pkTable');
    });

    //Кнопка для возвращения исходных данных
    $('#pkReloadButton').on('click', function () {
        pkTabFill2(unmodifiedData, false);
    });

    //Кнопка для обнуления текущего ПК
    $('#pkNewButton').on('click', function () {
        let selected = $('#pkSelect').val();
        data.state.arrays.SetDK.dk[selected].sts.forEach(row => {
            row.num = 0;
            row.start = 0;
            row.stop = 0;
            row.tf = 0;
            row.plus = false;
        });
        pkTabFill2(data, false);
    });

    //Кнопка для копирования всей информации выбранного ПК
    $('#switchCopy').on('click', function () {
        let index = $('#pkTable').find('tr.success').data('index');
        let selected = $('#pkSelect').val();
        let selectVal = 0;
        let oldData = [];
        let counter = 0;

        if (getSelectedRowData('pkTable', 'sts') === undefined) return;

        $('#pkTable tbody tr').each(function () {
            oldData.push(getSelectedRowData('pkTable', 'sts', counter));
            // });
            // counter = 0;
            // $('#pkTable tbody tr').each(function () {
            if (counter++ === index) {
                let selectPosition = 0;
                $(this).find('td').each(function () {
                    if (selectPosition++ === 2) selectVal = $(this).find('select').children("option:selected").val();
                })
            }
        });
        counter = 0;
        oldData.splice(index, 0, JSON.parse(JSON.stringify(oldData[index])));
        oldData.pop();
        oldData.forEach(rec => {
            rec.line = ++counter;
        });

        oldData[index].tf = Number(selectVal);
        oldData[index + 1].tf = Number(selectVal);
        setDK[selected].sts = oldData;

        pkTabFill('pkTable');
    });

    //Кнопка для перезаписи всей информации выбранного ПК
    $('#switchDel').on('click', function () {
        let index = $('#pkTable').find('tr.success').data('index');
        let selected = $('#pkSelect').val();
        let oldData = [];
        let counter = 0;
        let emptyRow = {len: 0, line: 12, num: 0, plus: false, start: 0, tf: 0};

        if (getSelectedRowData('pkTable', 'sts') === undefined) return;

        $('#pkTable tbody tr').each(function () {
            oldData.push(getSelectedRowData('pkTable', 'sts', counter++));
        });

        counter = 0;
        oldData.splice(index, 1);
        oldData.push(emptyRow);
        oldData.forEach(rec => {
            rec.line = ++counter;
        });

        setDK[selected].sts = oldData;

        pkTabFill('pkTable');
    });

//Функционал кнопок на вкладке "Сут. карты"
    //Кнопка для добавления новой строки
    $('#skAddButton').on('click', function () {
        let index = $('#skTable').find('tr.success').data('index');
        let selected = $('#mapNum').val();
        let oldData = [];
        let counter = 0;

        if (getSelectedRowData('skTable', 'lines') === undefined) return;

        $('#skTable tbody tr').each(function () {
            oldData.push(getSelectedRowData('skTable', 'lines', counter++));
        });

        counter = 0;
        oldData.splice(index, 0, Object.assign({}, oldData[index]));
        oldData.pop();
        oldData.forEach(rec => {
            rec.line = ++counter;
        });

        daySets[selected].lines = oldData;

        newTableFill('skTable', skTableFlag);
    });

    //Кнопка для удаления строки
    $('#skSubButton').on('click', function () {
        let index = $('#skTable').find('tr.success').data('index');
        let selected = $('#mapNum').val();
        let oldData = [];
        let counter = 0;
        let emptyRow = {npk: 0, hour: 0, min: 0, line: 12};

        if (getSelectedRowData('skTable', 'lines') === undefined) return;

        $('#skTable tbody tr').each(function () {
            oldData.push(getSelectedRowData('skTable', 'lines', counter++));
        });

        counter = 0;
        oldData.splice(index, 1);
        oldData.push(emptyRow);
        oldData.forEach(rec => {
            rec.line = ++counter;
        });

        daySets[selected].lines = oldData;

        newTableFill('skTable', skTableFlag);
    });

    //Кнопка для копирования суточной карты
    $('#skCopyButton').on('click', function () {
        let selected = $('#mapNum').val();
        copyArray = Object.assign({}, daySets[selected]);
    });

    //Кнопка для перезаписи суточной карты
    $('#skPasteButton').on('click', function () {
        let selected = $('#mapNum').val();
        daySets[selected] = Object.assign({}, copyArray);
        newTableFill('skTable', skTableFlag);
    });

    //Кнопка для загрузки исходных данных
    $('#skReloadButton').on('click', function () {
        skTabFill(unmodifiedData, false);
    });

    //Кнопка для обнуления текущей карты
    $('#skNewButton').on('click', function () {
        let selected = $('#mapNum').val();
        data.state.arrays.DaySets.daysets[selected].lines.forEach(row => {
            row.npk = 0;
            row.hour = 0;
            row.min = 0;
        });
        skTabFill(data, false);
    });

//Функционал кнопок на вкладке "Нед. карты"
    //Кнопка для копирования строки
    $('#nkCopyButton').on('click', function () {
        copyArray = getSelectedRowData('nkTable', 'days').slice();
    });

    //Кнопка для перезаписи строки
    $('#nkPasteButton').on('click', function () {
        let index = $('#nkTable').find('tr.success').data('index');
        if (getSelectedRowData('nkTable', 'days') === undefined) return;
        weekSets[index].days = copyArray.slice();
        tableFill(weekSets, 'nkTable', nkTableFlag);
    });

    //Кнопка для загрузки исходных данных
    $('#nkReloadButton').on('click', function () {
        nkTabFill(unmodifiedData);
    });

//Функционал кнопок на вкладке "Карта года"
    //Кнопка для копирования строки
    $('#gkCopyButton').on('click', function () {
        copyArray = getSelectedRowData('gkTable', 'days').slice();
    });

    //Кнопка для перезаписи строки
    $('#gkPasteButton').on('click', function () {
        let index = $('#gkTable').find('tr.success').data('index');
        if (getSelectedRowData('gkTable', 'days') === undefined) return;
        monthSets[index].days = copyArray.slice();
        tableFill(monthSets, 'gkTable', gkTableFlag);
    });

    //Кнопка для загрузки исходных данных
    $('#gkReloadButton').on('click', function () {
        gkTabFill(unmodifiedData);
    });

//Функционал кнопок на вкладке "Контроль входов"

    //Кнопка для копирования строки
    $('#kvCopyButton').on('click', function () {
        copyArray = Object.assign({}, getSelectedRowData('kvTable'));
    });

    //Кнопка для перезаписи строки
    $('#kvPasteButton').on('click', function () {
        let index = $('#kvTable').find('tr.success').data('index');
        if (getSelectedRowData('kvTable') === undefined) return;
        stageSets[index] = Object.assign({}, copyArray);
        newTableFill('kvTable', kvTableFlag);
    });

    //Кнопка для загрузки исходных данных
    $('#kvReloadButton').on('click', function () {
        kvTabFill(unmodifiedData);
    });

    //Выбор строк в таблицах по клику
    $('#pkTable').on('click-row.bs.table', function (e, row, $element) {
        $('.success').removeClass('success');
        $($element).addClass('success');
        colorizeSelectedRow('pkTable');
    });
    $('#skTable').on('click-row.bs.table', function (e, row, $element) {
        $('.success').removeClass('success');
        $($element).addClass('success');
        colorizeSelectedRow('skTable');
    });
    $('#nkTable').on('click-row.bs.table', function (e, row, $element) {
        $('.success').removeClass('success');
        $($element).addClass('success');
        colorizeSelectedRow('nkTable');
    });
    $('#gkTable').on('click-row.bs.table', function (e, row, $element) {
        $('.success').removeClass('success');
        $($element).addClass('success');
        colorizeSelectedRow('gkTable');
    });
    $('#kvTable').on('click-row.bs.table', function (e, row, $element) {
        $('.success').removeClass('success');
        $($element).addClass('success');
        colorizeSelectedRow('kvTable');
    });

    //Функционирование выбора СК и ПК
    $('#mapNum').on('change keyup', function () {
        newTableFill('skTable', skTableFlag);
    });

    $('#pkSelect').on('change keyup', function () {
        pkTabFill('pkTable');
    });

    //Функционирование карты для выбора координат
    ymaps.ready(function () {
        //Создание и первичная настройка карты
        let map = new ymaps.Map('map', {
            center: [points.Y, points.X],
            zoom: 18
        });

        map.events.add('click', function (e) {
            if (!map.balloon.isOpen()) {
                let coords = e.get('coords');
                points.Y = coords[0].toPrecision(9);
                points.X = coords[1].toPrecision(9);
                map.balloon.open(coords, {
                    contentHeader: 'Светофор появится на этом месте карты!',
                    contentBody: '<p>Щелкните на крестик в левом верхнем углу</p>'
                });
            } else {
                map.balloon.close();
            }
        });

    })
});

//Набор функций корректной работы АРМ

//Функция для загрузки данных с сервера
function loadData(newData, firstLoadFlag) {

    console.log(newData);
    points = newData.cross.points;
    data = newData;
    unmodifiedData = JSON.parse(JSON.stringify(data));


    $('#table').bootstrapTable('removeAll');
    $('#pkTable').bootstrapTable('removeAll');
    $('#skTable').bootstrapTable('removeAll');
    $('#nkTable').bootstrapTable('removeAll');
    $('#gkTable').bootstrapTable('removeAll');
    $('#vvTable').bootstrapTable('removeAll');
    $('#vv2Table').bootstrapTable('removeAll');
    $('#kvTable').bootstrapTable('removeAll');

    if (firstLoadFlag) {
        $('#forceSendButton').on('click', function () {
            controlSend(newData.state.idevice);
            disableControl('forceSendButton', false);
        });
    }

    //Основная вкладка
    mainTabFill(data, firstLoadFlag);

    //Вкладка ПК
    pkTabFill2(data, firstLoadFlag);

    //Вкладка сут. карт
    skTabFill(data, firstLoadFlag);

    //Вкладка нед. карт
    nkTabFill(data);

    //Вкладка карт года
    gkTabFill(data);

    //Вкладка внеш. входов
    vvTabFill(firstLoadFlag);

    //Вкладка контроля входов
    kvTabFill();

    setIDs();

    $('td > input').on('click', function (element) {
        recentlyClicked = Number(element.target.id);
    });
}

function setIDs(table) {
    if (table === undefined) {
        table = document;
    } else {
        table = $('#nav-tab').find('a[aria-selected=true]')[0].id;
    }
    let counter = 0;
    $(table).find('table').each(function () {
        $(this).find('tbody').find('tr').each(function () {
            $(this).find('td').each(function () {
                $(this).find('input').each(function () {
                    $($(this)[0]).attr('id', counter++);
                    $(this).on('keydown', function (event) {
                        handleKeyboard(event);
                    });
                });
            });
        });
    });
}

function handleKeyboard(key) {
    let tab = $('#nav-tab').find('a[aria-selected=true]')[0].id;
    let retValue;
    switch (key["which"]) {
        case 37: // left
            if (boundsCalc(tab, Number(recentlyClicked), 'left').ret) return;
            $('#' + (Number(recentlyClicked) - 1).toString()).focus();
            recentlyClicked -= 1;
            break;

        case 38: // up
            retValue = boundsCalc(tab, Number(recentlyClicked), 'up');
            if (retValue.ret) return;
            $('#' + (Number(recentlyClicked) - retValue.shift).toString()).focus();
            recentlyClicked -= retValue.shift;
            break;

        case 9:
        case 13:
        case 39: // right
            if (boundsCalc(tab, Number(recentlyClicked), 'right').ret) return;
            $('#' + (Number(recentlyClicked) + 1).toString()).focus();
            recentlyClicked = Number(recentlyClicked) + 1;
            break;

        case 40: // down
            retValue = boundsCalc(tab, Number(recentlyClicked), 'down');
            if (retValue.ret) return;
            $('#' + (Number(recentlyClicked) + retValue.shift).toString()).focus();
            recentlyClicked = Number(recentlyClicked) + retValue.shift;
            break;

        default:
            return; // exit this handler for other keys
    }
    key.preventDefault();
}

function boundsCalc(tab, clicked, key) {
    let retValue = {ret: true, shift: 0};
    let min;
    let max;
    switch (tab) {
        case 'nav-main-tab':
            if (key === 'left') (clicked === 0) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 5) ? retValue.ret = true : retValue.ret = false;
            break;
        case 'nav-pk-tab':
            if (key === 'left') (clicked === 6) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 53) ? retValue.ret = true : retValue.ret = false;
            if (key === 'up') {
                if ((clicked > 9)) {
                    retValue.ret = false;
                    retValue.shift = 4;
                }
            }
            if (key === 'down') {
                if ((clicked < 50)) {
                    retValue.ret = false;
                    retValue.shift = 4;
                }
            }
            break;
        case 'nav-sk-tab':
            if (key === 'left') (clicked === 54) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 89) ? retValue.ret = true : retValue.ret = false;
            if (key === 'up') {
                if ((clicked > 56)) {
                    retValue.ret = false;
                    retValue.shift = 3;
                }
            }
            if (key === 'down') {
                if ((clicked < 87)) {
                    retValue.ret = false;
                    retValue.shift = 3;
                }
            }
            break;
        case 'nav-nk-tab':
            if (key === 'left') (clicked === 90) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 173) ? retValue.ret = true : retValue.ret = false;
            if (key === 'up') {
                if ((clicked > 96)) {
                    retValue.ret = false;
                    retValue.shift = 7;
                }
            }
            if (key === 'down') {
                if ((clicked < 167)) {
                    retValue.ret = false;
                    retValue.shift = 7;
                }
            }
            break;
        case 'nav-gk-tab':
            if (key === 'left') (clicked === 174) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 545) ? retValue.ret = true : retValue.ret = false;
            if (key === 'up') {
                if ((clicked > 204)) {
                    retValue.ret = false;
                    retValue.shift = 31;
                }
            }
            if (key === 'down') {
                if ((clicked < 515)) {
                    retValue.ret = false;
                    retValue.shift = 31;
                }
            }
            break;
        case 'nav-vv-tab':
            max = 586;
            if (data.state.arrays.SetTimeUse.uses.length === 8) {
                max += 40;
            }
            if (key === 'left') (clicked === 546) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 643) ? retValue.ret = true : retValue.ret = false;
            if (key === 'up') {
                if ((clicked > 550)) {
                    retValue.ret = false;
                    retValue.shift = 5;
                }
            }
            if (key === 'down') {
                if ((clicked < max)) {//626
                    retValue.ret = false;
                    retValue.shift = 5;
                }
            }
            break;
        case 'nav-kv-tab':
            min = 597;
            max = 622;
            if (data.state.arrays.SetTimeUse.uses.length === 8) {
                min += 40;
                max += 40;
            }
            if (key === 'left') (clicked === 644) ? retValue.ret = true : retValue.ret = false;
            if (key === 'right') (clicked === 675) ? retValue.ret = true : retValue.ret = false;
            if (key === 'up') {
                if ((clicked > min)) {//637
                    retValue.ret = false;
                    retValue.shift = 4;
                }
            }
            if (key === 'down') {
                if ((clicked < max)) {//662
                    retValue.ret = false;
                    retValue.shift = 4;
                }
            }
            break;
    }
    return retValue;
}

//Заполнение вкладки "Основные"
function mainTabFill(data, firstLoadFlag) {
    let state = data.state;
    for (let area in data.areaMap) {
        if (firstLoadFlag) $('#area').append(new Option(data.areaMap[area], area));
    }
    $('#id').val(data.state.id).on('change', function () {
        checkNew();
    });
    if (firstLoadFlag) setChange('id', 'input', '', numberFlag);
    $('#idevice').val(data.state.idevice).on('change', function () {
        checkNew();
    });
    if (firstLoadFlag) setChange('idevice', 'input', '', numberFlag);
    $('#area option[value=' + data.state.area + ']').attr('selected', 'selected');
    if (firstLoadFlag) setChange('area', 'select', '');
    $('#subarea').val(data.state.subarea);
    if (firstLoadFlag) setChange('subarea', 'input', '', numberFlag);
    $('#name').val(data.state.name);
    if (firstLoadFlag) setChange('name', 'input', '', !numberFlag);
    $('#phone').val(data.state.phone);
    if (firstLoadFlag) setChange('phone', 'input', '', !numberFlag);
    $('#tz').val(data.state.arrays.timedev.tz);
    if (firstLoadFlag) setChange('tz', 'input', 'arrays.timedev', numberFlag);
    $('#summer').prop('checked', data.state.arrays.timedev.summer);
    if (firstLoadFlag) setChange('summer', 'checkbox', 'arrays.timedev', !numberFlag, !longPathFlag);

    anotherTableFill('table', mainTableFlag);
}

//Заполнение вкладки "ПК"
function pkTabFill2(newData, firstLoadFlag) {
    setDK = JSON.parse(JSON.stringify(newData)).state.arrays.SetDK.dk;

    if (firstLoadFlag) {
        setChange('tc', 'input', 'arrays.SetDK.dk', numberFlag, longPathFlag);
        setChange('twot', 'checkbox', 'arrays.SetDK.dk', !numberFlag, longPathFlag);
        setChange('shift', 'input', 'arrays.SetDK.dk', numberFlag, longPathFlag);
        setChange('tpu', 'select', 'arrays.SetDK.dk', !numberFlag, longPathFlag);
        setChange('razlen', 'checkbox', 'arrays.SetDK.dk', !numberFlag, longPathFlag);
    }
    pkTabFill('pkTable');
}

//Заполнение вкладки "Сут. карты"
function skTabFill(newData, firstLoadFlag) {
    daySets = newData.state.arrays.DaySets.daysets;
    daySets.forEach(daySet => {
        if (firstLoadFlag) $('#mapNum').append(new Option(daySet.num, daySet.num - 1));
    });

    newTableFill('skTable', skTableFlag);
}

//Заполнение вкладки "Нед. карты"
function nkTabFill(newData) {
    weekSets = newData.state.arrays.WeekSets.wsets;
    tableFill(weekSets, 'nkTable', nkTableFlag);
}

//Заполнение вкладки "Карты года"
function gkTabFill(newData) {
    monthSets = newData.state.arrays.MonthSets.monthset;
    tableFill(monthSets, 'gkTable', gkTableFlag);
}

//Заполнение вкладки "Внеш. входы"
function vvTabFill(firstLoadFlag) {
    anotherTableFill('vvTable', vvTableFlag);

    $('#ite').val(data.state.arrays.SetTimeUse.ite);
    if (firstLoadFlag) setChange('ite', 'input', 'arrays.SetTimeUse', numberFlag);
    $('#tuin').val(data.state.arrays.SetTimeUse.tuin);
    if (firstLoadFlag) setChange('tuin', 'input', 'arrays.SetTimeUse', numberFlag);

    tableFill([0], 'vv2Table', vv2TableFlag);
}

//Заполнение вкладки "Контроль входов"
function kvTabFill() {
    stageSets = data.state.arrays.SetCtrl.Stage;
    newTableFill('kvTable', kvTableFlag);
}


//Функция для заполнения таблиц недельных карт, годовых карт и длительности МГР при неисправности ДТ
function tableFill(set, table, staticFlag) {
    $('#' + table).bootstrapTable('removeAll');
    set.forEach(function () {
        $('#' + table).bootstrapTable('append', '');
    });

    let counter = -1;
    $('#' + table + ' tbody tr').each(function () {
        let dayCounter = 0;
        counter++;
        $(this).find('td').each(function () {
            if (dayCounter++ === 0) {
                $(this).append(staticFlag ? 'Интервал,с' : set[counter].num);
            } else {
                $(this).append(
                    '<input class="form-control border-0"' +
                    'style="min-width: 43px; max-width: 43px;" name="number" type="number" value="' +
                    (staticFlag ? data.state.arrays.SetTimeUse.notwork[dayCounter - 2] : set[counter].days[dayCounter - 2]) + '" required/>'
                );
            }
        })
    });

    $('#' + table + ' thead tr th').each(function () {
        $(this).attr('style', 'text-align: center; min-width: 45px;')
    });

    if (firstLoad) tableChange(set, table, !staticFlag);
    renewal(table, true);
}

//Функция для сохранения изменений в вышеперечисленных таблицах
function tableChange(set, table, daysFlag) {
    $('#' + table).on('change', function () {
        let counter = 0;
        $('#' + table + ' tbody tr').each(function () {
            let setArr = [];
            $(this).find('td').each(function () {
                let value = Number($(this).find('input').val());
                if (!isNaN(value)) setArr.push(value);
            });
            (daysFlag) ? set[counter++].days = setArr : data.state.arrays.SetTimeUse.notwork = setArr;
        });
        console.log(data.state);
    })
}

//Функция для заполнения таблиц параметров ДК и использования внешних входов
function anotherTableFill(table, tableFlag) {
    $('#' + table).bootstrapTable('removeAll')
        .bootstrapTable('append', (tableFlag ? data.state.arrays.SetupDK : data.state.arrays.SetTimeUse.uses));
    $('#' + table + ' tbody tr').each(function () {
        let counter = 0;
        $(this).find('td').each(function () {
            let value = $(this).text();
            let type = tableFlag ? 'number' : (((counter === 0) || (counter === 4)) ? 'text' : 'number');
            $(this).text('');
            if ((counter === 0) && (!tableFlag)) {
                $(this).append(value);
            } else {
                $(this).append('<input class="form-control border-0" type="' + type + '" value="' + value + '" />');
            }
            counter++;
        });
    });
    if (firstLoad) anotherTableChange(table, tableFlag);
    renewal(table);
}

//Функция для сохранения изменений в вышеперечисленных таблицах
function anotherTableChange(table, tableFlag) {
    $('#' + table).on('change', function () {
        let names = [];
        $('#' + table + ' thead th').each(function () {
            names.push($(this).attr('data-field'));
        });
        let recCounter = 0;
        $('#' + table + ' tbody tr').each(function () {
            let counter = -1;
            $(this).find('td').each(function () {
                let value = $(this).find('input').val();
                if (tableFlag) {
                    data.state.arrays.SetupDK[names[++counter]] = Number(value);
                } else {
                    data.state.arrays.SetTimeUse.uses[recCounter][names[++counter]] = ((counter === 0) || (counter === 4)) ? value : Number(value);
                }
            });
            recCounter++;
        });
        console.log(data.state);
    });
}

//Функция для заполнения таблиц суточных карт и контроля входов
function newTableFill(table, tableFlag) {
    let selected = $('#mapNum').val();
    if (skFlag || kvFlag) {
        if (firstLoad) (tableFlag ? skTableChange(table) : kvTableChange(table));
    }
    $('#' + table).bootstrapTable('removeAll');
    let set = (tableFlag ? daySets : stageSets);

    set.forEach(function () {
        $('#' + table).bootstrapTable('append', '');
    });

    let counter = -1;
    $('#' + table + ' tbody tr').each(function () {
        let dayCounter = 0;
        let endFlag = false;
        counter++;
        $(this).find('td').each(function () {
            let prevArr = [];
            if (counter > 0) prevArr = (tableFlag ? daySets[selected].lines[counter - 1] : stageSets[counter - 1]);
            let currArr = (tableFlag ? daySets[selected].lines[counter] : stageSets[counter]);
            if (tableFlag) {
                if ((prevArr.hour === 24) && (prevArr.min === 0)) endFlag = true;
            } else if (counter > 0) {
                if ((prevArr.end.hour === 24) && (prevArr.end.min === 0)) endFlag = true;
            }
            switch (dayCounter++) {
                case 0 :
                    $(this).append(counter + 1);
                    break;
                case 1 :
                    if (endFlag) {
                        $(this.append('00:00'));
                    } else {
                        $(this).append(((counter === 0) ? '00' : handsomeNumbers((tableFlag ? prevArr.hour : prevArr.end.hour))) + ':' +
                            ((counter === 0) ? '00' : handsomeNumbers((tableFlag ? prevArr.min : prevArr.end.min))));
                    }
                    break;
                case 2 :
                    $(this).append(
                        '<div class="container"><div class="row"><input class="form-control border-0 col-md-5" ' +
                        'style="max-width: 45px;" name="number" type="number" value="' +
                        handsomeNumbers((tableFlag ? currArr.hour : currArr.end.hour)) + '" required/> ' + '<div style="margin-top: 6px;">:</div>' +
                        '<input class="form-control border-0 col-md-5" style="max-width: 45px;" name="number" ' +
                        'type="number" value="' + handsomeNumbers((tableFlag ? currArr.min : currArr.end.min)) +
                        '" required/></div></div>'
                    );
                    break;
                case 3 :
                    $(this).append(
                        '<input class="form-control border-0" name="number" type="number" ' +
                        'style="max-width: 50px;" value="' + (tableFlag ? currArr.npk : currArr.lenTVP) + '"/>'
                    );
                    break;
                case 4 :
                    $(this).append(
                        '<input class="form-control border-0" name="number" type="number" ' +
                        'style="max-width: 50px;" value="' + currArr.lenMGR + '"/>'
                    );
                    break;
            }
        });
    });
    tableFlag ? renewal(table) : renewal(table, true);
}

//Функция для сохранения изменений в таблице суточных карт, а также заполнение столбца "T начала"
function skTableChange(table) {
    $('#' + table).on('change', function () {
        let selected = $('#mapNum').val();
        let tableData = [];
        $('#' + table + ' tbody tr').each(function () {
            let rec = [];
            $(this).find('td').each(function () {
                $(this).find('input').each(function () {
                    let value = $(this).val();
                    rec.push(Number((value.startsWith('0')) ? value.substring(1, 2) : value));
                })
            });
            tableData.push(rec);
        });
        let counter = 0;
        daySets[selected].lines.forEach(variable => {
            variable.npk = tableData[counter][2];
            variable.hour = tableData[counter][0];
            variable.min = tableData[counter++][1];
        });
        data.state.arrays.DaySets.daysets = daySets;
        newTableFill(table, true);
        console.log(data.state);
    });
    skFlag = false;
}

//Функция для сохранения изменений в таблице контроля входов, а также заполнение столбца "T начала"
function kvTableChange(table) {
    $('#' + table).on('change', function () {
        let tableData = [];
        $('#' + table + ' tbody tr').each(function () {
            let rec = [];
            $(this).find('td').each(function () {
                let text = $(this)[0].innerText;
                if ((text !== ':') && (text !== '')) {
                    if (text.includes(':')) {
                        let time = text.split(':');
                        rec.push(Number(time[0]));
                        rec.push(Number(time[1]));
                    } else {
                        rec.push(Number(text));
                    }
                }
                $(this).find('input').each(function () {
                    let value = $(this).val();
                    rec.push(Number(((value.startsWith('0')) && (value.length > 1)) ? value.substring(1, 2) : value));
                })
            });
            tableData.push(rec);
        });
        let counter = 0;
        stageSets.forEach(variable => {
            variable.line = tableData[counter][0];
            variable.start.hour = tableData[counter][1];
            variable.start.min = tableData[counter][2];
            variable.end.hour = tableData[counter][3];
            variable.end.min = tableData[counter][4];
            variable.lenTVP = tableData[counter][5];
            variable.lenMGR = tableData[counter++][6];
        });
        data.state.arrays.SetCtrl.Stage = stageSets;
        newTableFill(table, false);
        console.log(data.state);
    });
    kvFlag = false;
}

//Функция для заполнения вкладки ПК
function pkTabFill(table) {
    let selected = $('#pkSelect').val();
    let currPK = setDK[selected];

    $('#' + table).bootstrapTable('removeAll');

    if (pkFlag) {
        pkTableChange(table, currPK);
    }

    $('#tc').val(currPK.tc);
    $('#twot').prop('checked', currPK.twot);
    $('#shift').val(currPK.shift);
    $('#tpu').find('option').each(function () {
        $(this).removeAttr('selected');
    });
    $('#tpu option[value="' + currPK.tpu + '"]').attr('selected', 'selected');
    $('#razlen').prop('checked', currPK.razlen);

    currPK.sts.forEach(function () {
        $('#' + table).bootstrapTable('append', '');
    });

    let counter = -1;
    $('#' + table + ' tbody tr').each(function () {
        let dayCounter = 0;
        counter++;
        $(this).find('td').each(function () {
            let record = currPK.sts[counter];
            switch (dayCounter++) {
                case 0 :
                    $(this).append(record.line);
                    break;
                case 1 :
                    $(this).append(
                        '<input class="form-control border-0" name="number" type="number" ' +
                        'style="max-width: 50px;" value="' + record.start + '"/>'
                    );
                    break;
                case 2 :
                    $(this).append(
                        '<select>' +
                        '<option value="0"> </option>' +
                        '<option value="1">МГР</option>' +
                        '<option value="2">1 ТВП</option>' +
                        '<option value="3">2 ТВП</option>' +
                        '<option value="4">1,2 ТВП</option>' +
                        '<option value="5">Зам. 1ТВП</option>' +
                        '<option value="6">Зам. 2ТВП</option>' +
                        '<option value="7">Зам.</option>' +
                        '<option value="8">МДК</option>' +
                        '<option value="9">ВДК</option>' +
                        '</select>'
                    );
                    $(this).find('select').find('option').each(function () {
                        $(this).removeAttr('selected');
                    });
                    $(this).find('option[value="' + record.tf + '"]').attr('selected', 'selected');
                    break;
                case 3 :
                    $(this).append(
                        '<input class="form-control border-0" name="number" type="number" ' +
                        'style="max-width: 50px;" value="' + record.num + '"/>'
                    );
                    break;
                case 4 :
                    $(this).attr('class', 'justify-content-center');
                    $(this).append(
                        '<input class="form-control border-0" name="number" type="number" ' +
                        'style="max-width: 50px;" value="' + record.stop + '"/>'
                    );
                    break;
                case 5 :
                    $(this).append(
                        '<input class="form-control border-0" name="text" type="text" ' +
                        'style="max-width: 70px;" value="' + (record.plus ? '+' : '') + '"/>'
                    );
                    break;
            }
        });
    });
    renewal(table);
}

//Функция для сохранения изменений в таблице ПК
function pkTableChange(table, currPK) {
    $('#' + table).on('change', function () {
        let counter = -1;
        let selected = Number($('#pkSelect').val());
        $('#' + table + ' tbody tr').each(function () {
            let dayCounter = 0;
            counter++;
            $(this).find('td').each(function () {
                switch (dayCounter++) {
                    case 1 :
                        currPK.sts[counter].start = Number($(this).find('input').val());
                        break;
                    case 2 :
                        currPK.sts[counter].tf = Number($(this).find('select').val());
                        break;
                    case 3 :
                        currPK.sts[counter].num = Number($(this).find('input').val());
                        break;
                    case 4 :
                        currPK.sts[counter].stop = Number($(this).find('input').val());
                        break;
                    case 5 :
                        currPK.sts[counter].plus = ($(this).find('input').val() === '+');
                        break;
                }
            });
        });
        setDK[selected] = currPK;
//        console.log(data.state.arrays.SetDK.dk[selected].sts);
        data.state.arrays.SetDK.dk[selected] = setDK[selected];
    });
    pkFlag = false;
}

//Функция для преобразования вида цифр
function handsomeNumbers(num) {
    if (num.toString().length >= 2) return num;
    return (num < 10) ? '0' + num : num;
}

//Функция для сохранения изменений всех не табличных элементов
function setChange(element, type, fullPath, numFlag, hardFlag) {
    if (!firstLoad) return;
    let path = fullPath.split('.');
    if (path[1] !== undefined) {
        if (type === 'input') {
            $('#' + element).on('change', function () {
                if (numFlag) {
                    hardFlag ? data.state[path[0]][path[1]][path[2]][$('#pkSelect').val()][element] = Number($('#' + element).val())
                        : data.state[path[0]][path[1]][element] = Number($('#' + element).val());
                } else {
                    data.state[path[0]][path[1]][element] = $('#' + element).val();
                }
            });
        }
        if (type === 'select') {
            $('#' + element).on('change keyup', function () {
                hardFlag ? data.state[path[0]][path[1]][path[2]][$('#pkSelect').val()][element] = Number($('#' + element + ' option:selected').val())
                    : data.state[path[0]][path[1]][element] = Number($('#' + element + ' option:selected').val());
            });
        }
        if (type === 'checkbox') {
            $('#' + element).on('change', function () {
                hardFlag ? data.state[path[0]][path[1]][path[2]][$('#pkSelect').val()][element] = $('#' + element).prop('checked')
                    : data.state[path[0]][path[1]][element] = $('#' + element).prop('checked');
            });
        }
    } else {
        if (type === 'input') {
            $('#' + element).on('change', function () {
                if (numFlag) {
                    data.state[element] = Number($('#' + element).val());
                } else {
                    data.state[element] = $('#' + element).val();
                }
            });
        }
        if (type === 'select') {
            $('#' + element).on('change keyup', function () {
                data.state[element] = Number($('#' + element + ' option:selected').val());
            });
        }
        if (type === 'checkbox') {
            $('#' + element).on('change', function () {
                data.state[element] = $('#' + element).prop('checked');
            });
        }
    }
}

//Отображение кнопки для выбора координат и разблокирование кнопки создания нового перекрёстка
function checkNew() {
    let buttonClass = $('#addButton')[0].className.toString();
    if ((Number($('#id').val()) !== unmodifiedData.state.id) && (Number($('#idevice').val()) !== unmodifiedData.state.idevice)) {
        if (buttonClass.indexOf('disabled') !== -1) buttonClass = buttonClass.substring(0, buttonClass.length - 9);
        if (!$('#chooseCoordinates').length) {
            $('#forCoordinates').append('<div class="col-xs-8 ml-1">' +
                '<button type="button" class="btn btn-light ml-5 justify-content-center border" id="chooseCoordinates" style="">Выберите координаты</button>' +
                '</div>');
            chooseCoordinates();
        }
    } else {
        if (buttonClass.indexOf('disabled') === -1) buttonClass = buttonClass.concat(' disabled');
    }
    $('#addButton')[0].className = buttonClass;
}

//Отображение кнопки для выбора координат и разблокирование кнопки создания нового перекрёстка
function disableControl(button, status) {
    let buttonClass = $('#' + button)[0].className.toString();
    if (status) {
        if (buttonClass.indexOf('disabled') !== -1) buttonClass = buttonClass.substring(0, buttonClass.length - 9);
    } else {
        if (buttonClass.indexOf('disabled') === -1) buttonClass = buttonClass.concat(' disabled');
    }
    $('#' + button)[0].className = buttonClass;
}

//Открытие карты с выбором координат
function chooseCoordinates() {
    $('#chooseCoordinates').on('click', function () {
        $('#mapModal').attr('style', 'display : block;');
    });

    $('.close').on('click', function () {
        $('#mapModal').attr('style', 'display : none;');
    });

    window.onclick = function (event) {
        if (event.target === $('#mapModal')) {
            $('#mapModal').attr('style', 'display : none;');
        }
    }
}

//Отправка выбранной команды на сервер
function controlSend(id) {
    $.ajax({
        url: window.location.origin + window.location.pathname.substring(0, window.location.pathname.length - 8) + '/DispatchControlButtons',
        type: 'post',
        dataType: 'json',
        contentType: 'application/json',
        success: function (data) {
            console.log(data);
        },
        data: JSON.stringify({id: id, cmd: 1, param: 0}),
        error: function (request) {
            console.log(request.status + ' ' + request.responseText);
        }
    });
}