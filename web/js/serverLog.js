'use strict';

$(document).ready(function () {

    document.title = 'Просмотр логов';
    //Закрытие вкладки при закрытии карты
//    window.setInterval(function () {
//        if(localStorage.getItem('maintab') === 'closed') window.close();
//    }, 1000);

	let $table = $('#table');

	$.ajax({
		type: 'POST',
		url: window.location.href,
		success: function (data) {
		    data.fileNames.forEach(log => {
		        $('#logs').append('<button id="' + log + '" class="btn btn-secondary ml-3 mt-1">' + log + '</button>');
		    	$('#' + log).on('click', function () {
                	getLog(log, $table);
            	});
		    });
		},
		error: function (request) {
			console.log(request.status + ' ' + request.responseText);
		}
	});

});

//Функция для получения логов
function getLog(logName, $table) {
	$.ajax({
		type: 'GET',
		url: window.location.href + '/info?fileName=' + logName,
		success: function (data) {
		    $table.bootstrapTable('removeAll');
            $table.bootstrapTable('append', data.logData);
            $table.bootstrapTable('scrollTo', 'top');
            $table.bootstrapTable('refresh', {
                data: data.logData
            });
		},
		error: function (request) {
            console.log(request.status + ' ' + request.responseText);
//			location.href = window.location.origin;
		}
	});
}