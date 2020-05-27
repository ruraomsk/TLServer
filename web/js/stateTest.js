'use strict';

let links = [];

$(document).ready(function () {
	let $table = $('#table');
	load($table, true);
	$('#updateButton').on('click', function() {
	    load($table, false);
	});
});

//Загрузка данных о редактируемых ДК в таблицу
function load($table, firstLoadFlag) {
	$.ajax({
		type: 'POST',
		url: window.location.href,
		success: function (data) {
		    console.log(data);
		    links = [];
            let dataArray = [];
            let tempRecord;
            //Заполнение данных для записи в таблицу
            for (let tflight in data.arms) {
                let dk = data.arms[tflight];
                tempRecord = {description : '', open : ''};
                tempRecord.description = dk.description;
                links.push('?Region=' + dk.region + '&Area=' + dk.area + '&ID=' + dk.ID);
//                tempRecord.open = '?Region=' + dk.region + '&Area=' + dk.area + '&ID=' + dk.ID;
                dataArray.push(Object.assign({}, tempRecord));
            }

            $table.bootstrapTable('removeAll');
            $table.bootstrapTable('append', dataArray);
            $table.bootstrapTable('scrollTo', 'top');
            $table.bootstrapTable('refresh', {
                data: dataArray
            });

            makeButtons();
		},
		error: function (request) {
			console.log(request.status + ' ' + request.responseText);
		}
	});
	if(firstLoadFlag) {
	    $('.fixed-table-toolbar').append('<button id="updateButton" class="btn btn-secondary mb-2">Обновить</button>');
    }
}

function makeButtons() {
    let counter = 0;
    $('#table tbody tr').each(function() {
        let dayCounter = 0;
        $(this).find('td').each(function() {
            if (dayCounter++ === 1) {
                $(this).attr('class', 'text-center');
                let text = links[counter];
//                let text = $(this)[0].innerText;
                $(this)[0].innerText = '';
                $(this).append('<button id="' + text + '" class="btn btn-secondary" onclick="openARM(id)">Открыть</button>');
            }
        });
        counter++;
    });
}

function openARM(id) {
    window.open(window.origin + '/user/' + window.location.pathname.split('/')[2] + '/cross/control' + id);
}
