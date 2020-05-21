'use strict';

$(document).ready(function() {
    $('#password').on('keyup', function(event) {
        if (event.keyCode === 13) {
            event.preventDefault();
            $('#submit').trigger('click');
        }
    });
});

function check() {

    $('#loginMsg').remove();
    $('#passwordMsg').remove();

    if (($('#login').val() === '') || ($('#password').val() === '')) {
        if (!($('#loginMsg').length) && ($('#login').val() === '')){
            $('#loginForm').append('<div style="color: red;" id="loginMsg"><h5>Введите логин</h5></div>');
        }
        if (!($('#passwordMsg').length) && ($('#password').val() === '')){
            $('#passwordForm').append('<div style="color: red;" id="passwordMsg"><h5>Введите пароль</h5></div>');
        }
        return;
    }

	let account = {
		login: $('#login').val(),
		password: $('#password').val()
	};

	//Отправка на сервер запроса проверки данных
	$.ajax({
		type: 'POST',
		url: window.location.href + 'login',
		data: JSON.stringify(account),
		dataType: 'json',
		success: function (data) {
		    if (!data.status) return;
            document.cookie = ('Authorization=Bearer ' + data.token);
            //В случае успешного логина, перенаправление на участок карты данного пользователя
		    location.href = window.location.href + 'user/' + $('#login').val() + '/map';
		},
		error: function (request) {
			if (!($('#passwordMsg').length)){
                $('#passwordForm').append('<div style="color: red;" id="passwordMsg"><h5>Неверный логин и/или пароль</h5></div>');
            }
            console.log(request.status + ' ' + request.responseText);
		}
	});
}