'use strict';

let jsonData;
let accInfo;
let areaInfo;
let regionInfo;
// let permissionsInfo;
let noDup = false;
let firstLoadFlag = true;

/**
 * @return {number}
 */
function SortByLogin(a, b) {
    let aName = a.login.toLowerCase();
    let bName = b.login.toLowerCase();
    return ((aName < bName) ? -1 : ((aName > bName) ? 1 : 0));
}

$(function () {

    //Закрытие вкладки при закрытии карты
//    window.setInterval(function () {
//        if(localStorage.getItem('maintab') === 'closed') window.close();
//    }, 1000);

    let $table = $('#table');

    //Всплывающее окно для создания пользователя
    $('#addDialog').dialog({
        autoOpen: false,
        buttons: {
            'Подтвердить': function () {
                //Проверка корректности введённых данных
                if (($('#login').val() === '') || ($('#password').val() === '') || ($('#area option:selected').text() === '')) {
                    if (!($('#loginMsg').length) && ($('#login').val() === '')) {
                        $('#loginForm').append('<div style="color: red;" id="loginMsg"><h5>Введите логин</h5></div>');
                    }
                    if (!($('#passwordMsg').length) && ($('#password').val() === '')) {
                        $('#passwordForm').append('<div style="color: red;" id="passwordMsg"><h5>Введите пароль</h5></div>');
                    }
                    if (!($('#areasMsg').length) && ($('#area option:selected').text() === '')) {
                        $('#areasForm').append('<div style="color: red;" id="areasMsg"><h5>Выберите районы</h5></div>');
                    }
                    return;
                }
                let selectedAreas = $('#area option:selected').toArray().map(item => item.value);
                let areas = [];

                selectedAreas.forEach(area => {
                    areas.push({
                        num: area
                    });
                });

                //Сбор данных для отправки на сервер
                let toSend = {
                    login: $('#login').val(),
                    workTime: parseInt($('#workTime option:selected').text()),
                    password: $('#password').val(),
                    role: {name: $('#privileges option:selected').text(), permissions: permissionControl('a')},
                    region: {num: $('#region option:selected').val()},
                    area: areas
                };

                //Отправка данных на сервер
                $.ajax({
                    url: window.location.href + '/add',
                    type: 'post',
                    dataType: 'json',
                    contentType: 'application/json',
                    success: function (data) {
                        console.log(data.msg);
                        $('#addDialog').dialog('close');
                        $table.bootstrapTable('removeAll');
                        getUsers($table);
                    },
                    data: JSON.stringify(toSend),
                    error: function (request) {
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
        close: function () {
            $('#loginMsg').remove();
            $('#passwordMsg').remove();
            $('#areasMsg').remove();
        }
    });

    //Всплывающее окно для изменения пользователя
    $('#updateDialog').dialog({
        autoOpen: false,
        buttons: {
            'Подтвердить': function () {
                //Проверка корректности введённых данных
                if ($('#updateArea option:selected').text() === '') {
                    if (!($('#updateAreasMsg').length)) {
                        $('#updateAreasForm').append('<div style="color: red;" id="updateAreasMsg"><h5>Выберите районы</h5></div>');
                    }
                    return;
                }
                let selectedAreas = $('#updateArea option:selected').toArray().map(item => item.value);
                let areas = [];
                let login = $.map($table.bootstrapTable('getSelections'), function (row) {
                    return row.login;
                });

                selectedAreas.forEach(area => {
                    areas.push({
                        num: area
                    });
                });

                //Сбор данных для отправки на сервер
                let toSend = {
                    login: login[0],
                    role: {name: $('#updatePrivileges option:selected').text(), permissions: permissionControl('u')},
                    region: {num: $('#updateRegion option:selected').val()},
                    area: areas,
                    workTime: parseInt($('#updateWorkTime option:selected').text())
                };

                //Отправка данных на сервер
                $.ajax({
                    url: window.location.href + '/update',
                    type: 'post',
                    dataType: 'json',
                    contentType: 'application/json',
                    success: function (data) {
                        console.log(data.msg);
                        $('#updateDialog').dialog('close');
                        $table.bootstrapTable('removeAll');
                        getUsers($table);
                    },
                    data: JSON.stringify(toSend),
                    error: function (request) {
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
        close: function () {
            $('#updateAreasMsg').remove();
            $('#updateMsg').remove();
        }
    });

    //Добавление нового пользователя
    $('#appendButton').on('click', function () {
        $('#login').val('');
        $('#privileges').val('Viewer');
        $('#password').val('');
        fillAreas();

        $('#addDialog').dialog('open');
    });

    //Изменение существующего пользователя
    $('#updateButton').on('click', function () {
        let currPrivileges;
        let currRegion;
        let currAreas;
        let currWorkTime;

        let login = $.map($table.bootstrapTable('getSelections'), function (row) {
            return row.login;
        });

        if (login[0] === undefined) {
            if (!($('#updateMsg').length)) {
                $('#toolbar').append('<div style="color: red;" id="updateMsg"><h5>Выберите пользователя для изменения</h5></div>');
            }
            return;
        }

        accInfo.forEach(user => {
            if (user.login === login[0]) {
                currPrivileges = user.role;
                currRegion = user.region.num;
                currAreas = user.area;
                currWorkTime = user.workTime;
            }
        });

        //костыль для супера
        if ((currPrivileges.name === 'Admin') && (!window.location.href.includes('Super'))) {
            if (!($('#updateMsg').length)) {
                $('#toolbar').append('<div style="color: red;" id="updateMsg"><h5>Невозможно изменить администратора</h5></div>');
            }
            return;
        }

        currPrivileges.permissions.forEach(perm => {
            $('input[id*="' + perm + '"]').each(function () {
                $(this).prop('checked', true);
            })
        });

        $('#updatePrivileges').val(currPrivileges.name);
        $('#updateRegion').val(currRegion);
        $('#updateArea').val(currAreas);
        $('#updateWorkTime').val(currWorkTime);

        $('#updateDialog').dialog('open');
    });

    //Удаление пользователя
    $('#deleteButton').on('click', function () {
        let login = $.map($table.bootstrapTable('getSelections'), function (row) {
            return row.login;
        });
        let loginToSend = {login: login[0]};
        //Отпрака на сервер запроса на удаление пользователя
        $.ajax({
            url: window.location.href + '/delete',
            type: 'post',
            dataType: 'json',
            contentType: 'application/json',
            success: function (data) {
                console.log(data.msg);
                $table.bootstrapTable('removeAll');
                getUsers($table);
            },
            data: JSON.stringify(loginToSend),
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    //Открытие вкладки с логами сервера
    $('#serverLogButton').on('click', function () {
        openPage('/serverLog');
    });

    //Открыте вкладки с редактируемыми ДК
    $('#editControlButton').on('click', function () {
        openPage('/crossEditControl');
    });

    //Открыте вкладки с проверкой БД
    $('#stateTest').on('click', function () {
        openPage('/stateTest');
    });

    //Обнуление дополнительных прав и смене шаблона роли
    $('#updatePrivileges').change(function () {
        $('input[id*="u"]').each(function () {
            $(this).prop('checked', false);
        });
    });
    $('#privileges').change(function () {
        $('input[id*="u"]').each(function () {
            $(this).prop('checked', false);
        });
    });

    //обновление списка пользовтелей
    getUsers($table);
});

//Загрузка списка пользователей с сервера
function getUsers($table) {
    $.ajax({
        type: 'POST',
        url: window.location.href,
        success: function (data) {
            jsonData = data;
            regionInfo = data.regionInfo;
            areaInfo = data.areaInfo;
            // permissionsInfo = data.permInfo;
            if (data.accInfo !== null) {
                accInfo = data.accInfo.sort(SortByLogin);
                accInfo.forEach(account => {
                    let areas = '';
                    account.area.forEach(area => {
                        areas += area.nameArea + ', ';
                    });
                    let info = [];
                    //Заполнение структуры для дальнейшей записи в таблицу
                    info.push({
                        state: false,
                        login: account.login,
                        privileges: account.role.name,
                        region: account.region.nameRegion,
                        area: (areas !== '') ? areas.substring(0, areas.length - 2) : areas,
                        workTime: account.workTime
                    });
                    $table.bootstrapTable('append', info);
                    $table.bootstrapTable('scrollTo', 'top');
                });
            }

            if (!noDup) {
                document.title = 'Личный кабинет';

                //Заполнение поля выбора регионов для создания пользователя
                for (let reg in regionInfo) {
                    $('#region').append(new Option(regionInfo[reg], reg));
                    $('#updateRegion').append(new Option(regionInfo[reg], reg));
                }

                fillAreas();

                //Заполнение поля выбора прав для создания пользователя
                data.roles.forEach(role => {
                    $('#privileges').append(new Option(role, role));
                    $('#updatePrivileges').append(new Option(role, role));
                });

                noDup = true;
            }
            console.log(data);
            if (firstLoadFlag) {
                for (let perm in data.permInfo) {
                    $('#privilegesForm').append('<div class="col-xs-2 mt-2" style="">\n' +
                        '                            <div class="form-group row ml-2 mr-1">\n' +
                        '                                <div class="col-xs-10 ml-1 pt-1">\n' +
                        '                                    <div class="form-check">\n' +
                        '                                        <label class="form-check-label">\n' +
                        '                                            <input type="checkbox" class="form-check-input" id="' + ('a' + data.permInfo[perm].id) + '" value="">' +
                        '                                            ' + data.permInfo[perm].description + '</label>\n' +
                        '                                    </div>\n' +
                        '                                </div>\n' +
                        '                            </div>\n' +
                        '                        </div>');
                    $('#updatePrivilegesForm').append('<div class="col-xs-2 mt-2" style="">\n' +
                        '                            <div class="form-group row ml-2 mr-1">\n' +
                        '                                <div class="col-xs-10 ml-1 pt-1">\n' +
                        '                                    <div class="form-check">\n' +
                        '                                        <label class="form-check-label">\n' +
                        '                                            <input type="checkbox" class="form-check-input" id="' + ('u' + data.permInfo[perm].id) + '" value="">' +
                        '                                            ' + data.permInfo[perm].description + '</label>\n' +
                        '                                    </div>\n' +
                        '                                </div>\n' +
                        '                            </div>\n' +
                        '                        </div>');
                }
                firstLoadFlag = false;
            }
        },
        error: function (request) {
            console.log(request.status + ' ' + request.responseText);
        }
    });

}

function permissionControl(dialog) {
    let currPrivileges = [];
    $('input[id*="' + dialog + '"]').each(function () {
        if ($(this).prop('checked')) {
            currPrivileges.push(Number($(this)[0].id.substring(1)))
        }
    });

    currPrivileges.forEach(perm => {
        $('input[id*="' + perm + '"]').each(function () {
            let id = Number($(this)[0].id.substring(1));
            let arr = currPrivileges;
            if ($(this).prop('checked')) {
                if (!arr.includes(id)) {
                    arr.push(id);
                }
            } else {
                if (arr.includes(id)) {
                    arr.splice(arr.indexOf(id), 1);
                }
            }
        })
    });
    return currPrivileges;
}

//Заполнение поля выбора районов для создания или изменения пользователя
function fillAreas() {
    $('#area').empty();
    $('#updateArea').empty();
    for (let regAreaJson in areaInfo) {
        for (let areaJson in areaInfo[regAreaJson]) {
            if (regAreaJson === $('#region').find(':selected').text()) {
                $('#area').append(new Option(areaInfo[regAreaJson][areaJson], areaJson));
            }
            if (regAreaJson === $('#updateRegion').find(':selected').text()) {
                $('#updateArea').append(new Option(areaInfo[regAreaJson][areaJson], areaJson));
            }
        }
    }
}

//Функция для открытия новой вкладки
function openPage(url) {
    $.ajax({
        type: 'GET',
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