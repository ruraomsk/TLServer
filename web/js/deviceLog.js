// let selected;
let devices;
let IDs = [];
let regionInfo;
let areaInfo;
// let timeStart = '';
// let timeEnd = '';

function sortByTime(a, b) {
    let aName = a.time.toLowerCase();
    let bName = b.time.toLowerCase();
    return ((aName < bName) ? 1 : ((aName > bName) ? -1 : 0));
}

$(function () {
    // $('#sendButton').on('click', function () {
    $.ajax({
        url: window.location.href,
        type: 'post',
        dataType: 'json',
        contentType: 'application/json',
        success: function (data) {
            regionInfo = Object.assign({}, data.regionInfo);
            areaInfo = Object.assign({}, data.areaInfo);
            devices = data.devices.slice();
            let counter = 0;
            data.devices.forEach(device => {
                let region = regionInfo[devices[counter].region];
                let area = areaInfo[region][devices[counter].area];
                device.region = region;
                device.area = area;
                IDs.push({ID: device.ID, description: device.description});
                devices[counter++].ID = '';
            });
            console.log(data);

            $('#table')
                .bootstrapTable('removeAll')
                .bootstrapTable('append', data.devices)
                .bootstrapTable('scrollTo', 'top')
                .bootstrapTable('refresh', {
                    data: data.devices
                });
            // console.log(data.result);
            // disableControl('forceSendButton', true);
            // disableControl('sendButton', false);
            counter = 0;
            $('.fixed-table-toolbar').each(function () {
                $(this).attr('id', 'toolbar' + counter++)
            });
            $('#toolbar0').append('<button id="timeButton1" class="btn btn-secondary ml-4 mt-2">30 минут</button>')
                .append('<button id="getLog" class="btn btn-secondary ml-4 mt-2">6 часов</button>')
                .append('<button id="timeButton2" class="btn btn-secondary ml-4 mt-2">Выбранное время</button>')
                .append('<div class="row mt-3">' +
                        '   <label class="col-md-1" for="timeStart">С</label>\n' +
                        '   <div class="col-md-4 mr-3"><input id="timeStart" type="datetime-local"></div>' +
                        '   <label class="col-md-1 ml-3" for="timeEnd"> до</label>\n' +
                        '   <div class="col-md-4"><input id="timeEnd" type="datetime-local"></div>' +
                        '</div>');
            let now = new Date();
            now = new Date(now.getTime() - (now.getTimezoneOffset() * 60 * 1000));
            $('#timeStart').attr('value', (now.toISOString().substring(0, now.toISOString().length - 1)));
            $('#timeEnd').attr('value', (now.toISOString().substring(0, now.toISOString().length - 1)));
            $('#getLog').on('click', function () {
                // let now = new Date();
                // now = new Date(now.getTime() - (now.getTimezoneOffset() * 60 * 1000));
                let timeStart = timeCalc(now, 360 * 60);
                let timeEnd = now.toISOString();
                getLogs(timeStart, timeEnd);
            });
            $('#timeButton1').on('click', function () {
                // let now = new Date();
                // now = new Date(now.getTime() - (now.getTimezoneOffset() * 60 * 1000));
                let timeStart = timeCalc(now, 30 * 60);
                let timeEnd = now.toISOString();
                getLogs(timeStart, timeEnd);
            });
            $('#timeButton2').on('click', function () {
                // let now = new Date();
                // now = new Date(now.getTime() - (now.getTimezoneOffset() * 60 * 1000));
                let timeStart = $('#timeStart')[0].value;
                let timeEnd = $('#timeEnd')[0].value;
                getLogs(timeStart + 'Z', timeEnd + 'Z');
            });
        },
        // data: JSON.stringify(data.state),
        error: function (request) {
            console.log(request.status + ' ' + request.responseText);
        }
    });
    // });
});

function getLogs(start, end) {
    // {ID: '', area: '', region: ''}
    console.log('start:' + start + '   end' + end);
    let toSend = {devices: [], timeStart: start, timeEnd: end};
    let selected = $('#table').bootstrapTable('getSelections');
    selected.forEach(cross => {
        toSend.devices.push({
            ID: findIdByDescription(cross.description),
            area: getAreaNum(cross.region, cross.area),
            region: getRegionNum(cross.region),
            description: cross.description
        });
    });

    if (!selected.length) {
        $('#toolbar0').append('<div style="color: red;" id="toolbar0Msg"><h5>Выберите устройства</h5></div>');
        return;
    } else {
        $('#toolbar0Msg').remove();
    }
    //Отправка на сервер запроса проверки данных
    $.ajax({
        type: 'POST',
        url: window.location.href + '/info',
        data: JSON.stringify(toSend),
        dataType: 'json',
        success: function (data) {
            console.log(data);
            let allData = [];
            data.deviceLogs.forEach(log => {
                let date = new Date(log.time);
                // date = new Date(date.getTime() + (date.getTimezoneOffset() * 60 * 1000));
                let localDate = date.toLocaleString('ru-RU', {timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone});
                let text = log.text;
                let description = log.devices.description;
                let add = {cross: description, message: text, time: localDate};
                allData.push(add);
            });
            allData.sort(sortByTime);
            $('#logsTable').bootstrapTable('removeAll')
                .bootstrapTable('append', allData)
                .bootstrapTable('scrollTo', 'top')
                .bootstrapTable('refresh', {
                    data: allData
                });
        },
        error: function (request) {
            // if (!($('#passwordMsg').length)){
            //     $('#passwordForm').append('<div style="color: red;" id="passwordMsg"><h5>Неверный логин и/или пароль</h5></div>');
            // }
            console.log(request.status + ' ' + request.responseText);
        }
    });
}

function timeCalc(now, offset) {
    return (new Date(now.getTime() - (offset * 1000))).toISOString();
}

function findIdByDescription(description) {
    let id;
    IDs.forEach(ID => {
        if (ID.description === description) id = ID.ID;
    });
    return id;
}

//Получение номера региона по описанию
function getRegionNum(region) {
    let num = 0;
    for (let reg in regionInfo) {
        if (regionInfo[reg] === region) num = reg;
    }
    return num;
}

//Получение номера района по описанию
function getAreaNum(region, area) {
    let num = 0;
    for (let ar in areaInfo[region]) {
        if (areaInfo[region][ar] === area) num = ar;
    }
    return num.toString();
}